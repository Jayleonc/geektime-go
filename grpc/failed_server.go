package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type FailedServer struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *FailedServer) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	log.Println("进来了 failover")
	return nil, status.Errorf(codes.Unavailable, "假装我被熔断了")
}
