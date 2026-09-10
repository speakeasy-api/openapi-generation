package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
	"github.com/speakeasy-api/openapi-generation/v2/internal/changeset"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	"github.com/tcnksm/go-gitconfig"
)

type changesetArgs struct {
	generatorChange    bool
	features           []features.Feature
	targets            []string
	conventionalCommit *conventionalcommits.ConventionalCommit
	githubUserName     string
	bump               string
}

func main() {
	args := os.Args[1:]

	parsedArgs := parseArgs(args)
	if parsedArgs == nil {
		os.Exit(1)
	}

	if err := createChangeset(parsedArgs); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func validateRepoRoot() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to determine current working directory: %w", err)
	}

	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("unable to determine git repository root: %w", err)
	}

	repoRoot := strings.TrimSpace(string(output))
	if wd != repoRoot {
		return fmt.Errorf("please run this command from the root of the current git repository:\n  %s", repoRoot)
	}

	return nil
}

func createChangeset(args *changesetArgs) error {
	if args.generatorChange {
		return recordGeneratorChange(args)
	}

	bump := args.bump
	if bump == "" {
		versionBump := args.conventionalCommit.VersionBump(conventionalcommits.DefaultStrategy)

		if versionBump == conventionalcommits.MinorVersion {
			versionBump = conventionalcommits.PatchVersion
		}

		switch versionBump {
		case conventionalcommits.MajorVersion:
			bump = "major"
		default:
			bump = "patch"
		}
	}

	featureNames := make([]string, len(args.features))
	for i, f := range args.features {
		featureNames[i] = f.String()
	}

	cs := &changeset.Changeset{
		ID:          changeset.NewID(),
		Features:    featureNames,
		Targets:     args.targets,
		Type:        args.conventionalCommit.Type,
		Bump:        bump,
		Description: args.conventionalCommit.Description,
		Author:      args.githubUserName,
		Date:        time.Now().Format("2006-01-02"),
	}

	path, err := changeset.Write(cs)
	if err != nil {
		return err
	}

	fmt.Printf("Created changeset %s\n", path)
	fmt.Printf("  Features:    %s\n", strings.Join(featureNames, ", "))
	fmt.Printf("  Targets:     %s\n", strings.Join(args.targets, ", "))
	fmt.Printf("  Bump:        %s\n", bump)
	fmt.Printf("  Type:        %s\n", args.conventionalCommit.Type)
	fmt.Printf("  Description: %s\n", args.conventionalCommit.Description)

	return nil
}

func recordGeneratorChange(args *changesetArgs) error {
	recordFile := "changerecord.md"

	currentChangeRecordData, err := os.ReadFile(recordFile)
	if err != nil {
		return fmt.Errorf("error reading current changerecord: %w", err)
	}

	entry := fmt.Sprintf("%s - %s: %s", time.Now().Format("2006-01-02"), args.conventionalCommit.Type, args.conventionalCommit.Description)

	parts := strings.Split(string(currentChangeRecordData), "\n- ")

	if len(parts) == 1 {
		parts = append(parts, entry)
	} else {
		parts = append(parts[:2], parts[1:]...)
		parts[1] = entry
	}

	if err := os.WriteFile(recordFile, []byte(strings.Join(parts, "\n- ")), 0o644); err != nil {
		return fmt.Errorf("error writing changerecord: %w", err)
	}

	fmt.Printf("Wrote changerecord entry to file %s\n", recordFile)

	return nil
}

func parseArgs(args []string) *changesetArgs {
	parsedArgs := &changesetArgs{}

	allFeatures := features.GetAllFeatures()

	var featureList strings.Builder
	var templateList strings.Builder

	for _, feature := range allFeatures {
		fmt.Fprintf(&featureList, "\t%s\n", feature)
	}

	for _, language := range templates.GetAvailableTemplates() {
		fmt.Fprintf(&templateList, "\t%s\n", language)
	}

	helpString := fmt.Sprintf(`Usage:

go run cmd/changelog/main.go [--bump <patch|minor|major>] <feature/generator> <template/all> "<changelog message>"

For example: go run cmd/changelog/main.go core go "fix: fixed a bug with json serialization"
Or with explicit bump: go run cmd/changelog/main.go --bump patch core go "fix: fixed a bug"

where <feature> is one of:
%s
or "generator" for non-language specific changes that don't modify SDK output (use "all" when specifying "generator" in the first argument).
"core" should be used when output is modified for all SDKs within the specified language(s).
Multiple features can be specified by separating them with a comma, e.g. "core,flattening"

where <language> is one of:
%s
or "all" for all languages (use "all" when specifying "generator" in the first argument).
Multiple languages can be specified by separating them with a comma, e.g. "go,pythonv2"

where <changelog message> is a string enclosed in quotes using the Conventional Commits format

Optional flags:
  --bump <patch|minor|major>  Explicitly specify the version bump type instead of deriving from commit message`, featureList.String(), templateList.String())

	if err := validateRepoRoot(); err != nil {
		fmt.Println(err)
		fmt.Println()
		fmt.Println(helpString)

		return nil
	}

	githubUserName, err := gitconfig.GithubUser()
	if err != nil && !strings.Contains(err.Error(), "not found") {
		fmt.Println("Error getting Github username from gitconfig!")
		fmt.Println(err)
		fmt.Println()
		fmt.Println(helpString)

		return nil
	}
	if githubUserName == "" {
		fmt.Printf("Please enter your Github username: ")
		_, _ = fmt.Scanln(&githubUserName)

		_ = exec.Command("git", "config", "--global", "github.user", githubUserName).Run()
	}

	parsedArgs.githubUserName = githubUserName

	if len(args) == 0 {
		fmt.Println("No arguments provided!")
		fmt.Println()
		fmt.Println(helpString)

		return nil
	}

	if len(args) >= 2 && args[0] == "--bump" {
		if len(args) < 5 {
			fmt.Println("Incorrect number of arguments provided when using --bump flag!")
			fmt.Println()
			fmt.Println(helpString)

			return nil
		}

		bump := args[1]
		if bump != "patch" && bump != "minor" && bump != "major" {
			fmt.Printf("Invalid bump value: %s. Must be one of: patch, minor, major\n", bump)
			fmt.Println()
			fmt.Println(helpString)

			return nil
		}

		parsedArgs.bump = bump
		args = args[2:]
	}

	if len(args) < 3 {
		fmt.Println("Incorrect number of arguments provided!")
		fmt.Println()
		fmt.Println(helpString)

		return nil
	}

	if args[0] == "generator" && args[1] != "all" {
		fmt.Println("Language must be \"all\" when specifying \"generator\" as the first argument!")
		fmt.Println()
		fmt.Println(helpString)

		return nil
	}

	if args[0] != "generator" {
		feats := strings.Split(args[0], ",")

		for _, feature := range feats {
			if feature == "" {
				fmt.Println("Feature cannot be empty!")
				fmt.Println()
				fmt.Println(helpString)
			}

			if feature == "generator" {
				fmt.Println(`"generator" should be specified alone!`)
				fmt.Println()
				fmt.Println(helpString)

				return nil
			}

			f := features.FeatureFromString(feature)

			if !slices.Contains(allFeatures, f) {
				fmt.Printf(`Feature "%s" is not valid!\n`, feature)
				fmt.Println()
				fmt.Println(helpString)

				return nil
			}

			parsedArgs.features = append(parsedArgs.features, f)
		}
	} else {
		parsedArgs.generatorChange = true
	}

	if args[1] != "all" {
		targets := strings.Split(args[1], ",")

		for _, target := range targets {
			if target == "" {
				fmt.Println("Target cannot be empty!")
				fmt.Println()
				fmt.Println(helpString)

				return nil
			}

			if target == "all" {
				fmt.Println(`"all" should be specified alone!`)
				fmt.Println()
				fmt.Println(helpString)

				return nil
			}

			if !slices.Contains(templates.GetAvailableTemplates(), target) {
				fmt.Printf(`Target "%s" is not valid!\n`, target)
				fmt.Println()
				fmt.Println(helpString)

				return nil
			}

			if replacementVersion, isSunset := templates.GetMinimumTargetVersion()[target]; isSunset {
				fmt.Printf(`Target "%s" is sunset. Updates and changelog entries should be made against "%s%s"\n`, target, target, replacementVersion)
				fmt.Println()
				fmt.Println(helpString)

				return nil
			}

			parsedArgs.targets = append(parsedArgs.targets, target)
		}
	} else {
		parsedArgs.targets = templates.GetAvailableTemplates()
	}

	commitMessage := strings.Join(args[2:], " ")

	res, err := parser.NewMachine(conventionalcommits.WithTypes(conventionalcommits.TypesConventional)).Parse([]byte(commitMessage))
	if err != nil {
		fmt.Println("Error parsing changelog message! Not a valid Conventional Commit Message!")
		fmt.Println(err)
		fmt.Println()
		fmt.Println(helpString)

		return nil
	}

	parsedArgs.conventionalCommit = res.(*conventionalcommits.ConventionalCommit)

	return parsedArgs
}
