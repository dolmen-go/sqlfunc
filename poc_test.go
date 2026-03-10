package sqlfunc_test

import (
	"testing"

	"github.com/dolmen-go/sqlfunc/internal/sqlfuncgen"
)

func TestScanSrc(t *testing.T) {
	err := sqlfuncgen.Generate(t, "pattern=.")
	if err != nil {
		t.Fatal(err)
	}
}
