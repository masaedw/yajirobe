package cache

import (
	"errors"
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

func tempDirInfoCache(tempDir string) Cache {
	c, _ := NewFileCache(nil)
	c.(*fileCache).path = tempDir
	return c
}

func TestFileGet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)

	d, err := c.Get("12345")
	if d != "" || err == nil {
		t.Fatalf("expected empty but got something: %v", d)
	}
	switch err.(type) {
	default:
		t.Fatal("expected NotExistsError")

	case *NotExistsError:
		// do nothing
	}
}

func TestFileSet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)

	json := "{}"

	if err := c.Set("12345", json); err != nil {
		t.Fatal(err)
	}
}

func TestFileSetGet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)

	json := "{code: 12345}"

	if err := c.Set("12345", json); err != nil {
		t.Fatal(err)
	}

	cache, err := c.Get("12345")
	if err != nil {
		t.Fatal(err)
	}

	if cache != json {
		t.Fatalf("expected %v but got %v", json, cache)
	}
}

func TestIsNotExists(t *testing.T) {
	if (!IsNotExists(&NotExistsError{})) {
		t.Fatal("expected true but got false")
	}

	if IsNotExists(errors.New("some error")) {
		t.Fatal("expected false but got true")
	}
}
