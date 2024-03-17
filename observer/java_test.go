// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

const testJavaDir = "/Users/ruslan/java/src/github.com/tech-bureau/nem2-sdk-java/src"
const dstGoRoot = "/Users/ruslan/work/src/github.com/482solutions/proximax/sdk"

var wrgoupJava sync.WaitGroup

func TestObserver_Parse_Java(t *testing.T) {

	fmt.Print("start")
	enc = NewEncoder("win1251")
	dictPas = NewDictionary("dict/java.dct")

	wrgoupJava.Add(1)
	//todo: chg tet dir
	//err := imports.Walk(testJavaDir, observeJavaDir)
	//if err != nil {
	//	t.Error(err)
	//
	//}
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

var shrinkPaths = []string{
	"main", "java", "io", "nem", "sdk", "io", "nem", "test",
}
var regIsTest = regexp.MustCompile(`(\w*)Test`)
var regShrinkNames = regexp.MustCompile(`([A-Z]){1}([a-z]*)`)

func fileNameConvert(fileName string) string {
	if regIsTest.MatchString(fileName) {
		fileName = regIsTest.ReplaceAllString(fileName, `${1}_test`)
	}
	i := 0
	fileName = regShrinkNames.ReplaceAllStringFunc(fileName, func(z string) string {
		if i > 0 {
			z = "_" + z
		}
		i++
		return strings.ToLower(z)
	})
	return fileName
}
func doConvertJava(path string) {
	defer wrgoupJava.Done()
	ext := filepath.Ext(path)
	if ext != ".java" {
		return
	}

	newPath := strings.TrimSuffix(strings.TrimPrefix(path, testJavaDir), ext)

	listDir := strings.Split(filepath.Dir(newPath), "/")

	fileName := filepath.Base(newPath)

	if fileName == "." {
		return
	}

	newPath = dstGoRoot
	for _, dir := range listDir {
		var isShrink bool
		for _, val := range shrinkPaths {
			isShrink = (dir == val)
			if isShrink {
				break
			}
		}
		if !isShrink {
			newPath = filepath.Join(newPath, dir)
		}
	}

	newFilename := filepath.Join(newPath, fileNameConvert(fileName)+".go")

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
	//insert package name
	if len(listDir) > 0 {
		obs.writeName("{packageName}", strings.TrimSpace(listDir[len(listDir)-1]))
	}
	obs.writeName("{pathRepo}", "github.com/proximax/nem2-go-sdk")
	obs.Parse(ioReader, ioWriter)
}
