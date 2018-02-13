package yajirobe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// FundInfo ファンド情報
type FundInfo struct {
	Code  FundCode   `json:"code"`
	Class AssetClass `json:"class"`
	Name  string     `json:"name"`
}

// FundInfoFinder ファンドコードからファンド情報を取得する関数
type FundInfoFinder func(code FundCode) (*FundInfo, error)

// FundInfoCache ファンド情報のキャッシュ
type FundInfoCache interface {
	Get(code FundCode) (*FundInfo, error)
	GetOrFind(code FundCode, finder FundInfoFinder) (*FundInfo, error)
	Set(info *FundInfo) error
}

// DefaultGetOrFind GetとSetをつかったGetOrFindのデフォルト実装
func DefaultGetOrFind(fc FundInfoCache, code FundCode, finder FundInfoFinder) (*FundInfo, error) {
	fi, err := fc.Get(code)
	switch {
	case err != nil && !IsCacheNotExists(err):
		return nil, errors.WithStack(err)

	case err != nil && IsCacheNotExists(err):
		fi, err := finder(code)
		if fi != nil {
			err = fc.Set(fi)
		}
		return fi, errors.WithStack(err)

	default:
		return fi, nil
	}
}

// CacheNotExistsError キャッシュされた情報がない
type CacheNotExistsError struct {
	Code FundCode
}

// IsCacheNotExists eがCacheNotExistsErrorならtrueを返す
func IsCacheNotExists(e error) bool {
	switch e.(type) {
	case *CacheNotExistsError:
		return true
	default:
		return false
	}
}

func (e *CacheNotExistsError) Error() string {
	return fmt.Sprintf("no cached info of the code: %v", e.Code)
}

// FileFundInfoCache ローカルファイルにキャッシュするFundInfoCache
type FileFundInfoCache struct {
	path   string
	logger *zap.Logger
}

func cachePath() (string, error) {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return "", errors.New("APPDATA is not defined")
		}
		return filepath.Join(appdata, "yajirobe"), nil
	}

	return filepath.Join(os.Getenv("HOME"), ".yajirobe"), nil
}

// NewFileFundInfoCache NewFileFundInfoCacheを作る
func NewFileFundInfoCache(logger *zap.Logger) (FundInfoCache, error) {
	dirPath, err := cachePath()
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	fc := &FileFundInfoCache{
		path:   dirPath,
		logger: logger,
	}

	return fc, nil
}

func (fc *FileFundInfoCache) prepareDir() error {
	return os.MkdirAll(fc.fundPath(), 0755)
}

func (fc *FileFundInfoCache) fundPath() string {
	return filepath.Join(fc.path, "cache", "funds")
}

func (fc *FileFundInfoCache) fundFilePath(code FundCode) string {
	fname := filepath.Clean(string(code))
	return filepath.Join(fc.fundPath(), fname)
}

// Get 取得
func (fc *FileFundInfoCache) Get(code FundCode) (*FundInfo, error) {
	if fc == nil {
		return nil, errors.New("Get method called with nil")
	}

	fullPath := fc.fundFilePath(code)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fc.logger.Sugar().Debugf("cache miss: %v", code)
		return nil, &CacheNotExistsError{Code: code}
	}

	fc.logger.Sugar().Debugf("cache hit: %v", code)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, errors.Wrap(err, "can't read cache file")
	}

	info := &FundInfo{}

	if err = json.Unmarshal(data, info); err != nil {
		return nil, errors.Wrap(err, "can't unmarshal cache")
	}

	return info, nil
}

// Set 設定
func (fc *FileFundInfoCache) Set(info *FundInfo) error {
	if info == nil {
		return errors.New("info is nil")
	}

	if info.Code == FundCode("") {
		return errors.New("code must not be empty")
	}

	data, err := json.Marshal(info)
	if err != nil {
		return errors.Wrap(err, "can't marshal fundinfo")
	}

	if err = fc.prepareDir(); err != nil {
		return errors.Wrap(err, "can't prepare directory")
	}

	filePath := fc.fundFilePath(info.Code)

	return ioutil.WriteFile(filePath, data, 0644)
}

// GetOrFind キャッシュにあれば取得、なければFundInfoFinderを使って探しに行く
func (fc *FileFundInfoCache) GetOrFind(code FundCode, finder FundInfoFinder) (*FundInfo, error) {
	return DefaultGetOrFind(fc, code, finder)
}
