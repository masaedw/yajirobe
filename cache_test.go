package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func tempDir(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "yajirobe")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("tempDir: %s", tempDir)
	return tempDir
}

func TestFileGet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	fc := FundInfoCache(&FileFundInfoCache{path: tempDir})

	c, err := fc.Get(FundCode("12345"))
	if c != nil || err == nil {
		t.Fatalf("expected empty but got something: %v", c)
	}
	switch err.(type) {
	default:
		t.Fatal("expected CacheNotExistsError")

	case *CacheNotExistsError:
		// do nothing
	}
}

func TestFileSet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	fc := FundInfoCache(&FileFundInfoCache{path: tempDir})

	info := &FundInfo{
		Code:  FundCode("12345"),
		Name:  "TestFund",
		Class: InternationalStocks,
	}

	if err := fc.Set(info); err != nil {
		t.Fatal(err)
	}
}

func TestFileSetGet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	fc := FundInfoCache(&FileFundInfoCache{path: tempDir})

	info := &FundInfo{
		Code:  FundCode("12345"),
		Name:  "TestFund",
		Class: InternationalStocks,
	}

	if err := fc.Set(info); err != nil {
		t.Fatal(err)
	}

	cache, err := fc.Get(FundCode("12345"))
	if err != nil {
		t.Fatal(err)
	}

	if !(cache.Class == info.Class &&
		cache.Code == info.Code &&
		cache.Name == info.Name) {
		t.Fatalf("expected %v but got %v", info, cache)
	}
}
