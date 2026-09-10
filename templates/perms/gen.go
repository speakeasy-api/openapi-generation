//go:build ignore

package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"text/template"
)

const (
	outFile = "perms.go"
	out     = `// {{.GenComment}}
//go:generate go run ./perms/gen.go templates/*

package templates

import "io/fs"

var fileMode = map[string]fs.FileMode{
{{- range .Perms}}
	{{.FileName | printf "%q"}}: {{.Mode | printf "0o%o"}},
{{- end}}
}

// IsBinary returns true if the given path is a known binary file.
func IsBinary(path string) bool {
	switch path {
{{- range .BinaryFiles}}
	case {{. | printf "%q"}}:
		return true
{{- end}}
	default:
		return false
	}
}
`
)

var (
	ignoreDirectories = []string{
		"__pycache__",
	}

	// Binary file extensions to check
	binaryExtensions = []string{
		".jar", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".bmp", ".svg",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".zip", ".gz", ".tar", ".tgz", ".bz2", ".7z", ".rar",
		".so", ".dll", ".exe", ".dylib",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".mp3", ".mp4", ".avi", ".mov", ".wmv", ".flv",
		".class", ".pyc", ".o", ".a", ".lib",
	}
)

type perm struct {
	FileName string
	Mode     int
}

func main() {
	perms := findAllPermissions()
	binaryFiles := findBinaryFiles()

	w, err := os.Create(outFile)
	if err != nil {
		log.Panicf("cannot create %q %v", outFile, err)
	}
	defer w.Close()

	fn := getGeneratorFilename()
	generatePermissionsAsset(w, fn, perms, binaryFiles)
	formatPermissionsAsset()
}

// isBinaryFile checks if a file is binary by looking for null bytes in the first 8KB
// or by checking the file extension.
func isBinaryFile(path string) bool {
	// First check by extension
	ext := strings.ToLower(filepath.Ext(path))
	for _, binExt := range binaryExtensions {
		if ext == binExt {
			return true
		}
	}

	// Check file content for null bytes
	f, err := os.Open(path)
	if err != nil {
		return false // If we can't open it, assume it's not binary
	}
	defer f.Close()

	// Read first 8KB
	buf := make([]byte, 8192)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}

	// Look for null bytes
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	return false
}

// findBinaryFiles scans for binary files in the templates directory
func findBinaryFiles() []string {
	var binaryFiles []string
	binarySet := make(map[string]bool)

	// Process glob patterns from command line args
	globs := os.Args[1:]
	if len(globs) == 0 {
		globs = []string{"."} // Default to current directory
	}

	for _, glob := range globs {
		matches, err := fs.Glob(os.DirFS("."), glob)
		if err != nil {
			log.Panicf("cannot glob %q: %v", glob, err)
		}

		for _, match := range matches {
			err := filepath.WalkDir(match, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Skip directories and ignored directories
				if d.IsDir() {
					if slices.Contains(ignoreDirectories, d.Name()) {
						return filepath.SkipDir
					}
					return nil
				}

				// Check if file is binary
				if isBinaryFile(path) {
					binarySet[path] = true
				}

				return nil
			})

			if err != nil {
				log.Panicf("error walking directory %q: %v", match, err)
			}
		}
	}

	// Convert set to sorted slice
	for path := range binarySet {
		binaryFiles = append(binaryFiles, path)
	}
	slices.Sort(binaryFiles)

	return binaryFiles
}

func findAllPermissions() []perm {
	permsMap := map[string]int{}
	dirHasContent := map[string]bool{}

	// Process glob patterns from command line args
	globs := os.Args[1:]
	if len(globs) == 0 {
		globs = []string{"."} // Default to current directory
	}

	for _, glob := range globs {
		matches, err := fs.Glob(os.DirFS("."), glob)
		if err != nil {
			log.Panicf("cannot glob %q: %v", glob, err)
		}

		for _, match := range matches {
			err := filepath.WalkDir(match, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Skip ignored directories
				if d.IsDir() && slices.Contains(ignoreDirectories, d.Name()) {
					return filepath.SkipDir
				}

				info, err := d.Info()
				if err != nil {
					return err
				}

				permsMap[path] = int(info.Mode())

				// If this is a file, mark all parent directories as having content
				if !d.IsDir() {
					markParentDirsAsHavingContent(path, dirHasContent)
				}

				return nil
			})

			if err != nil {
				log.Panicf("error walking directory %q: %v", match, err)
			}
		}
	}

	// Filter out directories that don't have any non-directory content
	perms := make([]perm, 0, len(permsMap))
	for fileName, mode := range permsMap {
		info, err := os.Stat(fileName)
		if err != nil {
			log.Panicf("cannot stat %q: %v", fileName, err)
		}

		if info.IsDir() {
			// Only keep directories that have content (files or non-empty subdirectories)
			if dirHasContent[fileName] {
				perms = append(perms, perm{
					FileName: fileName,
					Mode:     mode,
				})
			}
		} else {
			// Always keep files
			perms = append(perms, perm{
				FileName: fileName,
				Mode:     mode,
			})
		}
	}

	slices.SortStableFunc(perms, func(a perm, b perm) int {
		return strings.Compare(a.FileName, b.FileName)
	})

	return perms
}

func markParentDirsAsHavingContent(filePath string, dirHasContent map[string]bool) {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "/" || dir == filePath {
		return
	}

	dirHasContent[dir] = true
	markParentDirsAsHavingContent(dir, dirHasContent)
}

func getGeneratorFilename() string {
	_, fn, _, ok := runtime.Caller(1)
	if !ok {
		log.Panicf("cannot run caller to find current filename")
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Panicf("unable to get the current working directory: %v", err)
	}

	rd := filepath.Clean(filepath.Join(wd, ".."))

	fn, err = filepath.Rel(filepath.Clean(filepath.Join(wd, "..")), fn)
	if err != nil {
		log.Panicf("unable to strip root directory %q from current filename %q: %v", rd, fn, err)
	}

	return fn
}

func generatePermissionsAsset(
	w io.Writer,
	fn string,
	perms []perm,
	binaryFiles []string,
) {
	tmpl := template.Must(template.New("").Parse(out))
	err := tmpl.Execute(w, map[string]any{
		"GenComment":  fmt.Sprintf("This code is automatically generated by %q. DO NOT EDIT.", fn),
		"Perms":       perms,
		"BinaryFiles": binaryFiles,
	})
	if err != nil {
		log.Panicf("cannot execute perms template: %v", err)
	}
}

func formatPermissionsAsset() {
	err := exec.Command("go", "fmt", outFile).Run()
	if err != nil {
		log.Panicf("cannot go fmt %q: %v", outFile, err)
	}
}
