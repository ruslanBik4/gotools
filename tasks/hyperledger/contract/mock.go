// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package contract

import (
	"crypto/x509"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
)

const (
	txId = "some ID"
)
var (
	stub = &shimtest.MockStub{
		TxID: txId,
	}
	ci1 = &mockClientIdentity{"1"}

)

type mockClientIdentity struct {
	id string
}

func (mci *mockClientIdentity) GetID() (string, error) {
	return mci.id, nil
}

func (mci *mockClientIdentity) GetMSPID() (string, error) {
	return "", nil
}

func (mci *mockClientIdentity) GetAttributeValue(string) (string, bool, error) {
	return "", false, nil
}

func (mci *mockClientIdentity) AssertAttributeValue(string, string) error {
	return nil
}

func (mci *mockClientIdentity) GetX509Certificate() (*x509.Certificate, error) {
	return nil, nil
}

func getMockCtx(t *testing.T, args ... string) *contractapi.TransactionContext {
	ctx := &contractapi.TransactionContext{}
	ctx.SetStub(stub)
	assert.Equal(t, stub, ctx.GetStub(), "should have returned same stub as set")

	ci := ci1
	if len(args) > 0 {
		ci = &mockClientIdentity{args[0]}
	}

	ctx.SetClientIdentity(ci)
	assert.Equal(t, ci1, ctx.GetClientIdentity(), "should have set the same client identity as passed")

	return ctx
}

// getMockCtxWithCC create ctx with mockStub & clientID/txName optional
func getMockCtxWithCC(t *testing.T, c contractapi.ContractInterface, args ... string) *contractapi.TransactionContext {
	ctx := &contractapi.TransactionContext{}
	cc, err := contractapi.NewChaincode(c)
	assert.Nil(t, err)

	stub := shimtest.NewMockStub("smartContractTest", cc)

	// set from args if present
	ci, txName := ci1, txId
	if len(args) > 0 {
		ci = &mockClientIdentity{args[0]}
		if len(args) > 1 {
			txName = args[1]
		}
	}

	stub.MockTransactionStart(txName)
cc.Start()
	ctx.SetStub(stub)
	assert.Equal(t, stub, ctx.GetStub(), "should have returned same stub as set")

	ctx.SetClientIdentity(ci)
	assert.Equal(t, ci, ctx.GetClientIdentity(), "should have set the same client identity as passed")

	return ctx
}

