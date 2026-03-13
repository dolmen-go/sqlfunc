/*
Copyright 2026 Olivier Mengué

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package genutils_test

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/dolmen-go/sqlfunc/internal/genutils"
)

func TestIsGeneratedFile(t *testing.T) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	rootFS := root.FS()

	const (
		fNotGen = "doc.go"
		fGen    = "sqlfunc_gen_x_test.go"
	)

	isGenerated, err := genutils.IsFileGenerated(rootFS, fNotGen)
	if err != nil {
		t.Fatal("../../"+fNotGen+":", err)
	}
	if isGenerated {
		t.Error("../../" + fNotGen + " isn't a generated file.")
	}

	if err != nil {
		t.Fatal("../../"+fGen+":", err)
	}
	isGenerated, err = genutils.IsFileGenerated(rootFS, fGen)
	if !isGenerated {
		t.Error("../../" + fGen + " is a generated file.")
	}
}

func TestWriteFS(t *testing.T) {
	fs := fstest.MapFS{
		"foo.txt": &fstest.MapFile{Mode: 0666},
		"bar.txt": &fstest.MapFile{Mode: 0666},
	}

	outDir, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal("OpenRoot:", err)
	}

	err = genutils.WriteFS(outDir, fs)
	if err != nil {
		t.Fatal("WriteFS:", err)
	}

	for f := range fs {
		_, err := outDir.Stat(f)
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
	}
}
