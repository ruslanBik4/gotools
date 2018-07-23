// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"os"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"io/ioutil"
)

type Observer struct {
	enc   *Encoder
	dicts []*Dictionary
	names map[string]string
	ifs   map[string] bool
	lock  sync.RWMutex
}
func NewObserver(enc *Encoder, dict ... *Dictionary) *Observer {
	return &Observer{
		enc:   enc,
		dicts: dict,
		names: map[string]string{"{indent}": "",},
		ifs: map[string]bool{},
	}
}
func (o *Observer) Parse(ioReader, ioWriter *os.File) {
	b, err := ioutil.ReadAll(ioReader)
	if err != nil {
		fmt.Println(err) // panic is used only as an example and is not otherwise recommended.
		return
	}
	b = bytes.Replace(b, []byte("\r\n"), []byte("\n"), -1)
	slBytes := bytes.Split(b, []byte("\n"))

	for _, line := range slBytes {

		// комментарии и пустые строки пропускаем
		if (len(line) == 0) {
			continue
		}
		if isComment(line) {
			o.write(ioWriter, line)
		} else {
			line = o.putNamesInLine(line)
			o.write(ioWriter, o.doReplacers(line) )
		}
		ioWriter.Write([]byte("\n"))
	}
}
func (o *Observer) doReplacers(line []byte) []byte {
	for _, dict := range o.dicts {
		for _, value := range dict.replacers {

			repl, valid := o.validIfs(value.repl)
			if !valid {
				continue
			}

			if value.src.Match(line) {
				repl = o.insertNames(repl)
				subExp := value.src.SubexpNames()
				for i, group := range subExp {
					switch group {
					case "indent":
							o.writeName("{"+group+"}", "}\n")
					case "indentNew":
						o.writeName("{indent}", "    return ref\n}\n")
					case "":
						continue
					default:
						if strings.HasPrefix(group, "is") {
							o.writeIfs(strings.TrimPrefix(group, "is"))
						} else {
							o.writeName("{"+group+"}", string(value.src.FindSubmatch(line)[i]))
						}
					}
				}
				return value.src.ReplaceAll(line, repl)
			}
		}
	}

	return line
}
func (o *Observer) writeIfs(key string)  {
	o.lock.Lock()
	defer o.lock.Unlock()
	if strings.HasPrefix(key, "Not") {
		o.ifs[strings.TrimPrefix(key, "Not")] = false
	} else {
		o.ifs[key] = true
	}
}
var containtIfs = regexp.MustCompile(`\{\{\w*\}\}`)
func (o *Observer) validIfs(line []byte) ([]byte, bool){
	if len(o.ifs) == 0 {
		return line, true
	}
	o.lock.Lock()
	defer o.lock.Unlock()

	for key, value := range o.ifs {
		keyWord := []byte("{{" + key + "}}")
		if bytes.Contains(line, keyWord) {
			return bytes.Replace(line, keyWord, []byte(""), -1), value
		}
	}

	return line, !containtIfs.Match(line)
}

func (o *Observer) writeName(key, value string)  {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.names[key] = value
}
func (o *Observer) putNamesInLine(line []byte) []byte{
	o.lock.Lock()
	defer o.lock.Unlock()

	for key, value := range o.names {
		switch key {
		case "indent", "indentNew", "":
		  continue
		}
		if (value > "") && strings.Contains(string(line), " "+value) {
			line = bytes.Replace(line, []byte(" "+value), []byte(" "+key), -1)
		}
	}

	return line
}
func (o *Observer) insertNames(line []byte) []byte{
	o.lock.Lock()
	defer o.lock.Unlock()

	for key, value := range o.names {
		line = bytes.Replace(line, []byte(key), []byte(value), -1)
	}

	return line
}
var regForgroup = regexp.MustCompile(`forgroup\(([^)]*)\)`)
func (o *Observer) write(ioWriter *os.File, line []byte) {
	if string(line) == "" {
		return
	}
	for _, dict := range o.dicts {
		dict.LockIteration( func (key *regexp.Regexp, value []byte) bool {
			repl, valid := o.validIfs(value)
			repl = o.insertNames(repl)
			if regForgroup.Match(repl) {
				replVal := regForgroup.Find(repl)
				for _, val := range key.FindAllSubmatch(line, -1) {
					for _, val := range val {
						fmt.Println(string(val))
						
					}
				}
				line = key.ReplaceAll(line, regForgroup.ReplaceAll(repl, replVal) )
			} else if valid {
				line = key.ReplaceAll(line, repl)
			}
			return true
		} )
	}
	if o.enc == nil {
		ioWriter.Write(line)
	} else {
		ioWriter.Write(o.enc.Encoding(line))
	}
}
// line is comment
func isComment(line []byte) bool {
	line = bytes.TrimSpace(line)
	return bytes.HasPrefix(line, []byte("//")) ||
		(bytes.HasPrefix(line, []byte("/*")) && bytes.HasSuffix(line, []byte("*/")))
}
