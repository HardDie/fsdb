package service

import (
	"errors"
	"os"
	"strings"
	"testing"

	fsStorage "github.com/HardDie/fsentry/internal/fs/storage"
	"github.com/HardDie/fsentry/pkg/fsentry"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

func TestFolderCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		s := New(fsStorage.New(), true)
		_, err = s.Create(dir, "success", nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}
func TestFolderGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "get_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"

		s := New(fsStorage.New(), true)
		info, err := s.Create(dir, name, nil)
		if err != nil {
			t.Fatal(err)
		}

		infoResp, err := s.Get(dir, name)
		if err != nil {
			t.Fatal(err)
		}
		if !compareInfo(t, infoResp, info) {
			t.Fatal("info must be equal")
		}
	})
}
func TestFolderMove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "move_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldName := "success"
		newName := "success_moved"

		s := New(fsStorage.New(), true)
		info, err := s.Create(dir, oldName, nil)
		if err != nil {
			t.Fatal(err)
		}

		newInfo, err := s.Move(dir, oldName, newName)
		if err != nil {
			t.Fatal(err)
		}
		if compareInfo(t, info, newInfo) {
			t.Fatal("after move info must be different")
		}

		infoResp, err := s.Get(dir, newName)
		if err != nil {
			t.Fatal(err)
		}
		if !compareInfo(t, infoResp, newInfo) {
			t.Fatal("info must be equal")
		}

		_, err = s.Get(dir, oldName)
		if err == nil {
			t.Fatal("folder was moved, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})
}
func TestFolderUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"

		s := New(fsStorage.New(), true)
		info, err := s.Create(dir, name, []byte("hello world"))
		if err != nil {
			t.Fatal(err)
		}

		updatedInfo, err := s.Update(dir, name, []byte("updated hello world"))
		if err != nil {
			t.Fatal(err)
		}
		if compareInfo(t, info, updatedInfo) {
			t.Fatal("after update info must be different")
		}

		updatedInfo, err = s.Get(dir, name)
		if err != nil {
			t.Fatal(err)
		}
		if compareInfo(t, info, updatedInfo) {
			t.Fatal("after update info must be different")
		}
	})
}
func TestFolderRemove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"

		s := New(fsStorage.New(), true)
		_, err = s.Create(dir, name, nil)
		if err != nil {
			t.Fatal(err)
		}

		err = s.Remove(dir, name)
		if err != nil {
			t.Fatal(err)
		}
	})
}
func TestFolderDuplicate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "duplicate_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldName := "success"
		newName := "success_duplicate"

		s := New(fsStorage.New(), true)
		ent, err := s.Create(dir, oldName, []byte("some data"))
		if err != nil {
			t.Fatal(err)
		}

		duplicateEnt, err := s.Duplicate(dir, oldName, newName)
		if err != nil {
			t.Fatal(err)
		}
		if compareInfo(t, ent, duplicateEnt) {
			t.Fatal("after duplicate info must be different")
		}
	})
}
func TestFolderMoveWithoutTimestamp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "move_without_timestamp_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldName := "success"
		newName := "success_moved"

		s := New(fsStorage.New(), true)
		info, err := s.Create(dir, oldName, nil)
		if err != nil {
			t.Fatal(err)
		}

		newInfo, err := s.MoveWithoutTimestamp(dir, oldName, newName)
		if err != nil {
			t.Fatal(err)
		}
		if compareInfo(t, info, newInfo) {
			t.Fatal("after move info must be different")
		}

		infoResp, err := s.Get(dir, newName)
		if err != nil {
			t.Fatal(err)
		}
		if !compareInfo(t, infoResp, newInfo) {
			t.Fatal("info must be equal")
		}

		_, err = s.Get(dir, oldName)
		if err == nil {
			t.Fatal("folder was moved, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})
}

func compareInfo(t *testing.T, got, want *fsentry.FolderInfo) bool {
	if want == nil && got == nil {
		return true
	}
	if want != nil && got == nil {
		t.Log("info; got: nil, want not nil")
		return false
	}
	if want == nil && got != nil {
		t.Log("info; got: not nil, want nil")
		return false
	}

	isEqual := true
	if want.ID != got.ID {
		t.Logf("info.ID; %q != %q", got.ID, want.ID)
		isEqual = false
	}
	if want.Name != got.Name {
		t.Logf("info.Name; %q != %q", got.Name, want.Name)
		isEqual = false
	}
	if !want.CreatedAt.Equal(got.CreatedAt) {
		t.Logf("info.CreatedAt; %v != %v", got.CreatedAt, want.CreatedAt)
		isEqual = false
	}
	if !want.UpdatedAt.Equal(got.UpdatedAt) {
		t.Logf("info.UpdatedAt; %v != %v", got.UpdatedAt, want.UpdatedAt)
		isEqual = false
	}

	// TODO: figure out why data before and after is different
	cleanString := func(in string) string {
		val := strings.ReplaceAll(in, "\r", "")
		val = strings.ReplaceAll(val, "\n", "")
		return val
	}
	wantData := cleanString(string(want.Data))
	gotData := cleanString(string(got.Data))
	if wantData != gotData {
		t.Logf("info.Data; %q != %q", gotData, wantData)
		isEqual = false
	}
	return isEqual
}
