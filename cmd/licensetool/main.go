//	licensetool status [--env-file .env]
//	licensetool fetch  [--env-file .env] [--server URL] [--target TARGET]
//	licensetool elect agpl-3.0-only [--env-file .env]
//
// Exit codes: 0 ok · 1 unexpected error · 2 nothing elected / no API key ·
// 3 token present but invalid or expired · 4 workspace not commercially
// entitled · 5 conflicting election · 6 unsupported election value · 64 usage.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/licensebootstrap"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/licensetoken"
)

const (
	exitOK           = 0
	exitError        = 1
	exitNothing      = 2
	exitInvalidToken = 3
	exitNotEntitled  = 4
	exitConflict     = 5
	exitUnsupported  = 6
	exitUsage        = 64

	envToken     = "SPEAKEASY_LICENSE_TOKEN"
	envElection  = "SPEAKEASY_GENERATED_LICENSE"
	agplElection = "agpl-3.0-only"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return exitUsage
	}
	switch args[0] {
	case "status":
		return runStatus(args[1:])
	case "fetch":
		return runFetch(args[1:])
	case "elect":
		return runElect(args[1:])
	case "-h", "--help", "help":
		usage()
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", args[0])
		usage()
		return exitUsage
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: licensetool status|fetch|elect [flags]")
	fmt.Fprintln(os.Stderr, "  status [--env-file .env]                 report the license election")
	fmt.Fprintln(os.Stderr, "  fetch  [--env-file .env] [--server URL] [--target TARGET]  fetch a license token with the API key from `speakeasy auth login`")
	fmt.Fprintln(os.Stderr, "  elect agpl-3.0-only [--env-file .env]    accept AGPL-3.0-only licensing for generated output")
}

type election struct {
	token   string
	license string
}

func readElection(envFile string) (election, error) {
	current := election{
		token:   strings.TrimSpace(os.Getenv(envToken)),
		license: strings.TrimSpace(os.Getenv(envElection)),
	}
	var err error
	if current.token == "" {
		if current.token, err = licensebootstrap.ReadEnvValue(envFile, envToken); err != nil {
			return current, err
		}
	}
	if current.license == "" {
		if current.license, err = licensebootstrap.ReadEnvValue(envFile, envElection); err != nil {
			return current, err
		}
	}
	return current, nil
}

func runStatus(args []string) int {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	envFile := flags.String("env-file", ".env", "env file holding the election")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	current, err := readElection(*envFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	switch {
	case current.token != "" && current.license != "":
		fmt.Fprintf(os.Stderr, "conflicting election: both %s and %s are set; keep exactly one\n", envToken, envElection)
		return exitConflict
	case current.token != "":
		info, err := licensetoken.Inspect([]byte(current.token))
		if err != nil {
			fmt.Fprintf(os.Stderr, "license token is not usable: %v\n", err)
			return exitInvalidToken
		}
		fmt.Println(describe(info))
		return exitOK
	case current.license == agplElection:
		fmt.Printf("license: %s (explicit election via %s)\n", agplElection, envElection)
		return exitOK
	case current.license != "":
		fmt.Fprintf(os.Stderr, "unsupported %s value %q (expected %s)\n", envElection, current.license, agplElection)
		return exitUnsupported
	default:
		fmt.Fprintln(os.Stderr, "no license elected for generated output")
		return exitNothing
	}
}

func describe(info licensetoken.TokenInfo) string {
	remaining := time.Until(info.ExpiresAt).Round(time.Hour)
	days := int(remaining.Hours() / 24)
	return fmt.Sprintf("license: commercial (workspace %s, org %s, tier %s, targets %s) expires %s (%d days)",
		info.WorkspaceSlug, info.OrgSlug, info.Tier, strings.Join(info.Targets, ", "), info.ExpiresAt.UTC().Format(time.RFC3339), days)
}

func runFetch(args []string) int {
	flags := flag.NewFlagSet("fetch", flag.ContinueOnError)
	envFile := flags.String("env-file", ".env", "env file to write the token into")
	server := flags.String("server", "", "Speakeasy platform URL (default SPEAKEASY_SERVER_URL or "+licensebootstrap.DefaultServerURL+")")
	target := flags.String("target", "", "generator target to request access for")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	apiKey, err := licensebootstrap.ResolveAPIKey(os.Getenv, home)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, licensebootstrap.ErrNoAPIKey) {
			return exitNothing
		}
		return exitError
	}

	serverURL := licensebootstrap.ResolveServerURL(*server, os.Getenv)
	fmt.Fprintf(os.Stderr, "contacting %s\n", serverURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var token string
	if *target != "" {
		var details licensebootstrap.AccessResponse
		token, details, err = licensebootstrap.FetchAccessLicenseToken(ctx, &http.Client{Timeout: 30 * time.Second}, serverURL, apiKey, *target)
		if err == nil && !details.GenerationAllowed {
			fmt.Fprintf(os.Stderr, "warning from the platform for target %s: %s\n", *target, strings.TrimSpace(details.Message))
		}
		if err != nil {
			switch {
			case errors.Is(err, licensebootstrap.ErrNotEntitled):
				message := strings.TrimSpace(details.Message)
				if message == "" {
					message = err.Error()
				}
				fmt.Fprintln(os.Stderr, message)
				return exitNotEntitled
			case errors.Is(err, licensebootstrap.ErrAccessTokenMissing):
				fmt.Fprintln(os.Stderr, err)
				return exitNotEntitled
			case errors.Is(err, licensebootstrap.ErrUnauthorized):
				fmt.Fprintln(os.Stderr, err)
				return exitNothing
			default:
				fmt.Fprintln(os.Stderr, err)
				return exitError
			}
		}
	} else {
		var details licensebootstrap.ValidateResponse
		token, details, err = licensebootstrap.FetchLicenseToken(ctx, &http.Client{Timeout: 30 * time.Second}, serverURL, apiKey)
		if err != nil {
			switch {
			case errors.Is(err, licensebootstrap.ErrNotEntitled):
				fmt.Fprintf(os.Stderr, "%v (workspace %s, account type %s)\n", err, details.WorkspaceSlug, details.AccountType)
				return exitNotEntitled
			case errors.Is(err, licensebootstrap.ErrUnauthorized):
				fmt.Fprintln(os.Stderr, err)
				return exitNothing
			default:
				fmt.Fprintln(os.Stderr, err)
				return exitError
			}
		}
	}

	info, err := licensetoken.Inspect([]byte(token))
	if err != nil {
		fmt.Fprintf(os.Stderr, "the platform issued a token this generator cannot validate: %v\n", err)
		return exitInvalidToken
	}

	if err := licensebootstrap.WriteElection(*envFile, envToken, token, envElection); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	fmt.Println(describe(info))
	fmt.Printf("wrote %s to %s\n", envToken, *envFile)
	return exitOK
}

func runElect(args []string) int {
	flags := flag.NewFlagSet("elect", flag.ContinueOnError)
	envFile := flags.String("env-file", ".env", "env file to write the election into")
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprintf(os.Stderr, "usage: licensetool elect %s [--env-file .env]\n", agplElection)
		return exitUsage
	}
	license := args[0]
	if err := flags.Parse(args[1:]); err != nil {
		return exitUsage
	}
	if license != agplElection {
		fmt.Fprintf(os.Stderr, "unsupported license %q: only %s can be elected; commercial output requires a license token (licensetool fetch)\n", license, agplElection)
		return exitUsage
	}

	if err := licensebootstrap.WriteElection(*envFile, envElection, agplElection, envToken); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	fmt.Printf("license: %s (wrote %s to %s)\n", agplElection, envElection, *envFile)
	return exitOK
}
