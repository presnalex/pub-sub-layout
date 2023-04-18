package storage

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	dbwrapper "github.com/presnalex/go-micro/v3/database/wrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnimalAdd(t *testing.T) {

	// create mock objects
	var clnt ClientMock
	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error creating mock database: %s", err)
	}
	defer db.Close()
	dbx := sqlx.NewDb(db, "postgres")
	dbwrapperMock := dbwrapper.NewWrapper(dbx)

	strg := NewStorage(dbwrapperMock, &clnt, Topics{AnimalAdd: "animaladd", AnimalAddRs: "animaladdrs"})

	tests := []struct {
		name             string
		testCaseFunction func() error
		expectErr        bool
		msgExpect        ResponseMsg
	}{
		{
			name: "Ok",
			testCaseFunction: func() error {
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec("insert into animal_store (animal_id, animal, price) values ($1, $2, $3)").WithArgs(1, "Zebra", 5000).WillReturnResult(sqlmock.NewResult(1, 1))
				sqlMock.ExpectCommit()
				clnt = NewClientMock()
				clnt.On("Publish",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("storage.ClientMessageMock"),
					mock.AnythingOfType("client.PublishOption"),
				).Return(nil)

				msg := []byte(`{"animalId":1, "animal":"Zebra", "price":5000}`)

				err := strg.AnimalAdd(NewContext(), msg)
				return err

			},
			msgExpect: ResponseMsg{MessageId: 1, Status: "Ok"},
		},
		{
			name: "Error empty response",
			testCaseFunction: func() error {
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec("insert into animal_store (animal_id, animal, price) values ($1, $2, $3)").WithArgs(1, "Zebra", 5000).WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectRollback()
				clnt = NewClientMock()
				clnt.On("Publish",
					mock.AnythingOfType("*context.valueCtx"),
					mock.AnythingOfType("storage.ClientMessageMock"),
					mock.AnythingOfType("client.PublishOption"),
				).Return(nil)

				msg := []byte(`{"animalId":1, "animal":"Zebra", "price":5000}`)

				err := strg.AnimalAdd(NewContext(), msg)
				return err

			},
			msgExpect: ResponseMsg{MessageId: 1, Status: "Error", ErrorDetails: "empty response from db"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultErr := tt.testCaseFunction()

			if tt.expectErr {
				assert.Error(t, resultErr)

			} else {
				assert.NoError(t, resultErr)
				assert.Equal(t, tt.msgExpect, *(*clnt.Msg).(*ResponseMsg))
			}
		})
	}

}
