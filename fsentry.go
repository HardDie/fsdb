// Package fsentry fsentry
//
// Allows storing hierarchical data in files and folders on the file system
// with json descriptions and creation/update timestamps.
package fsentry

import (
	binaryService "github.com/HardDie/fsentry/internal/binary"
	entryService "github.com/HardDie/fsentry/internal/entry"
	folderService "github.com/HardDie/fsentry/internal/folder"
	fsStorage "github.com/HardDie/fsentry/internal/fs/storage"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Config struct {
	log      Logger
	root     string
	isPretty bool
}

func WithLogger(log Logger) func(cfg *Config) {
	return func(cfg *Config) {
		if log == nil {
			return
		}
		cfg.log = log
	}
}
func WithPretty() func(cfg *Config) {
	return func(cfg *Config) {
		cfg.isPretty = true
	}
}

func NewFSEntry(root string, ops ...func(fs *Config)) *Service {
	cfg := &Config{
		root: root,
	}
	for _, op := range ops {
		op(cfg)
	}

	fileStorage := fsStorage.New()
	return New(
		cfg.log,
		cfg.root,
		cfg.isPretty,
		fileStorage,
		binaryService.New(fileStorage, cfg.isPretty),
		entryService.New(fileStorage, cfg.isPretty),
		folderService.New(fileStorage, cfg.isPretty),
	)
}
