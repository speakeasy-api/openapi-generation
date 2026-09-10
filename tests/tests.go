package main

import (
	_ "embed"
	"flag"
	"fmt"
)

func main() {
	// common flags
	mode := flag.String("mode", "coverage", "mode to run (coverage, usage)")
	lang := flag.String("lang", "go", "language to check (csharp, go, javav2, mockserver, php, pythonv2, ruby, typescriptv2, unity)")

	// usage flags
	group := flag.String("group", "primary", "test group to check (determining which spec file to check) - only used in usage mode")
	all := flag.Bool("all", false, "whether to build example codes for all operations - only used in usage mode")

	flag.Parse()

	switch *mode {
	case "coverage":
		checkTestCoverage(*lang)
	case "usage":
		testUsage(*lang, *group, *all)
	default:
		panic(fmt.Sprintf("unknown mode %q", *mode))
	}
}
