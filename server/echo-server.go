package main

import (
	"chat-grpc/echo/proto"
	"context"
)

type EchoServer struct {
	proto.UnimplementedEchoServiceServer
}

func NewEchoServer() *EchoServer {
	return &EchoServer{}
}

func (s *EchoServer) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	return &proto.EchoResponse{Message: req.Message}, nil
}
