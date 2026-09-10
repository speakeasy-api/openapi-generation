package languageserver

import (
	"errors"
	"io"
	"os"

	"github.com/speakeasy-api/openapi-generation/v2/languageserver/server"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

type WrappedLanguageServer struct {
	server.LanguageServer
	stdout *os.File
	stderr *os.File
}

func NewServer(version string, _ logging.Logger, fs filesystem.FileSystem) WrappedLanguageServer {
	// Logger isn't relevant in this context: we're keeping it to maintain the interface compatibility
	// We need to capture the stdout and stderr as the generator is leaking output to them that interferes with the LSP
	stdout := os.Stdout
	stderr := os.Stderr

	_, stdoutW, _ := os.Pipe()
	_, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	server := server.NewSpeakeasyServer(version, fs)

	return WrappedLanguageServer{
		LanguageServer: *server,
		stdout:         stdout,
		stderr:         stderr,
	}
}

func (s WrappedLanguageServer) Close() error {
	err := os.Stdout.Close()
	if err != nil {
		return err
	}
	err = os.Stderr.Close()
	if err != nil {
		return err
	}

	os.Stdout = s.stdout
	os.Stderr = s.stderr
	return nil
}

func (s WrappedLanguageServer) Run() error {
	s.Server.Log.Info("reading from stdin, writing to stdout")
	s.Server.ServeStream(Stdio{})
	return nil
}

type Stdio struct{}

var _ io.ReadWriteCloser = Stdio{}

func (Stdio) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (Stdio) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (Stdio) Close() error {
	return errors.Join(os.Stdin.Close(), os.Stdout.Close())
}
