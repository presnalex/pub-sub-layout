package handler

import (
	"context"

	"github.com/google/uuid"
	raw "github.com/presnalex/codec-bytes"
	log "github.com/presnalex/go-micro/v3/logger"
	"go.unistack.org/micro/v3/metadata"
)

type Handler struct {
	storage IStorage
}

type IStorage interface {
	AnimalAdd(ctx context.Context, msg []byte) error
}

func (h *Handler) Subscribe(ctx context.Context, msg *raw.Frame) error {
	logger := log.FromIncomingContext(ctx)
	logger.Debug(ctx, "pub-sub-layout handler started")
	// Set x-request-id
	md, _ := metadata.FromIncomingContext(ctx)
	uid, err := uuid.NewRandom()
	if err != nil {
		uid = uuid.Nil
	}
	md.Set("x-request-id", uid.String())
	ctx = metadata.NewOutgoingContext(ctx, md)

	h.storage.AnimalAdd(ctx, msg.Data)
	return nil
}

func NewHandler(storage IStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}
