package filesystem

import (
	"io"
	"io/fs"
	"os"

	config "github.com/speakeasy-api/sdk-gen-config"
)

// FileSystem is the main file system abstraction used throughout the generator.
// This interface is implemented by the CLI and injected into the Generator.
// This needs to be located in pkg because it is used by the cli usagegen.
type FileSystem interface {
	config.FS
	OpenFile(name string, flag int, perm fs.FileMode) (File, error)
	MkdirAll(path string, perm os.FileMode) error
	ReadDir(name string) ([]fs.DirEntry, error)
	Remove(name string) error

	// ScanForGeneratedIDs scans the output directory for files with @generated-id headers.
	// Returns a map of UUID -> relative file path.
	// This is used to detect when users have moved generated files.
	ScanForGeneratedIDs() (map[string]string, error)
}

type File interface {
	fs.File
	io.Writer
}
