// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package observer

import (
	"strings"
	"io/ioutil"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/charmap"
)

type Encoder struct {
	charmap *charmap.Charmap
}
func NewEncoder(codepage string) *Encoder {
	enc := &Encoder{}
	switch codepage {
	case "win1251":
		enc.charmap = charmap.Windows1251
	}

	return enc
}
func (e *Encoder) Encoding(str []byte) []byte{
	sr := strings.NewReader(string(str))
	tr := transform.NewReader(sr, e.charmap.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != err {
		// обработка ошибки
		panic(err)
	}

	return buf // строка в UTF-8
}
