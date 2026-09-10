package template

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/format"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/templates"
	"go.uber.org/zap"
)

const defaultFileMode = 0o644 // readable by all, writable by owner

// TODO hardcoded magic configuration is usually a bad thing
var usageSnippetFileName = map[string]string{
	"csharp":     "Example.cs",
	"go":         "example.go",
	"java":       "Example.java",
	"php":        "example.php",
	"python":     "example.py",
	"ruby":       "example.rb",
	"typescript": "example.ts",
	"unity":      "Example.cs",
}

func directoryExists(baseDir, dir string) bool {
	_, err := fs.ReadDir(templates.TemplateFS, path.Join(baseDir, dir))
	return err == nil
}

func getDirectoryFiles(baseDir, dir string) ([]string, error) {
	paths := []string{}

	files, err := fs.ReadDir(templates.TemplateFS, path.Join(baseDir, dir))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			subPaths, err := getDirectoryFiles(baseDir, path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}

			paths = append(paths, subPaths...)
		} else {
			paths = append(paths, path.Join(dir, file.Name()))
		}
	}

	return paths, nil
}

func getWriteFileFunc(ctx context.Context, target types.Target, cfg Config) func(string, []byte) error {
	return func(outFile string, data []byte) error {
		var err error
		copiedData, err := copyCustomContent(cfg, outFile, data)
		if err != nil {
			logging.From(ctx).Warn("failed to copy custom content, writing unformatted content", zap.String("file", outFile), zap.Error(err))
			_ = cfg.WriteFileFunc(ctx, outFile, data, defaultFileMode, true)
			return fmt.Errorf("failed to copy custom content: %w", err)
		}

		formattedData, err := format.Format(ctx, target, outFile, copiedData)
		if err != nil {
			logging.From(ctx).Warn("failed to format file, writing unformatted content", zap.String("file", outFile), zap.Error(err))
			_ = cfg.WriteFileFunc(ctx, outFile, copiedData, defaultFileMode, true)
			return fmt.Errorf("failed to format file: %w", err)
		}

		return cfg.WriteFileFunc(ctx, outFile, formattedData, defaultFileMode, true)
	}
}

type regionMarker struct {
	start *regexp.Regexp
	end   *regexp.Regexp
}

var regionMarkers = map[string]regionMarker{
	"typescript": {
		start: regexp.MustCompile(`# ?region ([a-z-]+)$`),
		end:   regexp.MustCompile(`# ?endregion ([a-z-]+)$`),
	},
	"java": {
		start: regexp.MustCompile(`// ?#region ([a-z-]+)$`),
		end:   regexp.MustCompile(`// ?#endregion ([a-z-]+)$`),
	},
	"python": {
		start: regexp.MustCompile(`# ?region ([a-z-]+)$`),
		end:   regexp.MustCompile(`# ?endregion ([a-z-]+)$`),
	},
	"go": {
		start: regexp.MustCompile(`// ?#region ([a-z-]+)$`),
		end:   regexp.MustCompile(`// ?#endregion ([a-z-]+)$`),
	},
	"cli": {
		start: regexp.MustCompile(`// ?#region ([a-z-]+)$`),
		end:   regexp.MustCompile(`// ?#endregion ([a-z-]+)$`),
	},
	"terraform": {
		start: regexp.MustCompile(`// ?#region ([a-z-]+)$`),
		end:   regexp.MustCompile(`// ?#endregion ([a-z-]+)$`),
	},
}

// copyCustomContent will take content that customers added to generated files
// and carry it forward on the newly generated versions of those files. The code
// is expected to live within demarcated regions. For example:
//
//	class SDK extends ClientSDK {
//	  ... generated code ...
//
//	  // #region sdk-body
//	  // This is custom code that will be carried forward.
//	  async greet(): Promise<string> {
//	    return "Hello, world!";
//	  }
//	  // #endregion sdk-body
//	}
func copyCustomContent(cfg Config, fileName string, content []byte) ([]byte, error) {
	// The overall process is as follows:
	// 1. For a newly generated file identified by fileName, extract all custom
	//    code regions from the existing file with the same name. This is a
	//    no-op if the file does not already exist. The code regions are
	//    collected into a map keyed by region ID which comes from the opening
	//    `#region <region-id>` comment.
	// 2. Scan the new content for custom code regions.
	// 3. For each found region, lookup the region ID in the map and and inject
	//    the custom content into the new file content.
	// 4. If there was no custom content for a region, then it is scrubbed from
	//    the new file content. We do not want to leave empty regions all over
	//    the generated SDK codebase.
	// 5. Finally, make some corrections to trailing newlines as this process,
	//    can unintentionally add them.

	featureEnabled := cfg.Subsystem.Features.IsFeatureUsed(features.FeatureCustomCodeRegions)
	var regionStartRE, regionEndRE *regexp.Regexp
	if regionMarker, ok := regionMarkers[cfg.Subsystem.Target.Target]; ok {
		regionStartRE = regionMarker.start
		regionEndRE = regionMarker.end
	}

	existingContent, err := cfg.ReadFileFunc(fileName)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s: failed to read file: %w", fileName, err)
	}

	customContent := make(map[string]string)

	const maxLineSize = 1024 * 1024 // 1MB max line size to handle large generated imports
	srcScanner := bufio.NewScanner(bytes.NewReader(existingContent))
	srcScanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineSize)
	srcScanner.Split(bufio.ScanLines)

	var region strings.Builder
	regionID := ""
	regionLineNo := -1
	lineNo := 0

	for featureEnabled && srcScanner.Scan() {
		lineNo += 1
		line := srcScanner.Text()
		switch {
		case regionStartRE.MatchString(line):
			if regionID != "" {
				return nil, fmt.Errorf("(current) %s:%d: region %s is not closed", fileName, regionLineNo, regionID)
			}

			regionLineNo = lineNo

			match := regionStartRE.FindStringSubmatch(line)
			regionID = strings.TrimSpace(match[1])
			if regionID == "" {
				return nil, fmt.Errorf("(current) %s:%d: region id cannot be empty", fileName, regionLineNo)
			}
			region.Reset()
		case regionEndRE.MatchString(line):
			regionLineNo = -1
			customContent[regionID] = region.String()
			regionID = ""
			region.Reset()
		case regionID != "":
			region.WriteString(line + "\n")
		}
	}

	if regionID != "" {
		return nil, fmt.Errorf("(current) %s:%d: region %s is not closed", fileName, regionLineNo, regionID)
	}

	hasTrailingNewLine := bytes.HasSuffix(content, []byte("\n"))
	dstScanner := bufio.NewScanner(bytes.NewReader(content))
	dstScanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineSize)
	var finalContent strings.Builder
	regionID = ""
	regionLineNo = -1
	lineNo = 0
	emit := false

	for dstScanner.Scan() {
		lineNo += 1
		line := dstScanner.Text()
		switch {
		case regionStartRE != nil && regionStartRE.MatchString(line):
			if regionID != "" {
				return nil, fmt.Errorf("(new) %s:%d: region %s is not closed", fileName, regionLineNo, regionID)
			}

			regionLineNo = lineNo
			match := regionStartRE.FindStringSubmatch(line)
			regionID = strings.TrimSpace(match[1])
			if regionID == "" {
				return nil, fmt.Errorf("(new) %s:%d: region id cannot be empty", fileName, regionLineNo)
			}

			_, emit = customContent[regionID]
			if emit {
				finalContent.WriteString(line + "\n")
			}
		case regionEndRE != nil && regionEndRE.MatchString(line):
			regionContent := customContent[regionID]
			if emit {
				finalContent.WriteString(regionContent)
				finalContent.WriteString(line + "\n")
			}

			regionLineNo = -1
			regionID = ""
			emit = false
		default:
			finalContent.WriteString(line + "\n")
		}
	}

	if regionID != "" {
		return nil, fmt.Errorf("(new) %s:%d: region %s is not closed", fileName, regionLineNo, regionID)
	}

	out := finalContent.String()
	if !hasTrailingNewLine {
		out = strings.TrimSuffix(out, "\n")
	}

	return []byte(out), nil
}

func copyFile(ctx context.Context, baseDir, src, dest string, writeFileFunc WriteFileFunc) error {
	srcFile := path.Join(baseDir, src)
	perm := templates.Mode(srcFile)

	data, err := fs.ReadFile(templates.TemplateFS, srcFile)
	if err != nil {
		return err
	}

	if err := writeFileFunc(ctx, dest, data, perm, false); err != nil {
		return err
	}

	return nil
}
