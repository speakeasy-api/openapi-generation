package tests

import (
	"io/fs"
	"os"
	"path/filepath"
)

func GetTestFilesDir(dir string) string {
	return filepath.Join(dir, ".speakeasy", "testfiles") // TODO we will need to deal with .gen eventually
}

func GenerateExampleFileIfNeeded(dir string) bool {
	foundExampleFile := false
	exampleFile := "example.file"

	_ = os.MkdirAll(dir, os.ModePerm)

	_ = fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if path == exampleFile {
			foundExampleFile = true
		}

		return nil
	})

	return !foundExampleFile
}
