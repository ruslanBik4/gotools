// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package contract

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/ruslanBik4/httpgo/logs"
)

type fncFilterItem func(i *QItem, condition string) bool

type OptFilterItem int8

const (
	FilterId OptFilterItem = iota
	FilterClient
	FilterTxName
)

var fncFilter = map[OptFilterItem] fncFilterItem {
	FilterId: func(i *QItem, condition string) bool {
		return switchCondition(condition, fmt.Sprintf("%8.8d",  i.Id) )
	},
	FilterClient: func(i *QItem, condition string) bool {
		id, err := i.Ctx.Ci.GetID()
		if err != nil {
			logs.ErrorLog( errors.Wrap(err, "filter client"), i)
			return false
		}

		return switchCondition(condition, id )
	},
	FilterTxName: func(i *QItem, condition string) bool {
		return switchCondition(condition, i.Ctx.TxName)
	},
}

var regOper = regexp.MustCompile(`([<>=]\s*([\w\d]+))+`)

func switchCondition(condition string, arg string) bool {
	switch 	list := regOper.FindStringSubmatch(condition); {
	// one operation (less or more)
	case len(list) == 2 && list[1] != "=":
		if list[1] == ">" {
			return arg > list[2]
		}

		return arg < list[2]
	// 	between
	case len(list) == 4:
		return (list[1] == ">" && arg > list[2] && list[3] == "<" && arg < list[4]) ||
				(list[1] == "<" && arg < list[2] && list[3] == ">" && arg > list[4])
	// 	equal
	default:
		return condition == arg
	}
}