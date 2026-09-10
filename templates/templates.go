package templates

import (
	"embed"
	"io/fs"
	"time"
)

//go:embed all:templates/*
var TemplateFS embed.FS

type TemplateFile struct {
	fs.File
	mode fs.FileMode
}

func (f *TemplateFile) Stat() (fs.FileInfo, error) {
	if _, err := f.File.Stat(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *TemplateFile) Name() string {
	info, _ := f.File.Stat()
	return info.Name()
}

func (f *TemplateFile) Size() int64 {
	info, _ := f.File.Stat()
	return info.Size()
}

func (f *TemplateFile) Mode() fs.FileMode {
	return f.mode
}

func (f *TemplateFile) ModTime() time.Time {
	info, _ := f.File.Stat()
	return info.ModTime()
}

func (f *TemplateFile) IsDir() bool {
	info, _ := f.File.Stat()
	return info.IsDir()
}

func (f *TemplateFile) Sys() any {
	info, _ := f.File.Stat()
	return info.Sys()
}

type TemplateFileSystem struct{}

var (
	_ fs.FS         = &TemplateFileSystem{}
	_ fs.ReadDirFS  = &TemplateFileSystem{}
	_ fs.ReadFileFS = &TemplateFileSystem{}
)

func (t *TemplateFileSystem) Open(name string) (fs.File, error) {
	file, err := TemplateFS.Open(name)
	if err != nil {
		return nil, err
	}

	return &TemplateFile{file, fileMode[name]}, nil
}

func Mode(name string) fs.FileMode {
	return fileMode[name]
}

func (t *TemplateFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return TemplateFS.ReadDir(name)
}

func (t *TemplateFileSystem) ReadFile(name string) ([]byte, error) {
	return TemplateFS.ReadFile(name)
}
