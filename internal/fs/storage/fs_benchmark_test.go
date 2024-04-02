package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkRawCreateFile(b *testing.B) {
	dir, err := os.MkdirTemp("", "benchmark_create_file")
	if err != nil {
		b.Fatal("error creating temp dir", err)
	}
	defer os.RemoveAll(dir)

	for i := 0; i < b.N; i++ {
		file, err := os.Create(filepath.Join(dir, strconv.Itoa(i)))
		if err != nil {
			b.Fatal("error create:", err)
		}
		file.Write([]byte{byte(i), byte(i), byte(i), byte(i)})
		err = file.Close()
		if err != nil {
			b.Fatal("error close:", err)
		}
	}
}
func BenchmarkRawUpdateFile(b *testing.B) {
	dir, err := os.MkdirTemp("", "benchmark_update_file")
	if err != nil {
		b.Fatal("error creating temp dir", err)
	}
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, "update.bin")
	file, err := os.Create(filePath)
	if err != nil {
		b.Fatal("error create:", err)
	}
	err = file.Close()
	if err != nil {
		b.Fatal("error close:", err)
	}

	for i := 0; i < b.N; i++ {
		file, err = os.OpenFile(filePath, UpdateFileFlags, CreateFilePerm)
		if err != nil {
			b.Fatal("error open:", err)
		}
		_, err = file.Write([]byte{byte(i), byte(i), byte(i), byte(i)})
		if err != nil {
			b.Fatal("error write:", err)
		}
		err = file.Close()
		if err != nil {
			b.Fatal("error close:", err)
		}
	}
}
func BenchmarkCreateFile(b *testing.B) {
	dir, err := os.MkdirTemp("", "benchmark_create_file")
	if err != nil {
		b.Fatal("error creating temp dir", err)
	}
	defer os.RemoveAll(dir)

	fs := New()

	for i := 0; i < b.N; i++ {
		err = fs.CreateFile(filepath.Join(dir, strconv.Itoa(i)), []byte{byte(i), byte(i), byte(i), byte(i)})
		if err != nil {
			b.Fatal("error create:", err)
		}
	}
}
func BenchmarkUpdateFile(b *testing.B) {
	dir, err := os.MkdirTemp("", "benchmark_update_file")
	if err != nil {
		b.Fatal("error creating temp dir", err)
	}
	defer os.RemoveAll(dir)

	fs := New()

	filePath := filepath.Join(dir, "update.bin")
	err = fs.CreateFile(filePath, nil)
	if err != nil {
		b.Fatal("error create:", err)
	}

	for i := 0; i < b.N; i++ {
		err = fs.UpdateFile(filePath, []byte{byte(i), byte(i), byte(i), byte(i)})
		if err != nil {
			b.Fatal("error update:", err)
		}
	}
}
func BenchmarkMoveFile(b *testing.B) {
	dir, err := os.MkdirTemp("", "benchmark_move_file")
	if err != nil {
		b.Fatal("error creating temp dir", err)
	}
	defer os.RemoveAll(dir)

	fs := New()

	currentName := "-1"
	nextName := ""

	err = fs.CreateFile(filepath.Join(dir, currentName), nil)
	if err != nil {
		b.Fatal("error create:", err)
	}

	for i := 0; i < b.N; i++ {
		nextName = strconv.Itoa(i)
		err = fs.Rename(filepath.Join(dir, currentName), filepath.Join(dir, nextName))
		if err != nil {
			b.Fatal("error move:", err)
		}
		currentName = nextName
	}
}
func BenchmarkUpdateViaMoveFile(b *testing.B) {
	dir, err := os.MkdirTemp("", "benchmark_update_via_move_file")
	if err != nil {
		b.Fatal("error creating temp dir", err)
	}
	defer os.RemoveAll(dir)

	fs := New()

	filePath := filepath.Join(dir, "update.bin")
	err = fs.CreateFile(filePath, nil)
	if err != nil {
		b.Fatal("error create:", err)
	}

	for i := 0; i < b.N; i++ {
		err = fs.CreateFile(filePath+".new", []byte{byte(i), byte(i), byte(i), byte(i)})
		if err != nil {
			b.Fatal("error create new file:", err)
		}
		err = fs.Rename(filePath, filePath+".old")
		if err != nil {
			b.Fatal("error backup old file:", err)
		}
		err = fs.Rename(filePath+".new", filePath)
		if err != nil {
			b.Fatal("error move new file as main:", err)
		}
		err = fs.RemoveFile(filePath + ".old")
		if err != nil {
			b.Fatal("error remove old file:", err)
		}
	}
}
