// Copyright 2017 Autho"os"

//r: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"regexp"
	"os"
	"strings"
	"path/filepath"
	"testing"
)


func TestObserver_Parse_CS(t *testing.T) {
	const fFileName = "test.cs"

	dict := NewDictionary("dict/cs.dct")
	dict.genRules[regexp.MustCompile("{package_name}")] = []byte("test")
	enc  := NewEncoder("win1251")
	obs := NewObserver(enc, dict)
	ioReader, _ := os.Open(fFileName)
	ioWriter, _ := os.Create("temp/" + strings.TrimSuffix( filepath.Base(fFileName), filepath.Ext(fFileName)) +".go")

	// insert package name self
	ioWriter.Write([]byte("package test\n"))
	obs.Parse(ioReader, ioWriter)

}
