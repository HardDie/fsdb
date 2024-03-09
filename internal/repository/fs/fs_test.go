package fs

import (
	"errors"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"

	acl "github.com/hectane/go-acl"

	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

func TestCreateFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_success")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := NewFS()
		err = f.CreateFile(path.Join(dir, "success"), nil)
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

		f := NewFS()
		err = f.CreateFile(path.Join(dir, "exist"), []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFile(path.Join(dir, "exist"), []byte("hello"))
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorExist, err)
		}
	})

	t.Run("exist_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_exist_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := NewFS()
		err = f.CreateFolder(path.Join(dir, "exist"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFile(path.Join(dir, "exist"), []byte("hello"))
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorExist, err)
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

		f := NewFS()
		err = f.CreateFile(path.Join(dir, "permissions"), []byte("hello"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
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

		filePath := path.Join(dir, "success")
		data := []byte("hello")

		f := NewFS()
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

		filePath := path.Join(dir, "invalid")

		f := NewFS()
		err = f.CreateFolder(filePath)
		if err != nil {
			t.Fatal(err)
		}

		_, err = f.ReadFile(filePath)
		if err == nil {
			t.Fatal("not file, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotFile) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotFile, err)
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := NewFS()
		_, err = f.ReadFile(path.Join(dir, "not_exist"))
		if err == nil {
			t.Fatal("file not exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})

	t.Run("permissions_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_permissions_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := path.Join(dir, "permissions")

		f := NewFS()
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
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
		}
	})

	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "read_file_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := path.Join(dir, "permissions")

		f := NewFS()
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
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
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

		filePath := path.Join(dir, "success")
		data := []byte("hello")

		f := NewFS()
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

		filePath := path.Join(dir, "invalid")

		f := NewFS()
		err = f.CreateFolder(filePath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.UpdateFile(filePath, []byte("new"))
		if err == nil {
			t.Fatal("not file, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotFile) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotFile, err)
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := NewFS()
		err = f.UpdateFile(path.Join(dir, "not_exist"), []byte("hello"))
		if err == nil {
			t.Fatal("file not exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})

	t.Run("permissions_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_permissions_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := path.Join(dir, "permissions")

		f := NewFS()
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
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
		}
	})

	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "update_file_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := path.Join(dir, "permissions")

		f := NewFS()
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
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
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

		filePath := path.Join(dir, "success")
		data := []byte("hello")

		f := NewFS()
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
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})

	t.Run("empty_folder", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_empty_folder")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := path.Join(dir, "empty_folder")

		f := NewFS()
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

		filePath := path.Join(dir, "not_empty_folder")

		f := NewFS()
		err = f.CreateAllFolder(path.Join(filePath, "data"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.RemoveFile(filePath)
		if err == nil {
			t.Fatal("not file, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorExist, err)
		}
	})

	t.Run("not_exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_not_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := NewFS()
		err = f.RemoveFile(path.Join(dir, "not_exist"))
		if err == nil {
			t.Fatal("file not exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotExist, err)
		}
	})

	t.Run("permissions_1", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_permissions_1")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := path.Join(dir, "permissions")

		f := NewFS()
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
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
		}
	})

	// Even if a user doesn't have write permissions, but it's their file, they have the right to delete it.
	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_file_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		filePath := path.Join(dir, "permissions")

		f := NewFS()
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

		f := NewFS()
		err = f.CreateFolder(path.Join(dir, "success"))
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

		folderPath := path.Join(dir, "exist")

		f := NewFS()
		err = f.CreateFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFolder(folderPath)
		if err == nil {
			t.Fatal("folder already exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorExist, err)
		}
	})

	t.Run("exist_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_exist_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := NewFS()
		err = f.CreateFile(path.Join(dir, "exist"), []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFolder(path.Join(dir, "exist"))
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorExist) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorExist, err)
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

		f := NewFS()
		err = f.CreateFolder(path.Join(dir, "permissions"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
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

		f := NewFS()
		err = f.CreateAllFolder(path.Join(dir, "success"))
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

		f := NewFS()
		err = f.CreateAllFolder(path.Join(dir, "middle", "success"))
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

		f := NewFS()
		err = f.CreateAllFolder(path.Join(dir, "exist"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateAllFolder(path.Join(dir, "exist"))
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

		f := NewFS()
		err = f.CreateFile(path.Join(dir, "exist"), []byte("hello"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateAllFolder(path.Join(dir, "exist"))
		if err == nil {
			t.Fatal("file already exist, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorNotDirectory) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorNotDirectory, err)
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

		f := NewFS()
		err = f.CreateAllFolder(path.Join(dir, "permissions"))
		if err == nil {
			t.Fatal("don't have permission, must be error")
		}
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
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

		folderPath := path.Join(dir, "success")

		f := NewFS()
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

		folderPath := path.Join(dir, "middle", "success")

		f := NewFS()
		err = f.CreateAllFolder(folderPath)
		if err != nil {
			t.Fatal(err)
		}

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

		folderPath := path.Join(dir, "file")

		f := NewFS()
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

		folderPath := path.Join(dir, "permissions")

		f := NewFS()
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
		if !errors.Is(err, fsentry_error.ErrorPermissions) {
			t.Fatalf("error wait: %q; got: %q", fsentry_error.ErrorPermissions, err)
		}
	})

	// Even if a user doesn't have write permissions, but it's their file, they have the right to delete it.
	t.Run("permissions_2", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "remove_folder_permissions_2")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		folderPath := path.Join(dir, "permissions")

		f := NewFS()
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
		if err != nil {
			t.Fatal(err)
		}
	})
}

func chmod(name string, mode os.FileMode) error {
	if runtime.GOOS == "windows" {
		return acl.Chmod(name, mode)
	}
	return os.Chmod(name, mode)
}
