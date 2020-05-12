// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package contract

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

func TestQueue_AddItem(t *testing.T) {
	q := NewQueue(0)
	ctx := getMockCtxWithCC(t, q)

	tests := []struct {
		name    string
		fields  Queue
		args    contractapi.TransactionContextInterface
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"First",
			Queue{0},
			ctx,
			false,
		},
		{
			"Second",
			Queue{0},
			getMockCtxWithCC(t, q, "2", ""),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := q.Push(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueue_Filter(t *testing.T) {
	ctx := getMockCtxWithCC(t, &Queue{})

	tests := []struct {
		idStartSeq uint64
		name string
		items []contractapi.TransactionContextInterface
		fnc  OptFilterItem
		cnd string
	}{
		// TODO: Add test cases.
		{
			0,
			"first",
			[]contractapi.TransactionContextInterface{
				ctx,
				ctx,
				ctx,
			},
			FilterId,
			"> 1 < 2<",
		},
		{
			1,
			"second",
			[]contractapi.TransactionContextInterface{
				getMockCtxWithCC(t, &Queue{}),
				getMockCtxWithCC(t, &Queue{}, "filtered"),
				getMockCtxWithCC(t, &Queue{}, " filtered"),
			},
			FilterClient,
			 "filtered",
		},
		{
			1,
			"third",
			[]contractapi.TransactionContextInterface{
				getMockCtxWithCC(t, &Queue{}),
				getMockCtxWithCC(t, &Queue{}, "filtered", "some tx"),
				getMockCtxWithCC(t, &Queue{}, " filtered"),
			},
			FilterTxName,
			 "some tx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue(tt.idStartSeq)
			for _, ctx := range tt.items {
				assert.Nil(t, q.Push(ctx))
			}

			assert.Nil(t, q.Filter(ctx, tt.fnc, tt.cnd) )
		})
	}
}

func TestQueue_GetInfo(t *testing.T) {
	q := NewQueue(0)
	assert.IsType(t, metadata.InfoMetadata{},  q.GetInfo(), "GetInfo() = %v, want %v")
}

func TestQueue_GetTransactionContextHandler(t *testing.T) {
	q := NewQueue(0)
	assert.Implements(t, (*contractapi.SettableTransactionContextInterface)(nil), q.GetTransactionContextHandler(),"GetTransactionContextHandler() = %v, want %v")
}

func TestQueue_InitItems(t *testing.T) {
	tests := []struct {
		name    string
		fields  Queue
		args    contractapi.TransactionContextInterface
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"first",
			Queue{0},
			getMockCtxWithCC(t, &Queue{}),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				idStartSeq: tt.fields.idStartSeq,
			}
			if err := q.InitItems(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("InitItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueue_Len(t *testing.T) {

	tests := []struct {
		name    string
		fields  Queue
		args    uint64
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"first",
			Queue{0},
			1,
			1,
			false,
		},
		{
			"second",
			Queue{999},
			10,
			10,
			false,
		},
		{
			"third",
			Queue{1},
			3,
			3,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				idStartSeq: tt.fields.idStartSeq,
			}
			ctx := getMockCtxWithCC(t, q)
			stub:= ctx.GetStub()
			for i := uint64(0); i < tt.args; i++ {
				assert.Nil(t, q.Push(ctx))
			}

			got, err := q.maxID(stub)
			assert.Equal(t, tt.wantErr, err != nil,"maxID() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.want, got ,"maxID() got = %v, want %v", got, tt.want)
		})
	}
}

func TestQueue_Reorder(t *testing.T) {

	tests := []struct {
		name    string
		fields  Queue
		args    OptOrderItem
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"first",
			Queue{0},
			OrderId,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				idStartSeq: tt.fields.idStartSeq,
			}
			ctx := getMockCtxWithCC(t, q)
			if err := q.Reorder(ctx, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Reorder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func Test_newQItem(t *testing.T) {
	ctx := getMockCtx(t)

	tests := []struct {
		name string
		args uint64
		want *QItem
	}{
		// TODO: Add test cases.
		{
			"first",
			0,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQItem(tt.args, ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newQItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newQItemFromByte(t *testing.T) {
	ctx := getMockCtx(t)
	wantCtx := fmt.Sprintf(`{"id":"%s","tx":"%s"}
`, ci1.id, stub.TxID)

	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		wantI   QItem
		wantErr bool
	}{
		// TODO: Add test cases.
		{"first",
			args{append([]byte{0, 0, 0, 0, 0, 0, 0, 0}, wantCtx...),
			},
			QItem{Id: 0, Ctx: NewItemContext(ctx)},
			false,
		},
		{"second",
			args{	append([]byte{0, 0, 0, 0, 0, 0, 3, 231}, wantCtx...),
			},
			QItem{Id: 999, Ctx: NewItemContext(ctx)},
			false,
		},
		{"third",
			args{append([]byte{0, 0, 0, 0, 0, 0, 0, 13}, wantCtx...),
			},
			QItem{Id: 13, Ctx: NewItemContext(ctx)},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotI, err := newQItemFromByte(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("newQItemFromByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, *gotI, tt.wantI,"newQItemFromByte() gotI = %v, want %v", gotI, tt.wantI)
		})
	}
}

func Test_qItem_Marshal(t *testing.T) {
	type caseFields struct {
		name    string
		fields  QItem
		want    []byte
		wantErr bool
	}

	ctx := getMockCtx(t)

	wantCtx := fmt.Sprintf(`{"id":"%s","tx":"%s"}
`, ci1.id, stub.TxID)

	var tests = []caseFields{
		// TODO: Add test cases.
		{"first",
			QItem{Id: 0, Ctx: NewItemContext(ctx)},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0},
			false,
		},
		{"second",
			QItem{Id: 999, Ctx: NewItemContext(ctx)},
			[]byte{0, 0, 0, 0, 0, 0, 3, 231},
			false,
		},
		{"third",
			QItem{Id: 13, Ctx: NewItemContext(ctx)},
			[]byte{0, 0, 0, 0, 0, 0, 0, 13},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &QItem{
				Id:  tt.fields.Id,
				Ctx: tt.fields.Ctx,
			}
			got, err := i.MarshalJSON()
			if (err != nil) != tt.wantErr {
				assert.Nil(t, err, "MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, got[:8], tt.want, "MarshalJSON() got = %v, want")
			assert.Equal(t, string(got[8:]), wantCtx, "MarshalJSON() got = %s, want %v", got[:8], wantCtx)
		})
	}
}

func Test_qItem_Put(t *testing.T) {

	tests := []struct {
		name    string
		fields  *QItem
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"first",
			nil,
			false,
		},
	}

	q := NewQueue(0)
	ctx := getMockCtxWithCC(t, q)
	stub := ctx.GetStub()

	for _, tt := range tests {
		id := uint64(0)
		t.Run(tt.name, func(t *testing.T) {
			id++
			i := NewQItem(id, ctx)

			if err := i.Put(stub); (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}