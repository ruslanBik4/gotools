// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

type Observer struct {
	enc   *Encoder
	dicts []*Dictionary
	names map[string]string
	ifs   map[string]bool
	lock  sync.RWMutex
}

func NewObserver(enc *Encoder, dict ...*Dictionary) *Observer {
	return &Observer{
		enc:   enc,
		dicts: dict,
		names: map[string]string{"{indent}": ""},
		ifs:   map[string]bool{},
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
		if len(line) == 0 {
			continue
		}
		if isComment(line) {
			o.write(ioWriter, line)
		} else {
			line = o.putNamesInLine(line)
			o.write(ioWriter, o.doReplacers(line))
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
					case "ClassName":
						className := string(value.src.FindSubmatch(line)[i])
						o.writeName("{indent}", "} /* "+className+" */ \n")
						o.writeName("{"+group+"}", className)
					case "indent":
						o.writeName("{"+group+"}", "")
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
func (o *Observer) writeIfs(key string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if strings.HasPrefix(key, "Not") {
		o.ifs[strings.TrimPrefix(key, "Not")] = false
	} else {
		o.ifs[key] = true
	}
}

var containtIfs = regexp.MustCompile(`\{\{\w*\}\}`)

func (o *Observer) validIfs(line []byte) ([]byte, bool) {
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

func (o *Observer) writeName(key, value string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.names[key] = value
}
func (o *Observer) putNamesInLine(line []byte) []byte {
	o.lock.Lock()
	defer o.lock.Unlock()

	for key, value := range o.names {
		switch key {
		case "indent", "indentNew", "":
			continue
		}
		if value > "" {
			reg, err := regexp.Compile(`\b` + value + `\b`)
			if err != nil {
				fmt.Println(err)

			} else if reg.Match(line) {
				line = reg.ReplaceAll(line, []byte(" "+key+" "))
				//fmt.Println(value)
			}
		}
	}

	return line
}
func (o *Observer) insertNames(line []byte) []byte {
	o.lock.Lock()
	defer o.lock.Unlock()

	for key, value := range o.names {
		line = bytes.Replace(line, []byte(key), []byte(value), -1)
	}

	return line
}

var regForgroup = regexp.MustCompile(`forgroup\(([^)]*)\)`)
var regPublic = regexp.MustCompile(`([\S\s]*)public`)

func (o *Observer) write(ioWriter *os.File, line []byte) {
	if string(line) == "" {
		return
	}
	if regPublic.Match(line) {
		posSpace := 0
		for i, val := range line {
			if val != ' ' {
				posSpace = i
				break
			}
		}
		lineForChange := line[posSpace:]

		if bytes.HasPrefix(lineForChange, []byte("type")) {
			offset := len([]byte("type"))
			for i, val := range bytes.TrimPrefix(lineForChange, []byte("type")) {
				if val != ' ' {
					lineForChange[i+offset] = byte(unicode.ToUpper(rune(val)))
					break
				}
			}
		} else if bytes.HasPrefix(lineForChange, []byte("interface")) {
			offset := len([]byte("interface"))
			for i, val := range bytes.TrimPrefix(lineForChange, []byte("interface")) {
				if val != ' ' {
					lineForChange[i+offset] = byte(unicode.ToUpper(rune(val)))
					break
				}
			}
		} else if bytes.HasPrefix(lineForChange, []byte("func")) {
			isSkip := false
			offset := len([]byte("func"))
			for i, val := range bytes.TrimPrefix(lineForChange, []byte("func")) {
				switch val {
				case ' ':
				case '(':
					isSkip = true
				case ')':
					isSkip = false
				default:
					if isSkip {
						break
					}
					lineForChange[i+offset] = byte(unicode.ToUpper(rune(val)))
					offset = -1
					break
				}
				if offset < 0 {
					break
				}
			}
		} else {
			for i, val := range lineForChange {
				if val != ' ' {
					lineForChange[i] = byte(unicode.ToUpper(rune(val)))
					break
				}
			}
		}
		if posSpace > 0 {
			line = append(line[:posSpace-1], lineForChange...)
		} else {
			line = lineForChange
		}
	}
	for _, dict := range o.dicts {
		dict.LockIteration(func(key *regexp.Regexp, value []byte) bool {
			repl, valid := o.validIfs(value)
			repl = o.insertNames(repl)
			if bytes.Equal(repl, []byte("toLower")) && key.Match(line) {

				line = bytes.ToLower(line)
			} else if regForgroup.Match(repl) {
				replVal := regForgroup.Find(repl)
				for _, val := range key.FindAllSubmatch(line, -1) {
					for _, val := range val {
						fmt.Println(string(val))

					}
				}
				line = key.ReplaceAll(line, regForgroup.ReplaceAll(repl, replVal))
			} else if valid {
				line = key.ReplaceAll(line, repl)
			}
			return true
		})
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
	return bytes.HasPrefix(line, []byte("//")) || bytes.HasPrefix(line, []byte("* ")) ||
		(bytes.HasPrefix(line, []byte("/*")) && bytes.HasSuffix(line, []byte("*/")))
}
