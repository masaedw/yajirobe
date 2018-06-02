package yajirobe

import (
	"encoding/json"

	"go.uber.org/zap"

	"github.com/masaedw/yajirobe/lib/storedmap"

	"github.com/pkg/errors"
)

// FundInfo ファンド情報
type FundInfo struct {
	Code  FundCode   `json:"code"`
	Class AssetClass `json:"class"`
	Name  string     `json:"name"`
}

// Cache ファンド情報のキャッシュ
type Cache interface {
	GetFund(code FundCode) (*FundInfo, error)
	CanGetFund(code FundCode) bool
	SetFund(info *FundInfo) error

	GetString(key string) (string, error)
	CanGetString(key string) bool
	SetString(key, data string) error
}

type cache struct {
	smap storedmap.StoredMap
}

func fundKey(code FundCode) string {
	return "fund." + string(code)
}

func (c *cache) GetFund(code FundCode) (*FundInfo, error) {
	data, err := c.smap.Get(fundKey(code))
	if err != nil {
		return nil, errors.Wrap(err, "cache doesn't exist")
	}

	info := &FundInfo{}

	if err = json.Unmarshal(data, info); err != nil {
		return nil, errors.Wrap(err, "can't unmarshal cache")
	}

	return info, nil
}

func (c *cache) CanGetFund(code FundCode) bool {
	return c.smap.CanGet(fundKey(code))
}

func (c *cache) SetFund(info *FundInfo) error {
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

	return errors.Wrap(
		c.smap.Set(fundKey(info.Code), data),
		"can't set to storedmap")
}

func stringKey(key string) string {
	return "string." + key
}

func (c *cache) GetString(key string) (string, error) {
	data, err := c.smap.Get(stringKey(key))
	if err != nil {
		return "", errors.Wrap(err, "can't get data from storedmap")
	}

	return string(data), nil
}

func (c *cache) CanGetString(key string) bool {
	return c.smap.CanGet(stringKey(key))
}

func (c *cache) SetString(key, data string) error {
	return errors.Wrap(
		c.smap.Set(stringKey(key), []byte(data)),
		"can't set to storedmap")
}

// NewMemoryCache creates a Cache
func NewMemoryCache() Cache {
	return &cache{
		smap: storedmap.NewMemoryMap(),
	}
}

// NewFileCache creates a Cache.
// If logger is nil, NewFileCache uses NewNop as logger.
func NewFileCache(logger *zap.Logger) (Cache, error) {
	smap, err := storedmap.NewFileMap(logger)
	if err != nil {
		return nil, errors.Wrap(err, "can't create storedmap")
	}
	return &cache{
		smap: smap,
	}, nil
}
