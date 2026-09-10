package main

import (
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	featuresFilePath := filepath.Join(wd, "features.go")
	testsFilePath := filepath.Join(wd, "tests.go")

	featuresAST, err := parser.ParseFile(token.NewFileSet(), featuresFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("failed to parse features.go file: %v", err)
	}

	testsAST, err := parser.ParseFile(token.NewFileSet(), testsFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("failed to parse tests.go file: %v", err)
	}

	if err := generateFeatures(featuresAST, wd); err != nil {
		log.Fatalf("failed to generate features: %v", err)
	}

	if err := generateTests(testsAST, wd); err != nil {
		log.Fatalf("failed to generate tests: %v", err)
	}

	log.Println("features and tests generated successfully")
}
