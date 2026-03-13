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

package genutils

import (
	"bufio"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"regexp"
)

// Standard: https://go.dev/s/generatedcode
var generatedFileRE = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)

func IsFileGenerated(root iofs.FS, path string) (bool, error) {
	f, err := root.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	r := bufio.NewReader(&io.LimitedReader{R: f, N: 4096})
	return generatedFileRE.MatchReader(r), nil
}

func WriteFS(root interface {
	FS() iofs.FS
	Create(name string) (*os.File, error)
}, fs iofs.FS) error {
	dir, err := iofs.ReadDir(fs, ".")
	if err != nil {
		return err
	}
	for _, f := range dir {
		if f.IsDir() {
			return fmt.Errorf("%s: directories are not handled", f.Name())
		}

		path := f.Name()
		isGenerated, err := IsFileGenerated(root.FS(), path)
		if !os.IsNotExist(err) {
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if !isGenerated {
				return fmt.Errorf("%s: not a generated file (safety belt)", path)
			}
		}

		fi, err := fs.Open(f.Name())
		if err != nil {
			return fmt.Errorf("open %s: %w", f.Name(), err)
		}
		out, err := root.Create(path)
		if err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		_, err = io.Copy(out, fi)
		if err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}
