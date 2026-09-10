//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

type WASMFileSystem struct {
	writtenFiles sync.Map
	logger       logging.Logger
}

var _ filesystem.FileSystem = &WASMFileSystem{}

func NewFileSystem(logger logging.Logger) *WASMFileSystem {
	return &WASMFileSystem{
		logger:       logger,
		writtenFiles: sync.Map{},
	}
}

func (fs *WASMFileSystem) ReadFile(path string) ([]byte, error) {
	if data, ok := fs.writtenFiles.Load(path); ok {
		// Successfully read - no need to log routine operations
		return data.([]byte), nil
	}
	return nil, os.ErrNotExist
}

func (fs *WASMFileSystem) WriteFile(path string, data []byte, mode os.FileMode) error {
	if fs.logger != nil {
		fs.logger.Info(fmt.Sprintf("🔄 (WriteFile) Writing file: %s", path))
	}
	fs.writtenFiles.Store(path, data)
	return nil
}

func (fs *WASMFileSystem) MkdirAll(path string, mode os.FileMode) error {
	return nil
}

var _ filesystem.File = &WASMFile{}

type WASMFile struct {
	name    string
	content []byte
	fs      *WASMFileSystem
	offset  int // Track read offset
}

func (f *WASMFile) Read(p []byte) (n int, err error) {
	if f.offset >= len(f.content) {
		return 0, io.EOF
	}
	n = copy(p, f.content[f.offset:])
	f.offset += n
	return n, nil
}

func (f *WASMFile) Write(p []byte) (n int, err error) {
	// Append to content
	f.content = append(f.content, p...)
	// Sync back to filesystem
	f.fs.writtenFiles.Store(f.name, f.content)
	return len(p), nil
}

func (f *WASMFile) Close() error {
	// Ensure final content is synced
	if f.fs != nil {
		f.fs.writtenFiles.Store(f.name, f.content)
	}
	return nil
}

func (fs *WASMFileSystem) Open(name string) (fs.File, error) {
	return fs.OpenFile(name, 0, 0)
}

func (fs *WASMFileSystem) OpenFile(name string, flag int, perm os.FileMode) (filesystem.File, error) {
	data, exists := fs.writtenFiles.Load(name)

	// If file doesn't exist and O_CREATE flag is set, create it
	if !exists {
		if flag&os.O_CREATE != 0 {
			if fs.logger != nil {
				fs.logger.Info(fmt.Sprintf("🔄 (Open) Creating new file: %s", name))
			}
			// Create empty file
			emptyData := []byte{}
			fs.writtenFiles.Store(name, emptyData)
			return &WASMFile{
				name:    name,
				content: emptyData,
				fs:      fs,
				offset:  0,
			}, nil
		}

		// File doesn't exist and not creating - this is expected behavior, no need to log
		return nil, os.ErrNotExist
	}

	// File exists, return it
	return &WASMFile{
		name:    name,
		content: data.([]byte),
		fs:      fs,
		offset:  0,
	}, nil
}

type WASMFileInfo struct {
	name string
	size int
}

func (f *WASMFileInfo) Name() string {
	return f.name
}

func (f *WASMFileInfo) Size() int64 {
	return int64(f.size)
}

func (f *WASMFileInfo) IsDir() bool {
	return false
}

func (f *WASMFileInfo) Mode() os.FileMode {
	return 0644
}

func (f *WASMFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *WASMFileInfo) Sys() any {
	return nil
}

func (f *WASMFile) Stat() (fs.FileInfo, error) {
	return &WASMFileInfo{
		name: f.name,
		size: len(f.content),
	}, nil
}

func (f *WASMFile) Name() string {
	return f.name
}

func (fs *WASMFileSystem) Stat(name string) (fs.FileInfo, error) {
	if data, ok := fs.writtenFiles.Load(name); ok {
		return &WASMFileInfo{
			name: name,
			size: len(data.([]byte)),
		}, nil
	}
	return nil, os.ErrNotExist
}

func (fs *WASMFileSystem) OutputFiles() map[string][]byte {
	result := make(map[string][]byte)
	fs.writtenFiles.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.([]byte)
		return true
	})
	return result
}

func (fs *WASMFileSystem) Remove(path string) error {
	fs.writtenFiles.Delete(path)
	return nil
}

func (wfs *WASMFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, os.ErrNotExist
}

func (fs *WASMFileSystem) ScanForGeneratedIDs() (map[string]string, error) {
	// WASM environment doesn't support persistent edits
	return nil, nil
}

type OutputFile struct {
	OperationID string `json:"operationId"`
	Content     string `json:"content"`
}

// OutputSnippetsAsJSON returns the usage snippets as a JSON array of objects
// We filter out other kinds of files in ,WriteFile().
func (fs *WASMFileSystem) OutputSnippetsAsJSON() (string, error) {
	files := []OutputFile{}
	fs.writtenFiles.Range(func(key, value interface{}) bool {
		fileName := key.(string)
		if !strings.Contains(fileName, ".usage.") {
			return true
		}
		operationID := strings.Split(fileName, ".")[0]
		files = append(files, OutputFile{OperationID: operationID, Content: string(value.([]byte))})
		return true
	})

	jsonFiles, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(jsonFiles), nil
}
