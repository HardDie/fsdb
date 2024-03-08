package fs

import (
	"errors"
	"os"
	"path"
	"reflect"
	"testing"

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

	t.Run("exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_exist")
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

	t.Run("permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_file_permissions")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		// Forbid creating something inside
		err = os.Chmod(dir, 0400)
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
		err = os.Chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = os.Chmod(dir, CreateDirPerm)
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
		err = os.Chmod(filePath, 0000)
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

		// Forbid reading
		err = os.Chmod(dir, 0000)
		if err != nil {
			t.Fatal("error updating permission", err)
		}
		defer func() {
			err = os.Chmod(dir, CreateDirPerm)
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
		err = os.Chmod(filePath, 0400)
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

	t.Run("exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "create_folder_exist")
		if err != nil {
			t.Fatal("error creating temp dir", err)
		}
		defer os.RemoveAll(dir)

		f := NewFS()
		err = f.CreateFolder(path.Join(dir, "exist"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.CreateFolder(path.Join(dir, "exist"))
		if err == nil {
			t.Fatal("folder already exist, must be error")
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
		err = os.Chmod(dir, 0400)
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
