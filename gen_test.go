package sqlfunc_test

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dolmen-go/sqlfunc/internal/sqlfuncgen"
)

// Dump fs dir in txtar style
func dumpDir(ffs fs.FS, path string) (string, error) {
	entries, err := fs.ReadDir(ffs, path)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	for _, de := range entries {
		if de.IsDir() {
			continue
		}
		fmt.Fprintf(&buf, "\033[1m-- %s --\033[m\n", de.Name())
		f, err := ffs.Open(filepath.Join(path, de.Name()))
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

func TestGenerate(t *testing.T) {
	fs, err := sqlfuncgen.Generate(sqlfuncgen.NewLogger(t.Log, t.Logf), "pattern=.")
	if err != nil {
		t.Fatal("Generate:", err)
	}
	dump, err := dumpDir(fs, ".")
	if err != nil {
		t.Fatal("Dump:", err)
	}
	t.Logf("\n%s", dump)
}
