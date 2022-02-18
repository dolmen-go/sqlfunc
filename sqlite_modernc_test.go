// Allow to disable use of modernc.org/sqlite:
//    go test -v -count=1 -tags nomodernc
// Enable only on platforms where modernc.org/sqlite is officially supported:
//
//go:build !nomodernc && ((linux && 386) || (linux && amd64) || (linux && arm) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (freebsd && amd64) || (windows && amd64))
// +build !nomodernc
// +build linux,386 linux,amd64 linux,arm linux,arm64 darwin,amd64 darwin,arm64 freebsd,amd64 windows,amd64

package sqlfunc_test

import _ "modernc.org/sqlite"

func init() {
	sqliteDriver = "sqlite"
}
