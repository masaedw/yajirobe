package yajirobe

import (
	"testing"

	"github.com/masaedw/yajirobe/lib/storedmap"
)

func TestFileGet(t *testing.T) {
	fc := NewMemoryCache()

	c, err := fc.GetFund(FundCode("12345"))
	if c != nil || err == nil {
		t.Fatalf("expected empty but got something: %v", c)
	}

	if !storedmap.IsNotExists(err) {
		t.Fatalf("expected NotExistsError but got %v", err)
	}
}

func TestFileSet(t *testing.T) {
	fc := NewMemoryCache()

	info := &FundInfo{
		Code:  FundCode("12345"),
		Name:  "TestFund",
		Class: InternationalStocks,
	}

	if err := fc.SetFund(info); err != nil {
		t.Fatal(err)
	}
}

func TestFileSetGet(t *testing.T) {
	fc := NewMemoryCache()

	info := &FundInfo{
		Code:  FundCode("12345"),
		Name:  "TestFund",
		Class: InternationalStocks,
	}

	if err := fc.SetFund(info); err != nil {
		t.Fatal(err)
	}

	cache, err := fc.GetFund(FundCode("12345"))
	if err != nil {
		t.Fatal(err)
	}

	if !(cache.Class == info.Class &&
		cache.Code == info.Code &&
		cache.Name == info.Name) {
		t.Fatalf("expected %v but got %v", info, cache)
	}
}
