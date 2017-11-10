// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"golang.org/x/tools/imports"
	"path/filepath"
	"os"
	"strings"
	"log"
	"testing"
	"sync"
	"regexp"
)
const srcPath = "/Users/ruslan/work/RobExt/pas/"
const dstPath = "/Users/ruslan/work/src/bitbucket.org/goext/"
var wrgroup  sync.WaitGroup
func Workfunc(path string, typeFile os.FileMode) error{
	if typeFile.IsDir() {
		return nil
	}
	go doConvert(path)

	return nil
}
func doConvert(path string) {
	var dict *Dictionary
	ext := filepath.Ext(path)
	newPath := strings.TrimSuffix( strings.TrimPrefix( path, srcPath ), ext)

	wrgroup.Add(1)
	defer wrgroup.Done()

	switch ext {
	case  ".pas":
		dict = NewDictionary("dict/pas.dct")
		listDir := strings.Split(filepath.Dir(newPath), "/")
		log.Println(listDir)
		dict.genRules[regexp.MustCompile("{package_name}")] = []byte(listDir[len(listDir)-1])
	case "py":
		dict = NewDictionary("dict/py.dct")
	default:
		return
	}
	enc  := NewEncoder("win1251")
	obs := NewObserver(enc, dict)

	newFilename := filepath.Join(dstPath, newPath + ".go")

	log.Println(newFilename)
	ioWriter, err := os.Create(newFilename)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(newFilename), os.ModeDir|os.ModePerm)
			if err == nil {
				ioWriter, err = os.Create(newFilename)
			}

		}
	}
	defer ioWriter.Close()

	ioReader, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer ioReader.Close()

	obs.Parse(ioReader, ioWriter)

}
func TestObserver_Parse2(t *testing.T) {

	err := imports.FastWalk(srcPath, Workfunc)
	log.Println(err)
	wrgroup.Wait()
}