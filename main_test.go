package fsentry

import (
	"errors"
	"testing"

	"github.com/HardDie/fsentry/internal/entry_error"
)

func TestFolder(t *testing.T) {
	folderDB := NewFSEntry("test")
	err := folderDB.Init()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("create", func(t *testing.T) {
		t.Parallel()

		db := NewFSEntry("test/test_folder_create")
		err := db.Init()
		if err != nil {
			t.Fatal(err)
		}

		// Try to create folder with empty name
		err = db.CreateFolder("", nil)
		if !errors.Is(err, entry_error.ErrorBadName) {
			t.Fatal("Bad name")
		}

		// Try to create folder in not exist subdirectory
		err = db.CreateFolder("bad_path", nil, "not_exist_folder")
		if !errors.Is(err, entry_error.ErrorBadPath) {
			t.Fatal("Bad path for folder")
		}

		err = db.CreateFolder("some_folder", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Try to create duplicate
		err = db.CreateFolder("some_folder", nil)
		if !errors.Is(err, entry_error.ErrorExist) {
			t.Fatal("Folder already exist")
		}

		err = db.Drop()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("get", func(t *testing.T) {
		t.Parallel()

		db := NewFSEntry("test/test_folder_get")
		err := db.Init()
		if err != nil {
			t.Fatal(err)
		}

		// Try to get bad name
		_, err = db.GetFolder("")
		if !errors.Is(err, entry_error.ErrorBadName) {
			t.Fatal("Bad name")
		}

		// Try to get not exist folder
		_, err = db.GetFolder("not_exist_folder")
		if !errors.Is(err, entry_error.ErrorNotExist) {
			t.Fatal("Folder not exist")
		}

		// Try to get from bad path
		_, err = db.GetFolder("not_exist_folder", "bad_path")
		if !errors.Is(err, entry_error.ErrorBadPath) {
			t.Fatal("Folder not exist")
		}

		err = db.CreateFolder("some_folder", nil)
		if err != nil {
			t.Fatal(err)
		}

		entry, err := db.GetFolder("some_folder")
		if err != nil {
			t.Fatal(err)
		}

		if entry.Name != "some_folder" {
			t.Fatal("Bad name")
		}

		err = db.Drop()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("move", func(t *testing.T) {
		t.Parallel()

		db := NewFSEntry("test/test_folder_move")
		err := db.Init()
		if err != nil {
			t.Fatal(err)
		}

		// Try to move not exist folder
		err = db.MoveFolder("not_exist", "new_not_exist")
		if !errors.Is(err, entry_error.ErrorNotExist) {
			t.Fatal("Folder not exist")
		}

		// Try to move bad path
		err = db.MoveFolder("not_exist", "new_not_exist", "bad_path")
		if !errors.Is(err, entry_error.ErrorBadPath) {
			t.Fatal("Bad path")
		}

		err = db.CreateFolder("first_folder", nil)
		if err != nil {
			t.Fatal(err)
		}

		err = db.CreateFolder("second_folder", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Try to move bad name
		err = db.MoveFolder("", "new_not_exist")
		if !errors.Is(err, entry_error.ErrorBadName) {
			t.Fatal("Bad name")
		}

		// Try to move bad name
		err = db.MoveFolder("first_folder", "")
		if !errors.Is(err, entry_error.ErrorBadName) {
			t.Fatal("Bad name")
		}

		// Try to move into exist folder
		err = db.MoveFolder("first_folder", "second_folder")
		if !errors.Is(err, entry_error.ErrorExist) {
			t.Fatal("Folder already exist")
		}

		err = db.MoveFolder("first_folder", "new_first_folder")
		if err != nil {
			t.Fatal(err)
		}

		err = db.Drop()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		db := NewFSEntry("test/test_folder_update")
		err := db.Init()
		if err != nil {
			t.Fatal(err)
		}

		// Try to update bad name
		err = db.UpdateFolder("", nil)
		if !errors.Is(err, entry_error.ErrorBadName) {
			t.Fatal("Bad name")
		}

		// Try to update not exist folder
		err = db.UpdateFolder("not_exist_folder", nil)
		if !errors.Is(err, entry_error.ErrorNotExist) {
			t.Fatal("Folder not exist")
		}

		// Try to update bad path folder
		err = db.UpdateFolder("not_exist_folder", nil, "bad_path")
		if !errors.Is(err, entry_error.ErrorBadPath) {
			t.Fatal("Bad path")
		}

		err = db.Drop()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("remove", func(t *testing.T) {
		t.Parallel()

		db := NewFSEntry("test/test_folder_remove")
		err := db.Init()
		if err != nil {
			t.Fatal(err)
		}

		// Try to remove bad name
		err = db.RemoveFolder("")
		if !errors.Is(err, entry_error.ErrorBadName) {
			t.Fatal("Bad name")
		}

		// Try to remove not exist folder
		err = db.RemoveFolder("not_exist_folder")
		if !errors.Is(err, entry_error.ErrorNotExist) {
			t.Fatal("Folder not exist")
		}

		err = db.CreateFolder("some_folder", nil)
		if err != nil {
			t.Fatal(err)
		}

		err = db.RemoveFolder("some_folder")
		if err != nil {
			t.Fatal(err)
		}

		// Try to remove twice
		err = db.RemoveFolder("some_folder")
		if !errors.Is(err, entry_error.ErrorNotExist) {
			t.Fatal("Folder not exist")
		}

		err = db.Drop()
		if err != nil {
			t.Fatal(err)
		}
	})
}
