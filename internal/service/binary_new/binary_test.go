package binary_new

import (
	"errors"
	"os"
	"reflect"
	"testing"

	fsStorage "github.com/HardDie/fsentry/internal/fs/storage"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

func TestBinaryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_binary_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		s := NewBinary(fsStorage.NewFS(), true)
		err = s.Create(dir, "success", nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}
func TestBinaryGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "get_binary_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"
		payload := []byte("check")

		s := NewBinary(fsStorage.NewFS(), true)
		err = s.Create(dir, name, payload)
		if err != nil {
			t.Fatal(err)
		}

		binResp, err := s.Get(dir, name)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(payload, binResp) {
			t.Fatal("binary must be equal")
		}
	})
}
func TestBinaryMove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "move_binary_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldName := "success"
		newName := "success_moved"

		payload := []byte("check")

		s := NewBinary(fsStorage.NewFS(), true)
		err = s.Create(dir, oldName, payload)
		if err != nil {
			t.Fatal(err)
		}

		err = s.Move(dir, oldName, newName)
		if err != nil {
			t.Fatal(err)
		}

		binResp, err := s.Get(dir, newName)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(payload, binResp) {
			t.Fatal("binary must be equal")
		}

		_, err = s.Get(dir, oldName)
		if err == nil {
			t.Fatal("binary was moved, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})
}
func TestBinaryUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_binary_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"
		payload := []byte("check")

		s := NewBinary(fsStorage.NewFS(), true)
		err = s.Create(dir, name, payload)
		if err != nil {
			t.Fatal(err)
		}

		err = s.Update(dir, name, []byte("updated hello world"))
		if err != nil {
			t.Fatal(err)
		}

		updatedBin, err := s.Get(dir, name)
		if err != nil {
			t.Fatal(err)
		}
		if reflect.DeepEqual(payload, updatedBin) {
			t.Fatal("after update binary must be different")
		}
	})
}
func TestBinaryRemove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_binary_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"
		payload := []byte("check")

		s := NewBinary(fsStorage.NewFS(), true)
		err = s.Create(dir, name, payload)
		if err != nil {
			t.Fatal(err)
		}

		err = s.Remove(dir, name)
		if err != nil {
			t.Fatal(err)
		}
	})
}
func TestBinaryDuplicate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "duplicate_binary_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldName := "success"
		newName := "success_duplicate"
		payload := []byte("check")

		s := NewBinary(fsStorage.NewFS(), true)
		err = s.Create(dir, oldName, payload)
		if err != nil {
			t.Fatal(err)
		}

		duplicateBin, err := s.Duplicate(dir, oldName, newName)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(payload, duplicateBin) {
			t.Fatal("after duplicate binary must be equal")
		}
	})
}
