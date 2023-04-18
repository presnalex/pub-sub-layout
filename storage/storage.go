package storage

import (
	"context"
	"encoding/json"
	"errors"

	dbwrapper "github.com/presnalex/go-micro/v3/database/wrapper"
	log "github.com/presnalex/go-micro/v3/logger"
	"go.unistack.org/micro/v3/client"
)

type IClient interface {
	Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error
	NewMessage(topic string, msg interface{}, opts ...client.MessageOption) client.Message
}

type Storage struct {
	DB     *dbwrapper.Wrapper
	Client IClient
	Topics Topics
}

type Topics struct {
	AnimalAdd   string
	AnimalAddRs string
}

func (s *Storage) AnimalAdd(ctx context.Context, msg []byte) error {
	logger := log.FromIncomingContext(ctx)
	// Unmarshal incoming message
	var animalMsg = new(AnimalMsg)
	err := json.Unmarshal(msg, animalMsg)
	if err != nil {
		logger.Errorf(ctx, "unable to unmarshall incoming message: %v", err)
	}

	// Create message for response
	responseMsg := new(ResponseMsg)
	responseMsg.MessageId = animalMsg.AnimalId

	txn, err := s.DB.BeginTxx(dbwrapper.QueryContext(ctx, "transaction start"), nil)
	if err != nil {
		logger.Errorf(ctx, "unable to start db transaction msg error: %v", err)
		s.rollBack(ctx, txn)
		s.publishResponse(ctx, responseMsg, err)
		return nil
	}

	insertResult, err := txn.ExecContext(dbwrapper.QueryContext(ctx, "queryAddAnimal"), queryAddAnimal, animalMsg.AnimalId, animalMsg.Animal, animalMsg.Price)
	if err != nil {
		logger.Errorf(ctx, "unable to add record db error: %v", err)
		s.rollBack(ctx, txn)
		s.publishResponse(ctx, responseMsg, err)
		return nil
	}

	rowsAffected, err := insertResult.RowsAffected()
	if err != nil || rowsAffected == 0 {
		err = errors.New("empty response from db")
		logger.Error(ctx, err)
		s.rollBack(ctx, txn)
		s.publishResponse(ctx, responseMsg, err)
		return nil
	}

	err = txn.Commit()
	if err != nil {
		logger.Errorf(ctx, "unable to commit transaction error: %v", err)
		s.publishResponse(ctx, responseMsg, err)
		return nil
	}

	s.publishResponse(ctx, responseMsg, nil)
	return nil
}

func (s *Storage) publishResponse(ctx context.Context, responseMsg *ResponseMsg, err error) {
	responseMsg.Status = "Ok"
	if err != nil {
		responseMsg.Status = "Error"
		responseMsg.ErrorDetails = err.Error()
	}

	s.publish(ctx, s.Topics.AnimalAddRs, responseMsg)
}

func (s *Storage) publish(ctx context.Context, topic string, msg interface{}) {
	logger := log.FromIncomingContext(ctx)
	logger.Debugf(ctx, "publishing msg to %s topic", topic)

	err := s.Client.Publish(ctx, s.Client.NewMessage(topic, msg, client.WithMessageContentType("application/json")), client.PublishBodyOnly(true))
	if err != nil {
		logger.Fatalf(ctx, "publication to topic %s error: %v. message: %s", topic, err, msg)
	}
}

func (s *Storage) rollBack(ctx context.Context, txn *dbwrapper.TxWrapper) {
	err := txn.Rollback()
	if err != nil {
		log.FromIncomingContext(ctx).Fatalf(ctx, "unable to rollback a transaction: %v", err)
	}
}

func NewStorage(db *dbwrapper.Wrapper, clnt IClient, topics Topics) *Storage {
	return &Storage{
		DB:     db,
		Client: clnt,
		Topics: topics,
	}
}
