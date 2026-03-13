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

package main

import (
	"context"
	"log"
	"os"

	"github.com/dolmen-go/sqlfunc/internal/genutils"
	"github.com/dolmen-go/sqlfunc/internal/sqlfuncgen"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("sqlfunc-gen: ")

	if len(os.Args) > 1 {
		log.Fatal("no flags expected.")
	}

	fsys, err := sqlfuncgen.Generate(context.Background(), sqlfuncgen.NewLogger(log.Println, log.Printf), "pattern=.")
	if err != nil {
		if err == context.Canceled {
			return
		}
		log.Fatal("generate: ", err)
	}

	root, err := os.OpenRoot(".")
	if err != nil {
		log.Fatal("open root: ", err)
	}

	err = genutils.WriteFS(root, fsys)
	if err != nil {
		log.Fatal("write: ", err)
	}
}
