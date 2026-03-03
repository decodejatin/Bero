package grpcclient

import (
	"context"
	"fmt"

	"time"

	jobpb "github.com/decodejatin/bero-backend/gen/pb/job"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JobClient is a typed gRPC client for the JobService.
// Used by the Matchmaker Service to fetch jobs and persist assignments
// over binary Protobuf instead of JSON.
type JobClient struct {
	conn   *grpc.ClientConn
	client jobpb.JobServiceClient
}

// NewJobClient creates a connected gRPC client to the JobService.
func NewJobClient(addr string) (*JobClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &JobClient{
		conn:   conn,
		client: jobpb.NewJobServiceClient(conn),
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *JobClient) Close() error {
	return c.conn.Close()
}

// GetJob fetches a single job by ID.
func (c *JobClient) GetJob(ctx context.Context, jobID string) (*jobpb.JobResponse, error) {
	return c.client.GetJob(ctx, &jobpb.GetJobRequest{JobId: jobID})
}

// ListOpenJobs returns unmatched open jobs created after createdAfter.
func (c *JobClient) ListOpenJobs(ctx context.Context, createdAfter time.Time, category string) ([]*jobpb.JobResponse, error) {
	resp, err := c.client.ListOpenJobs(ctx, &jobpb.ListJobsRequest{
		CreatedAfter: timestamppb.New(createdAfter),
		Category:     category,
		Limit:        500,
	})
	if err != nil {
		return nil, err
	}
	return resp.Jobs, nil
}

// AssignWorker assigns a worker to a job via gRPC.
func (c *JobClient) AssignWorker(ctx context.Context, jobID, workerID string) error {
	resp, err := c.client.AssignWorker(ctx, &jobpb.AssignWorkerRequest{
		JobId:    jobID,
		WorkerId: workerID,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("assign worker failed: %s", resp.Error)
	}
	return nil
}
