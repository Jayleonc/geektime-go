package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	gs := grpc.NewServer()
	us := &Server{}
	RegisterUserServiceServer(gs, us)

	listen, err := net.Listen("tcp", ":8090")

	require.NoError(t, err)

	err = gs.Serve(listen)
	t.Log(err)
}

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := NewUserServiceClient(cc)
	res, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(res.User)
}
