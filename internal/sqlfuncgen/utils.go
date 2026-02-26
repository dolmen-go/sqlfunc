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
	"runtime"
	"strings"
)

// alignLineNum tweaks a [text/template] with a long multiline comment that
// allows to align the source line number with the line number from where this
// function is called.
//
//go:noinline
func alignLineNum(template string) string {
	_, _, line, _ := runtime.Caller(1)
	return `{{/*` + strings.Repeat("\n", line-1) + ` */}}` + template
}
