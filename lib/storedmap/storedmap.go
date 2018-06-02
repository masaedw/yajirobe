package storedmap

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Finder キャッシュがない場合に実際にデータを取りに行く関数
type Finder func(key string) (string, error)

// StoredMap JSONを保存するインターフェイス
type StoredMap interface {
	// keyは `[a-zA-Z0-9.]+`` の形式
	Get(key string) (string, error)
	GetOrFind(key string, finder Finder) (string, error)
	Set(key, json string) error
}

type memoryMap struct {
	data map[string]string
}

func (c *memoryMap) Get(key string) (string, error) {
	if d, ok := c.data[key]; ok {
		return d, nil
	}
	return "", &NotExistsError{Key: key}
}

func (c *memoryMap) GetOrFind(key string, finder Finder) (string, error) {
	if d, ok := c.data[key]; ok {
		return d, nil
	}

	d, err := finder(key)

	if err != nil {
		return "", errors.Wrap(err, "finder failed")
	}

	c.data[key] = d
	return d, nil
}

func (c *memoryMap) Set(key, json string) error {
	c.data[key] = json
	return nil
}

// NewMemoryMap memoryCacheを作る
func NewMemoryMap() StoredMap {
	return &memoryMap{
		data: map[string]string{},
	}
}

type fileMap struct {
	path   string
	logger *zap.Logger
}

var keyPattern = regexp.MustCompile(`[^a-zA-Z0-9.]+`)

func escapeKey(key string) string {
	return keyPattern.ReplaceAllString(key, "")
}

func (c *fileMap) prepareDir() error {
	return os.MkdirAll(c.fundPath(), 0755)
}

func (c *fileMap) fundPath() string {
	return filepath.Join(c.path, "cache")
}

func (c *fileMap) fundFilePath(key string) string {
	fname := escapeKey(key)
	return filepath.Join(c.fundPath(), fname)
}

func (c *fileMap) Get(key string) (string, error) {
	if c == nil {
		return "", errors.New("Get method called with nil")
	}

	fullPath := c.fundFilePath(key)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.logger.Sugar().Debugf("cache miss: %v", key)
		return "", &NotExistsError{Key: key}
	}

	c.logger.Sugar().Debugf("cache hit: %v", key)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", errors.Wrap(err, "can't read cache file")
	}

	return string(data), nil
}

func (c *fileMap) Set(key, json string) error {
	key = escapeKey(key)
	if key == "" {
		return errors.New("key is empty")
	}

	if err := c.prepareDir(); err != nil {
		return errors.Wrap(err, "can't prepare directory")
	}

	filePath := c.fundFilePath(key)

	return ioutil.WriteFile(filePath, []byte(json), 0644)
}

func (c *fileMap) GetOrFind(key string, finder Finder) (string, error) {
	d, err := c.Get(key)
	switch {
	case err != nil && !IsNotExists(err):
		return "", errors.WithStack(err)

	case err != nil && IsNotExists(err):
		d, err := finder(key)
		if d != "" {
			err = c.Set(key, d)
		}
		return d, errors.WithStack(err)

	default:
		return d, nil
	}
}

// NewFileMap FileMapを作る
func NewFileMap(logger *zap.Logger) (StoredMap, error) {
	dirPath, err := cachePath()
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	fc := &fileMap{
		path:   dirPath,
		logger: logger,
	}

	return fc, nil
}

// NotExistsError 保存された情報がない
type NotExistsError struct {
	Key string
}

// IsNotExists eがNotExistsErrorならtrueを返す
func IsNotExists(e error) bool {
	switch e.(type) {
	case *NotExistsError:
		return true
	default:
		return false
	}
}

func (e *NotExistsError) Error() string {
	return fmt.Sprintf("no cached info of the key: %v", e.Key)
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
