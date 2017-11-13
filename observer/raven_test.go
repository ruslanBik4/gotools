// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
var wrgroup1  sync.WaitGroup
func Workfunc1(path string, typeFile os.FileMode) error{
	var ignoreFiles = []string {"setup", "__init__", }
	if typeFile.IsDir() {
		return nil
	}
	for _, name := range ignoreFiles {
		if filepath.Base(path) == name + ".py" {
			return nil
		}
	}
	go doConvert1(path)

	return nil
}
func doConvert1(path string) {
	const srcPath = "/Users/ruslan/GoglandProjects/RavenDB-Python-Client/pyravendb/"
	const dstPath = "./temp/"
	var dict *Dictionary
	ext := filepath.Ext(path)
	newPath := strings.TrimSuffix( strings.TrimPrefix( path, srcPath ), ext)

	wrgroup1.Add(1)
	defer wrgroup1.Done()

	listDir := strings.Split(filepath.Dir(newPath), "/")

	if ext != ".py" {
		return
	}
		dict = NewDictionary("dict/py.dct")
	var packName []byte
	if len(listDir) > 0 {
		packName = []byte(listDir[len(listDir)-1])
		//log.Println(string(packName))
		dict.genRules[regexp.MustCompile("{package_name}")] = packName
	}
	enc  := NewEncoder("win1251")
	obs := NewObserver(enc, dict, NewDictionary("py_add.dct"))

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

	//insert package name
	if len(listDir) > 0 {
		if _, err := ioWriter.WriteString("package "); err != nil {
			log.Println(err)
		}
		obs.write(ioWriter, packName)
		if _, err := ioWriter.WriteString("\n"); err != nil {
			log.Println(err)
		}
	}
	obs.write(ioWriter, []byte(`import (
	"errors"
	"fmt"
	"net/http"
	"github.com/ravendb-go-client/store"
	SrvNodes "github.com/ravendb-go-client/http/server_nodes"
	"strconv"
)`))
	ioReader, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer ioReader.Close()

	obs.Parse(ioReader, ioWriter)

	obs.write(ioWriter, []byte("\n}"))

}
func TestObserver_Parse_Raven(t *testing.T) {
	const srcPath = "/Users/ruslan/GoglandProjects/RavenDB-Python-Client/pyravendb/"

	err := imports.FastWalk(srcPath, Workfunc1)
	log.Println(err)
	wrgroup1.Wait()
}