// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package contract

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

type sortItems struct {
	items []*QItem
	fnc   fncLessItem
}

func (s *sortItems) Len() int {
	return len(s.items)
}

func (s *sortItems) Less(i, j int) bool {
	return s.fnc(s.items[i], s.items[j])
}

func (s *sortItems) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

type QItem struct {
	Id  uint64		`json:"id"`
	Ctx ItemContext `json:"ctx"`
}

func NewQItem(id uint64, ctx contractapi.TransactionContextInterface) *QItem {
	c := &contractapi.TransactionContext{}
	c.SetStub(ctx.GetStub())
	c.SetClientIdentity(ctx.GetClientIdentity())

	return &QItem{id, NewItemContext(c)}
}

func newQItemFromByte(b []byte) (i *QItem, err error) {

	i = &QItem{Ctx: NewItemContext(nil)}

	err = i.UnmarshalJSON(b)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal data "+string(b))
	}

	return
}

func (i *QItem) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, i.Id)
	if err != nil {
		return nil, errors.Wrap(err, "index of item writing failed")
	}

	enc := json.NewEncoder(buf)
	err = enc.Encode(i.Ctx)
	if err != nil {
		return nil, errors.Wrap(err, "items marshaling failed")
	}

	return buf.Bytes(), nil
}

func (i *QItem) UnmarshalJSON(b []byte) error {
	r := bytes.NewReader(b)
	err := binary.Read(r, binary.BigEndian, &i.Id)
	if err == nil {
		dec := json.NewDecoder(r)
		err = dec.Decode(&i.Ctx)
	}

	return err
}

func (i *QItem) Put(stub shim.ChaincodeStubInterface) error {
	b, err := i.MarshalJSON()
	if err != nil {
		return err
	}

	return stub.PutState(itemsKey+strconv.Itoa(int(i.Id)), b)
}

type ItemContext struct {
	Deadline *time.Time
	Ci mockClientIdentity
	TxName string
	Values map[string]interface{}
}

func NewItemContext(ctx *contractapi.TransactionContext) ItemContext {
	if ctx == nil {
		return ItemContext{}
	}

	id, _ := ctx.GetClientIdentity().GetID()

	return ItemContext{
		Ci:       mockClientIdentity{id},
		TxName:   ctx.GetStub().GetTxID(),
		Values:   nil,
	}
}

func (i ItemContext) UnmarshalJSON(b []byte) error {
	dto := struct{
		id string
		tx string
		dl int64
	} {}

	err := json.Unmarshal(b, &dto)
	if err != nil {
		return errors.Wrap(err, "")
	}

	i.Ci = mockClientIdentity{dto.id}
	i.TxName = dto.tx

	if dto.dl > 0 {
		*(i.Deadline) = time.Unix(dto.dl, 0 )
	}

	return nil
}
// MarshalJSON implement part of marshalling itemContext
func (i ItemContext) MarshalJSON() ([]byte, error) {
	id, err := i.Ci.GetID()
	if err != nil {
		return nil, errors.Wrap(err, "get id context failed")
	}

	buf := bytes.NewBufferString(fmt.Sprintf(`{"id":"%s"`, id) )
	_, err = buf.WriteString(fmt.Sprintf(`, "tx":"%s"`, i.TxName ) )
	if err != nil {
		return nil, errors.Wrap(err, "write to buffer failed")
	}

	if i.Deadline != nil {
		_, err = buf.WriteString(fmt.Sprintf(`, "dl":%d`,  i.Deadline.UnixNano() ) )
		if err != nil {
			return nil, errors.Wrap(err, "write to buffer failed")
		}
	}

	_, err = buf.WriteString(`}`)
	if err != nil {
		return nil, errors.Wrap(err, "write to buffer failed")
	}

	return buf.Bytes(), nil
}
