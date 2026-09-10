package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/licensetoken"
)

func ResolveGenerationContextFactory(licenseElection, licenseTokenPath string, getenv func(string) string) func() (context.Context, error) {
	election := strings.TrimSpace(licenseElection)
	if election == "" {
		election = strings.TrimSpace(getenv("SPEAKEASY_GENERATED_LICENSE"))
	}
	envToken := strings.TrimSpace(getenv("SPEAKEASY_LICENSE_TOKEN"))
	tokenPresent := licenseTokenPath != "" || envToken != ""

	switch {
	case election != "" && election != "agpl-3.0-only" && election != "commercial":
		return errorGenerationContext(fmt.Errorf("unknown --license value %q: expected agpl-3.0-only or commercial", election))
	case election == "agpl-3.0-only" && tokenPresent:
		return errorGenerationContext(errors.New("--license agpl-3.0-only conflicts with a supplied license token: choose one"))
	case election == "commercial" && !tokenPresent:
		return errorGenerationContext(errors.New("--license commercial requires a license token via --license-token or SPEAKEASY_LICENSE_TOKEN"))
	case election == "agpl-3.0-only":
		return agplElectionGenerationContext
	case licenseTokenPath != "":
		return licenseTokenFileGenerationContext(licenseTokenPath)
	case envToken != "":
		return rawLicenseTokenGenerationContext(envToken)
	default:
		return nil
	}
}

func agplElectionGenerationContext() (context.Context, error) {
	return generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), nil
}

func errorGenerationContext(err error) func() (context.Context, error) {
	return func() (context.Context, error) { return nil, err }
}

func licenseTokenFileGenerationContext(path string) func() (context.Context, error) {
	return func() (context.Context, error) {
		token, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read license token file: %w", err)
		}
		if len(bytes.TrimSpace(token)) == 0 {
			return nil, fmt.Errorf("license token file %s is empty", path)
		}

		return licensetoken.WithToken(context.Background(), token), nil
	}
}

func rawLicenseTokenGenerationContext(token string) func() (context.Context, error) {
	return func() (context.Context, error) {
		return licensetoken.WithToken(context.Background(), []byte(token)), nil
	}
}
