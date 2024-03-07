package utils

import (
	"os"
	"path"
	"testing"
)

func FuzzNameToID(f *testing.F) {
	// Check before starting fuzzing
	tmp, err := os.MkdirTemp("", "fsentry")
	if err != nil {
		f.Fatal("error creating temp dir", err)
	}
	err = os.Remove(tmp)
	if err != nil {
		f.Fatal("error removing temp dir", err)
	}

	f.Fuzz(func(t *testing.T, name string) {
		resName := NameToID(name)
		if resName == "" {
			t.Logf("Empty name: %q", name)
			return
		}

		tmpDirName, err := os.MkdirTemp("", "fsentry")
		if err != nil {
			f.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(tmpDirName)

		filePath := path.Join(tmpDirName, resName)

		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("error create file %q: %s", resName, err.Error())
		}
		err = f.Close()
		if err != nil {
			t.Fatalf("error close file %q: %s", resName, err.Error())
		}
		err = os.Remove(filePath)
		if err != nil {
			t.Fatalf("error removing file %q: %s", resName, err.Error())
		}
	})
}
