package storedmap

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"

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

func tempDirInfoCache(tempDir string) StoredMap {
	l, _ := zap.NewDevelopment()
	c, _ := NewFileMap(l)
	c.(*fileMap).path = tempDir
	return c
}

func TestFileGet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)

	d, err := c.Get("12345")
	if d != nil || err == nil {
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

	if err := c.Set("12345", []byte(json)); err != nil {
		t.Fatal(err)
	}
}

func TestFileSetGet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)

	json := "{code: 12345}"

	if err := c.Set("12345", []byte(json)); err != nil {
		t.Fatal(err)
	}

	cache, err := c.Get("12345")
	if err != nil {
		t.Fatal(err)
	}

	if string(cache) != json {
		t.Fatalf("expected %v but got %v", json, cache)
	}
}

func TestFileCanGet(t *testing.T) {
	tempDir := tempDir(t)
	defer os.RemoveAll(tempDir)

	c := tempDirInfoCache(tempDir)
	c.Set("existskey", []byte("existsdata"))

	if c.CanGet("notexistskey") == true {
		t.Errorf(`expected c.CanGet("notexistskey") is false but got true`)
	}

	if c.CanGet("existskey") == false {
		t.Errorf(`expected c.CanGet("existskey") is true but got false`)
	}
}

func TestIsNotExists(t *testing.T) {
	if !IsNotExists(&NotExistsError{}) {
		t.Fatal("expected true but got false")
	}

	if !IsNotExists(errors.WithStack(&NotExistsError{})) {
		t.Fatal("expected true but got false")
	}

	if IsNotExists(errors.New("some error")) {
		t.Fatal("expected false but got true")
	}
}
