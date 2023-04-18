package handler

import (
	"testing"

	raw "github.com/presnalex/codec-bytes"
	"github.com/presnalex/pub-sub-layout/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscribe(t *testing.T) {
	var err error
	//create mock objects
	strg := NewStorageMock()
	//Ok
	msgOk := `{something}`
	strg.On("AnimalAdd",
		mock.AnythingOfType("*context.valueCtx"),
		[]byte(msgOk),
	).Return(nil)

	hdl := NewHandler(strg)
	if err != nil {
		t.Fatal(err)
	}

	//table with tests
	tests := []struct {
		name             string       //test case name
		testCaseFunction func() error //behaviour function that involves a test scenario
		expectErr        bool         //error expectation flag
	}{
		{
			name: "Ok",
			testCaseFunction: func() error {
				return hdl.Subscribe(storage.NewContext(), &raw.Frame{Data: []byte(msgOk)})
			},
		},
	}
	//iterate our test cases and compare results
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = tt.testCaseFunction()
			assert.NoError(t, err)
		})
	}
}
