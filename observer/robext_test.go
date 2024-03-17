// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const srcPath = "/Users/ruslan/work/RobExt/pas/"
const dstPath = "/Users/ruslan/work/src/bitbucket.org/goext/"

var (
	wrgroup sync.WaitGroup
	dictPas *Dictionary
	enc     *Encoder
)

func Workfunc(path string, typeFile os.FileMode) error {
	if typeFile.IsDir() {
		return nil
	}
	go doConvert(path)

	return nil
}
func doConvert(path string) {

	ext := filepath.Ext(path)
	if ext != ".pas" {
		return
	}

	newPath := strings.TrimSuffix(strings.TrimPrefix(path, srcPath), ext)

	wrgroup.Add(1)
	defer wrgroup.Done()

	listDir := strings.Split(filepath.Dir(newPath), "/")
	log.Println(listDir)
	newFilename := filepath.Join(dstPath, newPath+".go")

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

	//insert package name
	if len(listDir) > 0 {
		if _, err := ioWriter.WriteString("package " + strings.TrimSpace(listDir[len(listDir)-1]) + "\r\n"); err != nil {
			log.Println(err)
		}
	}
	obs := NewObserver(enc, dictPas)
	obs.Parse(ioReader, ioWriter)

}
func TestObserver_Parse2(t *testing.T) {

	enc = NewEncoder("win1251")
	dictPas = NewDictionary("dict/pas.dct")

	//err := imports.FastWalk(srcPath, Workfunc)
	//log.Println(err)
	wrgroup.Wait()
}
