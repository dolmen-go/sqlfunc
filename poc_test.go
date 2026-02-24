package sqlfunc_test

import (
	"testing"

	"github.com/dolmen-go/sqlfunc/internal/sqlfuncgen"
)

func TestScanSrc(t *testing.T) {
	sqlfuncgen.Generate(t, "pattern=.")
}
