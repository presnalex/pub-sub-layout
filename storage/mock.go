package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/presnalex/go-micro/v3/wrapper/requestid"
	"github.com/stretchr/testify/mock"
	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/metadata"
)

type ClientMock struct {
	mock.Mock
	Msg *interface{}
}

func (m *ClientMock) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	m.Msg = &[]interface{}{msg.Payload()}[0]
	argsCalles := []interface{}{ctx, msg}
	optsI := make([]interface{}, len(opts))
	for i, opt := range opts {
		optsI[i] = opt
	}
	argsCalles = append(argsCalles, optsI...)
	arguments := m.Called(argsCalles...)
	return arguments.Error(0)
}

func (m *ClientMock) NewMessage(topic string, msg interface{}, opts ...client.MessageOption) client.Message {
	return ClientMessageMock{
		msg:   msg,
		topic: topic,
	}
}

func NewClientMock() ClientMock {
	return ClientMock{}
}

type ClientMessageMock struct {
	msg   interface{}
	topic string
}

func (m ClientMessageMock) Topic() string {
	return m.topic
}
func (m ClientMessageMock) Payload() interface{} {
	return m.msg
}
func (m ClientMessageMock) Metadata() metadata.Metadata {
	return nil
}
func (m ClientMessageMock) ContentType() string {
	return fmt.Sprintf("%s", m.msg)
}

func NewContext() context.Context {
	ctx := context.Background()
	uid, err := uuid.NewRandom()
	if err != nil {
		uid = uuid.Nil
	}
	id := uid.String()
	md := make(metadata.Metadata)
	ctx = metadata.NewIncomingContext(ctx, md)
	ctx = requestid.SetIncomingRequestId(ctx, id)
	ctx = requestid.SetOutgoingRequestId(ctx, id)
	return ctx
}
