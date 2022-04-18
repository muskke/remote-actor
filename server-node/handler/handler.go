package handler

import (
	"context"
	pb "github.com/muskke/remote-actor/proto"
)

type handler struct{}

func (h handler) Call(ctx context.Context, ready *pb.Ready, done *pb.Done) error {
	panic("implement me")
}

func NewHandler() *handler{
	return &handler{}
}

