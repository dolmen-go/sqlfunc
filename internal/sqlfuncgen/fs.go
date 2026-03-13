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

package sqlfuncgen

import (
	"errors"
	"io"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"time"
)

type genFS map[dirEntry]*Generator

func (genfs genFS) addFile(name string, gen *Generator) {
	if !fs.ValidPath(name) {
		panic("not a valid path")
	}
	if strings.Contains(name, "/") {
		panic("subdirectories are not supported")
	}
	genfs[dirEntry(name)] = gen
}

func (genfs genFS) Open(name string) (fs.File, error) {
	de := dirEntry(name)
	if gen := genfs[de]; gen != nil {
		return &file{dirEntry: de, generator: gen}, nil
	}

	if name == "." {
		// TODO forward to ReadDir
		return nil, &fs.PathError{Op: "open", Path: name, Err: errors.New("FIXME use fs.ReadDirFS")}
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (genfs genFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if genfs == nil {
		if name == "." {
			return nil, nil
		}
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	names := slices.Sorted(maps.Keys(genfs))
	de := make([]fs.DirEntry, len(names))
	for i := range names {
		de[i] = dirEntry(names[i])
	}
	return de, nil
}

type dirEntry string

var (
	_ fs.DirEntry = dirEntry("foo")
	_ fs.FileInfo = dirEntry("foo")
)

func (de dirEntry) Name() string {
	return string(de)
}

func (_ dirEntry) IsDir() bool {
	return false
}
func (_ dirEntry) Type() fs.FileMode {
	return 0444
}

func (de dirEntry) Info() (fs.FileInfo, error) {
	return de, nil
}

// Methods for fs.FileInfo

func (_ dirEntry) Mode() fs.FileMode {
	return 0444
}

func (_ dirEntry) ModTime() time.Time {
	return time.Now()
}

func (_ dirEntry) Size() int64 {
	return -1
}

func (_ dirEntry) Sys() any {
	return nil
}

type file struct {
	dirEntry
	generator *Generator
	r         io.Reader
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.dirEntry.Info()
}

func (f *file) Read(b []byte) (int, error) {
	if f.r == nil {
		s, err := f.generator.generateCode()
		if err != nil {
			return 0, err
		}
		f.r = strings.NewReader(s)
	}
	n, err := f.r.Read(b)
	if err != nil {
		f.generator = nil
	}
	return n, err
}

// Close is [io/fs.File.Close].
func (f *file) Close() error {
	f.r = nil
	return nil
}
