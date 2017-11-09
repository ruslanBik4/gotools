// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"testing"
	"os"
	"path/filepath"
	"strings"
	"regexp"
)
const fFileName = "test.pas"
func TestObserver_Parse(t *testing.T) {

	dict := NewDictionary("dict/pas.dct")
	dict.genRules[regexp.MustCompile("{package_name}")] = []byte("test")
	enc  := NewEncoder("win1251")
	obs := NewObserver(enc, dict)
	ioReader, _ := os.Open(fFileName)
	ioWriter, _ := os.Create("temp/" + strings.TrimSuffix( filepath.Base(fFileName), filepath.Ext(fFileName)) +".go")

	obs.Parse(ioReader, ioWriter)

}
