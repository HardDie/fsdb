package storage

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	acl "github.com/hectane/go-acl"

	"github.com/HardDie/fsentry/internal/fs"
)

func TestCreateFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := New()
		err = f.CreateFile(filepath.Join(dir, "success"), nil)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("exist_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_exist_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "exist")

		f := New()
		err = f.CreateFile(filePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFile(filePath, []byte("hello"))
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorFileExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFileExist, err)
		}
	})

	t.Run("exist_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_exist_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "exist")

		f := New()
		err = f.CreateFolder(filePath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFile(filePath, []byte("hello"))
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorFileExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFileExist, err)
		}
	})

	t.Run("bad_path", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_bad_path")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "bad_path", "file")

		f := New()

		err = f.CreateFile(filePath, []byte("hello"))
		if err == nil {
			t.Fatal("bad path, must be error")
		}
		if !errors.Is(err, fs.ErrorBadPath) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorBadPath, err)
		}
	})

	t.Run("permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_permissions")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		// Forbid creating something inside
		err = chmod(dir, 0400)
		if err != nil {
			t.Fatal("error updating permission", err)
		}

		f := New()
		err = f.CreateFile(filepath.Join(dir, "permissions"), []byte("hello"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})
}
func TestReadFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "success")
		data := []byte("hello")

		f := New()
		err = f.CreateFile(filePath, data)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := f.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(data, resp) {
			t.Fatalf("bad data readed; got: %q, want: %q", string(resp), string(data))
		}
	})

	t.Run("invalid", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_invalid")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "invalid")

		f := New()
		err = f.CreateFolder(filePath)
		if err != nil {
			t.Fatal(err)
		}

		_, err = f.ReadFile(filePath)
		if err == nil {
			t.Fatal("not file, must be error")
		}
		if !errors.Is(err, fs.ErrorFileNotExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFileNotExist, err)
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := New()
		_, err = f.ReadFile(filepath.Join(dir, "not_exist"))
		if err == nil {
			t.Fatal("file not exist, must be error")
		}
		if !errors.Is(err, fs.ErrorFileNotExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFileNotExist, err)
		}
	})

	t.Run("permissions_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_permissions_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFile(filePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		// Forbid reading
		err = chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = chmod(dir, CreateDirPerm)
			if err != nil {
				t.Fatal("error updating permission", err)
			}
		}()

		_, err = f.ReadFile(filePath)
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})

	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFile(filePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		// Forbid reading
		err = chmod(filePath, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}

		_, err = f.ReadFile(filePath)
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})
}
func TestUpdateFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "success")
		data := []byte("hello")

		f := New()
		err = f.CreateFile(filePath, []byte("init"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.UpdateFile(filePath, data)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := f.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(data, resp) {
			t.Fatalf("bad data readed; got: %q, want: %q", string(resp), string(data))
		}
	})

	t.Run("invalid", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_invalid")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "invalid")

		f := New()
		err = f.CreateFolder(filePath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.UpdateFile(filePath, []byte("new"))
		if err == nil {
			t.Fatal("not file, must be error")
		}
		if !errors.Is(err, fs.ErrorFileNotExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFileNotExist, err)
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := New()
		err = f.UpdateFile(filepath.Join(dir, "not_exist"), []byte("hello"))
		if err == nil {
			t.Fatal("file not exist, must be error")
		}
		if !errors.Is(err, fs.ErrorFileNotExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFileNotExist, err)
		}
	})

	t.Run("permissions_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_permissions_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFile(filePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		// Forbid writing
		err = chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = chmod(dir, CreateDirPerm)
			if err != nil {
				t.Fatal("error updating permission", err)
			}
		}()

		err = f.UpdateFile(filePath, []byte("new"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})

	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFile(filePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		// Forbid writing
		err = chmod(filePath, 0400)
		if err != nil {
			t.Fatal("error updating permission", err)
		}

		err = f.UpdateFile(filePath, []byte("new"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})
}
func TestRemoveFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "success")
		data := []byte("hello")

		f := New()
		err = f.CreateFile(filePath, data)
		if err != nil {
			t.Fatal(err)
		}

		err = f.RemoveFile(filePath)
		if err != nil {
			t.Fatal(err)
		}

		_, err = f.ReadFile(filePath)
		if err == nil {
			t.Fatal("file was removed, must be error")
		}
		if !errors.Is(err, fs.ErrorFileNotExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFileNotExist, err)
		}
	})

	// You can remove empty folder
	t.Run("empty_folder", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_empty_folder")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "empty_folder")

		f := New()
		err = f.CreateFolder(filePath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.RemoveFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not_empty_folder", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_not_empty_folder")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "not_empty_folder")

		f := New()
		err = f.CreateAllFolder(filepath.Join(filePath, "data"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.RemoveFile(filePath)
		if err == nil {
			t.Fatal("not file, must be error")
		}
		if !errors.Is(err, fs.ErrorFolderNotEmpty) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFolderNotEmpty, err)
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := New()
		err = f.RemoveFile(filepath.Join(dir, "not_exist"))
		if err == nil {
			t.Fatal("file not exist, must be error")
		}
		if !errors.Is(err, fs.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorNotExist, err)
		}
	})

	t.Run("permissions_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_permissions_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFile(filePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		// Forbid removing
		err = chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = chmod(dir, CreateDirPerm)
			if err != nil {
				t.Fatal("error updating permission", err)
			}
		}()

		err = f.RemoveFile(filePath)
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})

	// Even if a user doesn't have write permissions, but it's their file, they have the right to delete it.
	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFile(filePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		// Forbid writing
		err = chmod(filePath, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}

		err = f.RemoveFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
	})
}
func TestCreateFolder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := New()
		err = f.CreateFolder(filepath.Join(dir, "success"))
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("exist_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_exist_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "exist")

		f := New()
		err = f.CreateFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFolder(folderPath)
		if err == nil {
			t.Fatal("folder already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorFolderExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFolderExist, err)
		}
	})

	t.Run("exist_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_exist_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "exist")

		f := New()
		err = f.CreateFile(folderPath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFolder(folderPath)
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorFolderExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFolderExist, err)
		}
	})

	t.Run("bad_path", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_bad_path")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "bad", "path")

		f := New()

		err = f.CreateFolder(folderPath)
		if err == nil {
			t.Fatal("bad path, must be error")
		}
		if !errors.Is(err, fs.ErrorBadPath) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorBadPath, err)
		}
	})

	t.Run("permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_permissions")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		// Forbid creating something inside
		err = chmod(dir, 0400)
		if err != nil {
			t.Fatal("error updating permission", err)
		}

		f := New()
		err = f.CreateFolder(filepath.Join(dir, "permissions"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})
}
func TestCreateAllFolder(t *testing.T) {
	// Creating a single folder in an existing folder.
	t.Run("success_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_all_folder_success_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := New()
		err = f.CreateAllFolder(filepath.Join(dir, "success"))
		if err != nil {
			t.Fatal(err)
		}
	})

	// Creating a hierarchy of non-existing folders.
	t.Run("success_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_all_folder_success_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := New()
		err = f.CreateAllFolder(filepath.Join(dir, "middle", "success"))
		if err != nil {
			t.Fatal(err)
		}
	})

	// If the specified folder already exists, there will be no error.
	t.Run("exist_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_all_folder_exist_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "exist")

		f := New()
		err = f.CreateAllFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateAllFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("exist_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_all_folder_exist_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "exist")

		f := New()
		err = f.CreateFile(folderPath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateAllFolder(folderPath)
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorFolderExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorFolderExist, err)
		}
	})

	t.Run("permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_all_folder_permissions")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		// Forbid creating something inside
		err = chmod(dir, 0400)
		if err != nil {
			t.Fatal("error updating permission", err)
		}

		f := New()
		err = f.CreateAllFolder(filepath.Join(dir, "permissions"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})
}
func TestRemoveFolder(t *testing.T) {
	// Removing a single folder.
	t.Run("success_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_success_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "success")

		f := New()
		err = f.CreateAllFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.RemoveFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}
	})

	// Removing a hierarchy folders.
	t.Run("success_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_success_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "middle", "success")

		f := New()
		err = f.CreateAllFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.RemoveFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "not", "exist")

		f := New()

		err = f.RemoveFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}
	})

	// RemoveAll also allows you to delete a file.
	t.Run("file", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_file")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "file")

		f := New()
		err = f.CreateFile(folderPath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.RemoveFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("permissions_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_permissions_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

		// Forbid removing
		err = chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = chmod(dir, CreateDirPerm)
			if err != nil {
				t.Fatal("error updating permission", err)
			}
		}()

		err = f.RemoveFolder(folderPath)
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})

	// Even if a user doesn't have write permissions, but it's their file, they have the right to delete it.
	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := filepath.Join(dir, "permissions")

		f := New()
		err = f.CreateFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

		// Forbid writing
		err = chmod(folderPath, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}

		err = f.RemoveFolder(folderPath)
		if runtime.GOOS == "windows" {
			if err == nil {
				t.Fatal("on windows must be error!")
			}
			if !errors.Is(err, fs.ErrorPermission) {
				t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}
func TestRename(t *testing.T) {
	t.Run("file success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_file_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldFilePath := filepath.Join(dir, "old_success")
		newFilePath := filepath.Join(dir, "new_success")
		data := []byte("hello")

		f := New()
		err = f.CreateFile(oldFilePath, data)
		if err != nil {
			t.Fatal(err)
		}

		err = f.Rename(oldFilePath, newFilePath)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := f.ReadFile(newFilePath)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(data, resp) {
			t.Fatalf("bad data readed; got: %q, want: %q", string(resp), string(data))
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_file_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldFilePath := filepath.Join(dir, "old_not_exist")
		newFilePath := filepath.Join(dir, "new_not_exist")

		f := New()
		err = f.Rename(oldFilePath, newFilePath)
		if err == nil {
			t.Fatal("file not exist, must be error")
		}
		if !errors.Is(err, fs.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorNotExist, err)
		}
	})

	t.Run("file_permissions_src", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_file_permissions_src")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldFilePath := filepath.Join(dir, "old_file_permissions")
		newFilePath := filepath.Join(dir, "new_file_permissions")

		f := New()
		err = f.CreateFile(oldFilePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		// Forbid writing
		err = chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = chmod(dir, CreateDirPerm)
			if err != nil {
				t.Fatal("error updating permission", err)
			}
		}()

		err = f.Rename(oldFilePath, newFilePath)
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})

	t.Run("dir_permissions_src", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_dir_permissions_src")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldPath := filepath.Join(dir, "old_dir_permissions")
		newPath := filepath.Join(dir, "new_dir_permissions")

		f := New()
		err = f.CreateFolder(oldPath)
		if err != nil {
			t.Fatal(err)
		}

		// Forbid writing
		err = chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = chmod(dir, CreateDirPerm)
			if err != nil {
				t.Fatal("error updating permission", err)
			}
		}()

		err = f.Rename(oldPath, newPath)
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fs.ErrorPermission) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorPermission, err)
		}
	})

	// Old file just remove new and replace it
	t.Run("file_file_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_file_file_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldFilePath := filepath.Join(dir, "old_file_permissions")
		newFilePath := filepath.Join(dir, "new_file_permissions")

		f := New()
		err = f.CreateFile(oldFilePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}
		err = f.CreateFile(newFilePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.Rename(oldFilePath, newFilePath)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("file_folder_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_file_folder_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldFilePath := filepath.Join(dir, "old_file_permissions")
		newFilePath := filepath.Join(dir, "new_file_permissions")

		f := New()
		err = f.CreateFile(oldFilePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}
		err = f.CreateFolder(newFilePath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.Rename(oldFilePath, newFilePath)
		if err == nil {
			t.Fatal("folder already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorExist, err)
		}
	})

	t.Run("folder_folder_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_folder_folder_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldFilePath := filepath.Join(dir, "old_file_permissions")
		newFilePath := filepath.Join(dir, "new_file_permissions")

		f := New()
		err = f.CreateFolder(oldFilePath)
		if err != nil {
			t.Fatal(err)
		}
		err = f.CreateFolder(newFilePath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.Rename(oldFilePath, newFilePath)
		if err == nil {
			t.Fatal("folder already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorExist, err)
		}
	})

	t.Run("folder_file_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "rename_folder_file_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		oldFilePath := filepath.Join(dir, "old_file_permissions")
		newFilePath := filepath.Join(dir, "new_file_permissions")

		f := New()
		err = f.CreateFolder(oldFilePath)
		if err != nil {
			t.Fatal(err)
		}
		err = f.CreateFile(newFilePath, []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.Rename(oldFilePath, newFilePath)
		if err == nil {
			t.Fatal("folder already exist, must be error")
		}
		if !errors.Is(err, fs.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fs.ErrorExist, err)
		}
	})
}

func chmod(name string, mode os.FileMode) error {
	if runtime.GOOS == "windows" {
		return acl.Chmod(name, mode)
	}
	return os.Chmod(name, mode)
}
