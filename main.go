// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// goimportsubtree is a utility designed to be used with go:generate to import
// all packages in a sub-folder for side-effect imports.
//
// Using github.com/progrium/go-extpoints you can have a project with a plugins/
// folder, where plugins register themselves using go-extpoints as a side-effect
// of being imported.
// When combined with go-import-subtree you can import all packages in the
// plugins/ folder automatically. Freeing you from maintaining a file importing
// all your plugins, just run 'go generate'.

package main

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/docopt/docopt-go"
)

const version = "go-import-subtree 1.0.0"
const usage = `
Usage: go-import-subtree [options] [--] <folder> [<folder> ...]

Creates a go file with side-effect imports for all sub-folders in a folder.

Options:
  -V, --version            Display the version of go-import-subtree and exit.
  -h, --help               Print this help information.
  -r, --recursive          Import sub-trees recursively.
  -o, --output=<file>      Output file to write import statements to
                           [default: subtree_imports.go].

Report bugs to https://github.com/jonasfj/go-import-subtree/issues
`

func renderImports(b *bytes.Buffer, importPath string, folder string, recursive bool) {
	// List contents of folder
	entries, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatalf("Couldn't list contents of folder: %s, error: %s", folder, err)
	}

	for _, f := range entries {
		if f.IsDir() {
			subImportPath := path.Join(importPath, f.Name())
			log.Println(subImportPath)
			line := fmt.Sprintf("import _ \"%s\"\n", subImportPath)
			b.WriteString(line)
			if recursive {
				renderImports(b, subImportPath, filepath.Join(folder, f.Name()), recursive)
			}
		}
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("import-subtree: ")

	// Parse docopt string
	args, _ := docopt.Parse(usage, nil, true, version, false, true)
	outputPath := args["--output"].(string)
	recursive := args["--recursive"].(bool)
	folders := args["<folder>"].([]string)

	// Get working directory
	currentFolder, err := os.Getwd()
	if err != nil {
		log.Fatalf("Unable to obtain current working directory: %s", err)
	}

	// Read current package
	pkg, err := build.ImportDir(currentFolder, build.AllowBinary)
	if err != nil {
		log.Fatalf("Failed to import current package: %s", err)
	}
	log.Printf("Identified current package as: %s", pkg.Name)
	log.Printf("Determined current import path: %s", pkg.ImportPath)

	// Generate source
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("package %s\n", pkg.Name))
	log.Println("Finding sub-packages to import:")
	for _, folder := range folders {
		importPath := path.Join(pkg.ImportPath, folder)
		folder = filepath.Join(currentFolder, folder)
		renderImports(&b, importPath, folder, recursive)
	}

	// Run go.fmt to format source
	output, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatalf("Failed to format source, internal error: %s", err)
	}

	// Write output
	ioutil.WriteFile(outputPath, output, 0644)
	if err != nil {
		log.Fatalf("Failed to write output file %s: %s", outputPath, err)
	}
}
