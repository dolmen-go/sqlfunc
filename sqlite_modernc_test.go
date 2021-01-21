// Platforms where modernc.org/sqlite is supported:
// +build linux,386 linux,amd64 linux,arm linux,arm64 darwin,amd64
//
// Allow to disable use of modernc.org/sqlite:
//    go test -v -count=1 -tags nomodernc
// +build !nomodernc

package sqlfunc_test

import _ "modernc.org/sqlite"

func init() {
	sqliteDriver = "sqlite"
}
