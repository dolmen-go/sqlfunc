// Platforms where modernc.org/sqlite is supported:
// +build linux,386 linux,amd64 linux,arm linux,arm64 darwin,amd64

package sqlfunc_test

import _ "modernc.org/sqlite"

func init() {
	sqliteDriver = "sqlite"
}
