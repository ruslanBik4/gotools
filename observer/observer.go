// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"os"
	"bytes"
	"log"
)

type Observer struct {
	enc *Encoder
	dicts []*Dictionary
	names map[string] string
}
func NewObserver(enc *Encoder, dict ... *Dictionary) *Observer {
	return &Observer{enc: enc, dicts: dict, names: make(map[string]string,0)}
}
func (o *Observer) Parse(ioReader, ioWriter *os.File) {
	stat, err := ioReader.Stat()
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	b := make([]byte, stat.Size())
	n, err := ioReader.Read(b)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	log.Print(n)
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

			o.write(ioWriter, o.doReplacers(line) )
		}
		ioWriter.Write([]byte("\n"))
	}
}
func (o *Observer) doReplacers(line []byte) []byte {
	for _, dict := range o.dicts {
		for _, value := range dict.replacers {

			if value.src.Match(line) {
				subExp := value.src.SubexpNames()
				for i, group := range subExp {
					if group > "" {
						o.names["{" + group + "}"] = string( value.src.FindSubmatch(line)[i] )
					}
				}
				return value.src.ReplaceAll(line, value.repl)
			}
		}
	}

	return line
}
func (o *Observer) write(ioWriter *os.File, line []byte) {
	if string(line) == "" {
		return
	}
	for _, dict := range o.dicts {
		for key, value := range dict.genRules {
			line = key.ReplaceAll(line, value)
		}
	}
	for key, value := range o.names {
		line = bytes.Replace(line, []byte(key), []byte(value), -1)
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
