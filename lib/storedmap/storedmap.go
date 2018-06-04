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

// StoredMap バイト列を保存するインターフェイス
type StoredMap interface {
	// keyは `[a-zA-Z0-9.]+`` の形式
	Get(key string) ([]byte, error)
	CanGet(key string) bool
	Set(key string, data []byte) error
}

type memoryMap struct {
	data map[string][]byte
}

func (c *memoryMap) Get(key string) ([]byte, error) {
	if d, ok := c.data[key]; ok {
		return d, nil
	}
	return nil, &NotExistsError{Key: key}
}

func (c *memoryMap) CanGet(key string) bool {
	_, ok := c.data[key]
	return ok
}

func (c *memoryMap) Set(key string, data []byte) error {
	c.data[key] = data
	return nil
}

// NewMemoryMap memoryCacheを作る
func NewMemoryMap() StoredMap {
	return &memoryMap{
		data: map[string][]byte{},
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

func (c *fileMap) Get(key string) ([]byte, error) {
	if c == nil {
		return nil, errors.New("Get method called with nil")
	}

	fullPath := c.fundFilePath(key)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.logger.Sugar().Debugf("cache miss: %v", key)
		return nil, &NotExistsError{Key: key}
	}

	c.logger.Sugar().Debugf("cache hit: %v", key)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, errors.Wrap(err, "can't read cache file")
	}

	return data, nil
}

func (c *fileMap) CanGet(key string) bool {
	_, err := c.Get(key)
	return err == nil
}

func (c *fileMap) Set(key string, data []byte) error {
	key = escapeKey(key)
	if key == "" {
		return errors.New("key is empty")
	}

	if err := c.prepareDir(); err != nil {
		return errors.Wrap(err, "can't prepare directory")
	}

	filePath := c.fundFilePath(key)

	return ioutil.WriteFile(filePath, data, 0644)
}

// NewFileMap creates a FileMap
// If logger is nil, NewFileMap uses NewNop as logger.
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
	switch errors.Cause(e).(type) {
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
