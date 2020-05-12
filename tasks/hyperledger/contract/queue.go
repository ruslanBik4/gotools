// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package contract

import (
	"bytes"
	"encoding/binary"
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"

	"github.com/pkg/errors"

	"github.com/ruslanBik4/httpgo/logs"
)

// ● Every queue item should have a unique ID.
// ● Queue should provide an API to attach extra context to the item
// ● Queue should provide an API for items reordering
// ● Queue should provide an API for items filtering and applying above mentioned operations on the filtered batch

const (
	itemsKey       = "qItem"
	collectionName = "itemsQueue"
	seqKeyName     = "itemsSeq"
	indKeyName     = "itemsInd"
	fltKeyName     = "itemsFilter"
)

type Queue struct {
	idStartSeq uint64
}

func NewQueue(idSeq uint64) *Queue {
	return &Queue{idStartSeq: idSeq}
}

func (q *Queue) GetInfo() metadata.InfoMetadata {
	return metadata.InfoMetadata{
		Description: "Hyperledger Fabric (c) smart contract that implements a simple queue",
		Title:       "Queue of items of context",
		Contact:     &metadata.ContactMetadata{
			Name:  "Ruslan Bikchentaev",
			URL:   "https://github.com/ruslanBik4/gotools",
			Email: "bik4ruslan@gmail.com",
		},
		License:     &metadata.LicenseMetadata{
			Name: "Apache License Version 2.0",
			URL:  "https://opensource.org/licenses/Apache-2.0",
		},
		Version:     "0.0.1",
	}
}

func (q *Queue) GetUnknownTransaction() interface{} {
	return nil
}

func (q *Queue) GetBeforeTransaction() interface{} {
	d := logs.SetDebug(true)
	defer logs.SetDebug(d)

	return func() {
		logs.DebugLog("begin tx")
	}
}

func (q *Queue) GetAfterTransaction() interface{} {
	d := logs.SetDebug(true)
	defer logs.SetDebug(d)

	return func() {
		logs.DebugLog("end tx")
	}
}

func (q *Queue) GetName() string {
	return "queueCtxItems"
}

func (q *Queue) GetTransactionContextHandler() contractapi.SettableTransactionContextInterface {
	return &contractapi.TransactionContext{}
}

func (q *Queue) GetIgnoredFunctions() []string {
	return []string{"GetAllItems"}
}

func (q *Queue) InitItems(ctx contractapi.TransactionContextInterface) error {
	return q.putIdSeq(ctx.GetStub(), q.idStartSeq)
}

func (q *Queue) putIdSeq(stub shim.ChaincodeStubInterface, idSeq uint64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, idSeq)

	return stub.PutState(seqKeyName, b)
}

func (q *Queue) maxID(stub shim.ChaincodeStubInterface) (uint64, error) {
	b, err := stub.GetState(seqKeyName)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get length of items")
	}

	if len(b) < 8 {
		return 0, errors.Wrapf(err, "wrong length of bytes %s", b)
	}

	return binary.BigEndian.Uint64(b), nil
}

func (q *Queue) Push(ctx contractapi.TransactionContextInterface) error {

	stub := ctx.GetStub()
	if stub.GetTxID() == "" {
		return errors.Errorf("ctx hasn't tx name %+v", stub)
	}

	idSeq, err := q.maxID(stub)
	if err != nil {
		return err
	}
	idSeq++
	err = q.putIdSeq(stub, idSeq)
	if err != nil {
		return errors.Wrapf(err, "during put new idStartSeq %d", idSeq)
	}

	return NewQItem(idSeq, ctx).Put(stub)
}

func (q *Queue) Pop(ctx contractapi.TransactionContextInterface) (*QItem, error) {
	// get filter
	item, err := q.readFirstId(ctx, fltKeyName)
	if err != nil {
		return nil, errors.Wrap(err, "get filter failed")
	}

	if item != nil {
		return item, nil
	}
	// get order
	return q.readFirstId(ctx, fltKeyName)
// 	todoL implement read according to id
}

func (q *Queue) readFirstId(ctx contractapi.TransactionContextInterface, key string) (*QItem, error) {
	b, err := stub.GetPrivateData(collectionName, key)
	if err != nil {
		return nil, errors.Wrap(err, "get filter failed")
	}

	if b != nil {
		r := bytes.NewReader(b)
		var id uint64
		err = binary.Read(r, binary.BigEndian, &id)
		if err != nil {
			return nil, errors.Wrap(err, "read failed")
		}

		return q.GetItem(ctx, id)
	}

	return nil, nil
}

func (q *Queue) Reorder(ctx contractapi.TransactionContextInterface, fnc OptOrderItem) error {

	stub := ctx.GetStub()
	items, err := q.GetAllItems(ctx)
	if err != nil {
		return err
	}

	s := &sortItems{
		items: items,
		fnc:   fncOrder[fnc],
	}

	sort.Sort(s)

	buf := &bytes.Buffer{}
	filter := q.GetLastFilter(ctx)
	for _, item := range s.items {
		if len(items) > 0 {
			ok := false
			for _, id := range filter {
				ok = id == item.Id
				if ok {
					break
				}
			}

			if !ok {
				continue
			}
		}
		err = binary.Write(buf, binary.BigEndian, item.Id)
		if err != nil {
			return errors.Wrap(err, "index writing failed")
		}
	}

	return stub.PutPrivateData(collectionName, indKeyName, buf.Bytes())
}

func (q *Queue) GetAllItems(ctx contractapi.TransactionContextInterface) ([]*QItem, error) {
	stub := ctx.GetStub()
	idSeq, err := q.maxID(stub)
	if err != nil {
		return nil, err
	}

	sKey, eKey := itemsKey+"0", itemsKey+strconv.Itoa(int(idSeq))
	iter, err := stub.GetStateByRange(sKey, eKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get iteration of items")
	}

	defer func() {
		err := iter.Close()
		if err != nil {
			logs.ErrorLog(errors.Wrap(err, "Reorder can't close iterator"))
		}
	}()

	items := make([]*QItem, 0)
	for iter.HasNext() {
		resp, err := iter.Next()
		if err != nil {
			return nil, errors.Wrap(err, "during iterator.Next")
		}

		item, err := newQItemFromByte(resp.Value)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (q *Queue) GetItem(ctx contractapi.TransactionContextInterface, id uint64) (*QItem, error) {
	b, err := ctx.GetStub().GetState(itemsKey + strconv.Itoa(int(id)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed during get item #%d", id)
	}

	i, err := newQItemFromByte(b)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (q *Queue) AddCtx(ctx contractapi.TransactionContextInterface, id uint64, deadline *time.Time) error {
	i, err := q.GetItem(ctx, id)
	if err != nil {
		return err
	}

	if clientIdCur, err := i.Ctx.Ci.GetID(); err != nil {
		return errors.Wrapf(err, "failed get client Id #%d", id)
	} else if clientIdNew, err := ctx.GetClientIdentity().GetID(); err != nil {
		return errors.Wrapf(err, "failed get client Id #%d", id)
	} else if clientIdCur != clientIdNew {
		return errors.Wrapf(err, "new context has wrong client Id #%d", id)
	}

	item := NewQItem(id, ctx)
	if deadline != nil {
		item.Ctx.Deadline = deadline
	}

	return item.Put(ctx.GetStub())
}

func (q *Queue) Apply(ctx contractapi.TransactionContextInterface, cmd string) error {
	// todo: imlement
	return nil
}

func (q *Queue) Filter(ctx contractapi.TransactionContextInterface, fnc OptFilterItem, condition string) error {
	stub := ctx.GetStub()
	items, err := q.GetAllItems(ctx)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	for _, item := range items {
		if fncFilter[fnc](item, condition) {
			err = binary.Write(buf, binary.BigEndian, item.Id)
			if err != nil {
				return errors.Wrap(err, "index writing failed")
			}
		}
	}

	return stub.PutPrivateData(collectionName, fltKeyName, buf.Bytes())
}

func (q *Queue) GetLastFilter(ctx contractapi.TransactionContextInterface) []uint64 {
	stub := ctx.GetStub()
	b, err := stub.GetPrivateData(collectionName, fltKeyName)
	if err != nil {
		logs.ErrorLog( errors.Wrap(err, "get filter failed") )
		return nil
	}

	if b != nil {
		items := make([]uint64, 0)
		r := bytes.NewReader(b)
		for {
			var id uint64
			err = binary.Read(r, binary.BigEndian, &id)
			if err == io.EOF {
				break
			} else if err != nil {
				logs.ErrorLog( errors.Wrap(err, "read failed") )
				return nil
			}

			items = append(items, id)
		}

		return items
	}

	return nil
}