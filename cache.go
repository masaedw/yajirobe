package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

// FundInfo ファンド情報
type FundInfo struct {
	Code  FundCode   `json:"code"`
	Class AssetClass `json:"class"`
	Name  string     `json:"name"`
}

// FundInfoCache ファンド情報のキャッシュ
type FundInfoCache interface {
	Get(code FundCode) (*FundInfo, error)
	Set(info *FundInfo) error
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
	path string
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
func NewFileFundInfoCache() (FundInfoCache, error) {
	dirPath, err := cachePath()
	if err != nil {
		return nil, err
	}

	fc := &FileFundInfoCache{
		path: dirPath,
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
		return nil, &CacheNotExistsError{Code: code}
	}

	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	info := &FundInfo{}

	if err = json.Unmarshal(data, info); err != nil {
		return nil, err
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
		return err
	}

	if err = fc.prepareDir(); err != nil {
		return nil
	}

	filePath := fc.fundFilePath(info.Code)

	return ioutil.WriteFile(filePath, data, 0644)
}
