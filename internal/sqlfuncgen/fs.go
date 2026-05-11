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
	"io"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"time"
)

type genFS map[dirEntry]*Generator

func (genfs *genFS) addFile(name string, gen *Generator) {
	if !fs.ValidPath(name) {
		panic("not a valid path")
	}
	if strings.Contains(name, "/") {
		panic("subdirectories are not supported")
	}
	if *genfs == nil {
		*genfs = make(genFS)
	}
	(*genfs)[dirEntry(name)] = gen
}

func (genfs genFS) Open(name string) (fs.File, error) {
	de := dirEntry(name)
	if gen := genfs[de]; gen != nil {
		return &file{dirEntry: de, generator: gen}, nil
	}

	if name == "." {
		root := &rootDir{dirEntry: ".", fs: genfs}
		if genfs == nil { // empty map
			root.entries = []fs.DirEntry{}
		}
		return root, nil
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

func (d dirEntry) IsDir() bool {
	return d == "."
}

func (d dirEntry) Type() fs.FileMode {
	if d.IsDir() { // rootDir
		return fs.ModeDir
	}
	return d.Mode().Type()
}

func (de dirEntry) Info() (fs.FileInfo, error) {
	return de, nil
}

// Methods for fs.FileInfo

func (d dirEntry) Mode() fs.FileMode {
	if d.IsDir() { // rootDir
		return fs.ModeDir | 0555
	}
	return 0444
}

func (_ dirEntry) ModTime() time.Time {
	return time.Now().Round(time.Hour).Add(-12 * time.Hour)
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

type rootDir struct {
	dirEntry
	fs      genFS
	entries []fs.DirEntry
}

func (d *rootDir) Stat() (fs.FileInfo, error) {
	return d.dirEntry.Info()
}

func (_ rootDir) Read(b []byte) (int, error) {
	return -1, fs.ErrInvalid
}

func (d *rootDir) Close() error {
	d.fs = nil
	d.entries = nil
	return nil
}

func (d *rootDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.entries == nil {
		if d.fs == nil {
			return nil, fs.ErrClosed
		}
		d.entries, _ = d.fs.ReadDir(string(d.dirEntry))
	}
	if n <= 0 {
		entries := d.entries
		d.entries = []fs.DirEntry{}
		return entries, nil
	}
	if len(d.entries) == 0 {
		d.Close()
		return nil, io.EOF
	}
	m := min(n, len(d.entries))
	entries := d.entries[:m:m]
	d.entries = d.entries[m:]
	return entries, nil
}
