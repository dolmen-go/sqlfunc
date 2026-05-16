package sqlfunc_test

import (
	"bytes"
	"fmt"
	"io"
	iofs "io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/dolmen-go/sqlfunc/internal/sqlfuncgen"
)

// Dump fs dir in txtar style
func dumpDir(fs iofs.FS, path string) (string, error) {
	entries, err := iofs.ReadDir(fs, path)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	for _, de := range entries {
		if de.IsDir() {
			continue
		}
		fmt.Fprintf(&buf, "\033[1m-- %s --\033[m\n", de.Name())
		f, err := fs.Open(filepath.Join(path, de.Name()))
		if err != nil {
			return "", err
		}
		if _, err = io.Copy(&buf, f); err != nil {
			f.Close()
			return "", fmt.Errorf("%s: %w", de.Name(), err)
		}
		f.Close()
	}
	return buf.String(), nil
}

type testingLogWriter func(...any)

func (w testingLogWriter) Write(b []byte) (int, error) {
	b = bytes.TrimRight(b, "\n")
	w(string(b))
	return len(b), nil
}

func TestGenerate(t *testing.T) {
	fs, err := sqlfuncgen.Generate(
		t.Context(),
		slog.New(
			slog.NewTextHandler(testingLogWriter(t.Log), &slog.HandlerOptions{
				Level: slog.LevelDebug,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					// drop time
					if a.Key == slog.TimeKey && len(groups) == 0 {
						return slog.Attr{}
					}
					return a
				},
			}),
		),
		".",
		"pattern=.",
	)
	if err != nil {
		t.Fatal("Generate:", err)
	}
	if err := fstest.TestFS(fs, "sqlfunc_gen_x_test.go"); err != nil {
		t.Fatal(err)
	}
	dump, err := dumpDir(fs, ".")
	if err != nil {
		t.Fatal("Dump:", err)
	}
	t.Logf("\n%s", dump)
}
