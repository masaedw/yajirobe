package cache

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"go.uber.org/zap"
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
	l, _ := zap.NewDevelopment()
	c, _ := NewFileCache(l)
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

func TestFileGetOrFind_NewKey(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)

	json, err := c.GetOrFind("newkey", func(key string) (string, error) {
		return "newkey_json", nil
	})

	if json != "newkey_json" {
		t.Errorf("expected newkey_json but got %s", json)
	}

	if err != nil {
		t.Errorf("error occured %v", err)
	}
}

func TestFileGetOrFind_ExistsKey(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)
	c.Set("existskey", "existskey_json")

	json, err := c.GetOrFind("existskey", func(key string) (string, error) {
		t.Errorf("finder unexpectedly called")
		return "", nil
	})

	if json != "existskey_json" {
		t.Errorf("expected existskey_json but got %s", json)
	}

	if err != nil {
		t.Errorf("error occured %v", err)
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
