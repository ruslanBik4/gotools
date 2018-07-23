// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"golang.org/x/tools/imports"
	"os"
	"strings"
	"path/filepath"
	"testing"
	"fmt"
	"sync"
)

const testJavaDir = "/Users/ruslan/java/src/github.com/tech-bureau/nem2-sdk-java/src"
const dstJavaPath = "/Users/ruslan/work/src/github.com/482solutions/proximax"
var wrgoupJava  sync.WaitGroup

func TestObserver_Parse_Java(t *testing.T) {


	fmt.Print("start")
	enc = NewEncoder("win1251")
	dictPas = NewDictionary("dict/java.dct")

	wrgoupJava.Add(1)
	err := imports.FastWalk(testJavaDir, observeJavaDir)
	if err != nil {
		t.Error(err)
		
	}
	wrgroup.Wait()
}

func observeJavaDir(path string, typeFile os.FileMode) error {
		if typeFile.IsDir() {
			return nil
		}
		wrgoupJava.Add(1)
		doConvertJava(path)

		return nil
}
func doConvertJava(path string) {
	defer wrgoupJava.Done()
	ext := filepath.Ext(path)
	if ext != ".java" {
		return
	}


	newPath := strings.TrimSuffix(strings.TrimPrefix(path, testJavaDir), ext)
	
	listDir := strings.Split(filepath.Dir(newPath), "/")
	fmt.Println(listDir)
	newFilename := filepath.Join(dstJavaPath, newPath+".go")

	fmt.Println(newFilename)
	ioWriter, err := os.Create(newFilename)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(newFilename), os.ModeDir|os.ModePerm)
			if err == nil {
				ioWriter, err = os.Create(newFilename)
			}
		}
		if err != nil {
			fmt.Println(err)
			return 
		}
	}
	defer ioWriter.Close()

	ioReader, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ioReader.Close()

	obs := NewObserver(enc, dictPas)
	obs.Parse(ioReader, ioWriter)
}
