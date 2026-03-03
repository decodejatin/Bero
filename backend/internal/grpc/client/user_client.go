package grpcclient

import (
	"context"

	userpb "github.com/decodejatin/bero-backend/gen/pb/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserClient is a typed gRPC client for the UserService.
// Used by the Matchmaker Service to fetch online workers over binary Protobuf
// instead of JSON — 5-10× faster parse time at scale.
type UserClient struct {
	conn   *grpc.ClientConn
	client userpb.UserServiceClient
}

// NewUserClient creates a connected gRPC client to the UserService.
// addr example: "localhost:50051"
func NewUserClient(addr string) (*UserClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &UserClient{
		conn:   conn,
		client: userpb.NewUserServiceClient(conn),
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *UserClient) Close() error {
	return c.conn.Close()
}

// GetWorkerProfile fetches a single worker profile by ID.
func (c *UserClient) GetWorkerProfile(ctx context.Context, userID string) (*userpb.WorkerProfileResponse, error) {
	return c.client.GetWorkerProfile(ctx, &userpb.GetWorkerRequest{UserId: userID})
}

// ListOnlineWorkers returns all online workers, optionally filtered by H3 cells or category.
func (c *UserClient) ListOnlineWorkers(ctx context.Context, h3Cells []string, category string) ([]*userpb.WorkerProfileResponse, error) {
	resp, err := c.client.ListOnlineWorkers(ctx, &userpb.ListWorkersRequest{
		H3Cells:  h3Cells,
		Category: category,
	})
	if err != nil {
		return nil, err
	}
	return resp.Workers, nil
}
