// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"regexp"
	"os"
	"bytes"
	"sync"
)

type replacer struct {
	src *regexp.Regexp
	repl []byte

}
type Dictionary struct {
	replacers [] replacer
	genRules  map[*regexp.Regexp][]byte
	lock      sync.RWMutex
}
func NewDictionary(fileName string) *Dictionary {
	d := &Dictionary{}
	d.genRules = make(map[*regexp.Regexp] []byte, 0)
	d.readDict(fileName)

	return d
}
const (
	modeRules = iota
	modeReplaces
	modeComments
)
func (d *Dictionary) readDict(fileName string) {
	ioReader, err := os.Open(fileName)
	defer ioReader.Close()

	stat, err := ioReader.Stat()
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	b := make([]byte, stat.Size())
	_, err = ioReader.Read(b)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	b = bytes.Replace(b, []byte("\r\n"), []byte("\n"), -1)
	slBytes := bytes.Split(b, []byte("\n"))

	mode := modeReplaces
	for _, line := range slBytes {
		if len(line) == 0 || string(line) == "" {
			continue
		}
		if string(line) == "#general rules" {
			mode = modeRules
			continue
		}
		list := bytes.Split(line, []byte(" :: ") )
		if len(list) < 2 {
			continue
		}

		repl := bytes.Replace(list[1], []byte("\\n"), []byte("\n"), -1)

		repl = bytes.Replace(repl, []byte("\\s"), []byte(" "), -1)

		src := regexp.MustCompile(string(bytes.Trim(list[0], " ")) )

		switch mode {
		case modeRules:
			d.genRules[src] = repl
		case modeReplaces:
			d.replacers = append(d.replacers, replacer{src: src, repl: repl})
		}
	}
}
func (dict *Dictionary) Write(name string, value []byte) {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	dict.genRules[regexp.MustCompile(name)] = value
}
func (dict *Dictionary) LockIteration( funcEval func( *regexp.Regexp, []byte) bool) {
	dict.lock.RLock()
	defer dict.lock.RUnlock()

	for key, value := range dict.genRules {
		if !funcEval(key, value) {
			break
		}
	}
}
func (dict *Dictionary) UnLockIteration() {
	dict.lock.RUnlock()
}