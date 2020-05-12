// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package contract

type fncLessItem func(l, r *QItem) bool

type OptOrderItem int8

const (
	OrderId OptOrderItem = iota
	OrderClient
	OrderTxName
)

var fncOrder = map[OptOrderItem] fncLessItem {
	OrderId: func(l, r *QItem) bool {
		return l.Id < r.Id
	},
	OrderClient: func(l, r *QItem) bool {
		return l.Ctx.Ci.id < r.Ctx.Ci.id
	},
	OrderTxName: func(l, r *QItem) bool {
		return l.Ctx.TxName < r.Ctx.TxName
	},
}
