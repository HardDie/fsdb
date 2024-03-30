package service

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/HardDie/fsentry/internal/entry"
	fsStorage "github.com/HardDie/fsentry/internal/fs/storage"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

func TestEntryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_entry_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		s := New(fsStorage.NewFS(), true)
		_, err = s.Create(dir, "success", nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}
func TestEntryGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "get_entry_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"

		s := New(fsStorage.NewFS(), true)
		ent, err := s.Create(dir, name, nil)
		if err != nil {
			t.Fatal(err)
		}

		entResp, err := s.Get(dir, name)
		if err != nil {
			t.Fatal(err)
		}

		if !compareEntry(t, entResp, ent) {
			t.Fatal("entry must be equal")
		}
	})
}
func TestEntryMove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "move_entry_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldName := "success"
		newName := "success_moved"

		s := New(fsStorage.NewFS(), true)
		info, err := s.Create(dir, oldName, nil)
		if err != nil {
			t.Fatal(err)
		}

		newInfo, err := s.Move(dir, oldName, newName)
		if err != nil {
			t.Fatal(err)
		}
		if compareEntry(t, info, newInfo) {
			t.Fatal("after move entry must be different")
		}

		infoResp, err := s.Get(dir, newName)
		if err != nil {
			t.Fatal(err)
		}
		if !compareEntry(t, infoResp, newInfo) {
			t.Fatal("entry must be equal")
		}

		_, err = s.Get(dir, oldName)
		if err == nil {
			t.Fatal("entry was moved, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})
}
func TestEntryUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_entry_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"

		s := New(fsStorage.NewFS(), true)
		ent, err := s.Create(dir, name, []byte("hello world"))
		if err != nil {
			t.Fatal(err)
		}

		updatedEnt, err := s.Update(dir, name, []byte("updated hello world"))
		if err != nil {
			t.Fatal(err)
		}
		if compareEntry(t, ent, updatedEnt) {
			t.Fatal("after update entry must be different")
		}

		updatedEnt, err = s.Get(dir, name)
		if err != nil {
			t.Fatal(err)
		}
		if compareEntry(t, ent, updatedEnt) {
			t.Fatal("after update entry must be different")
		}
	})
}
func TestEntryRemove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_entry_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		name := "success"

		s := New(fsStorage.NewFS(), true)
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
func TestEntryDuplicate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "duplicate_entry_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldName := "success"
		newName := "success_duplicate"

		s := New(fsStorage.NewFS(), true)
		ent, err := s.Create(dir, oldName, []byte("some data"))
		if err != nil {
			t.Fatal(err)
		}

		duplicateEnt, err := s.Duplicate(dir, oldName, newName)
		if err != nil {
			t.Fatal(err)
		}
		if compareEntry(t, ent, duplicateEnt) {
			t.Fatal("after duplicate entry must be different")
		}
	})
}

func compareEntry(t *testing.T, got, want *entry.Entry) bool {
	if want == nil && got == nil {
		return true
	}
	if want != nil && got == nil {
		t.Log("entry; got: nil, want not nil")
		return false
	}
	if want == nil && got != nil {
		t.Log("entry; got: not nil, want nil")
		return false
	}

	isEqual := true
	if want.ID != got.ID {
		t.Logf("entry.ID; %q != %q", got.ID, want.ID)
		isEqual = false
	}
	if want.Name != got.Name {
		t.Logf("entry.Name; %q != %q", got.Name, want.Name)
		isEqual = false
	}
	if !want.CreatedAt.Equal(got.CreatedAt) {
		t.Logf("entry.CreatedAt; %v != %v", got.CreatedAt, want.CreatedAt)
		isEqual = false
	}
	if !want.UpdatedAt.Equal(got.UpdatedAt) {
		t.Logf("entry.UpdatedAt; %v != %v", got.UpdatedAt, want.UpdatedAt)
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
		t.Logf("entry.Data; %q != %q", gotData, wantData)
		isEqual = false
	}
	return isEqual
}
