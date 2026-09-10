package processrunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	generatorErrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap"
)

const (
	ErrDependencyNotFound = generatorErrors.Error("dependency not found")
	ErrDependencyVersion  = generatorErrors.Error("dependency version not met")
)

// An individual command dependency.
type CommandDependency struct {
	// Command name.
	Command string

	// Documentation about how to install the command.
	InstallDocumentation string

	// Verify the dependency version.
	Version *CommandVersionDependency
}

// Verifies the command dependency.
func (d CommandDependency) check(ctx context.Context, dir string) (string, error) {
	if _, err := exec.LookPath(d.Command); err != nil {
		return "Dependency Not Found - " + d.InstallDocumentation, ErrDependencyNotFound.Wrap(fmt.Errorf("%s: %w", d.Command, err))
	}

	return d.checkVersion(ctx, dir)
}

// Verifies the command version dependency.
func (d CommandDependency) checkVersion(ctx context.Context, dir string) (string, error) {
	if d.Version == nil {
		return "", nil
	}

	if err := d.Version.check(ctx, d.Command, dir); err != nil {
		if errors.Is(err, ErrDependencyNotFound) || errors.Is(err, ErrDependencyVersion) {
			return "Dependency Version Mismatch - " + d.InstallDocumentation, err
		}

		return "", err
	}

	return "", nil
}

// Command version dependency verification.
type CommandVersionDependency struct {
	// Arguments to pass to the command to retrieve the version.
	Args []string

	// Minimum supported version for the command.
	MinimumVersion *version.Version

	// Regular expression to extract the version for the command.
	Pattern *regexp.Regexp
}

// Verifies the command version dependency.
func (d CommandVersionDependency) check(ctx context.Context, command string, dir string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, command, d.Args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if env.IsDebug() {
		logging.From(ctx).Debug(fmt.Sprintf("running %s %s in %s", command, strings.Join(d.Args, " "), dir))
	}

	if err := cmd.Run(); err != nil {
		return ErrDependencyNotFound.Wrap(err)
	}

	matches := d.Pattern.FindAllStringSubmatch(stdout.String(), -1)

	if len(matches) == 0 {
		matches = d.Pattern.FindAllStringSubmatch(stderr.String(), -1)
	}

	if len(matches) == 0 {
		logging.From(ctx).Debug(
			fmt.Sprintf("Dependency Version Not Found (%s %s)", command, d.MinimumVersion),
			zap.String("output", stdout.String()),
			zap.String("error", stderr.String()),
		)

		return ErrDependencyVersion.Wrap(fmt.Errorf("failed to find version in output running %s", command))
	}

	matchedVersion, err := version.NewVersion(matches[0][1])
	if err != nil {
		return ErrDependencyVersion.Wrap(err)
	}

	if matchedVersion.LessThan(d.MinimumVersion) {
		return ErrDependencyVersion.Wrap(fmt.Errorf("%s - version %s is less than required version %s", command, matchedVersion, d.MinimumVersion))
	}

	logging.From(ctx).Debug(fmt.Sprintf("Dependency Version Found (%s) - %s", command, matchedVersion))

	return nil
}
