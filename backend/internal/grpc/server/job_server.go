package grpcserver

import (
	"context"
	"time"

	jobpb "github.com/decodejatin/bero-backend/gen/pb/job"
	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/matchmaker/h3"
	"github.com/decodejatin/bero-backend/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JobServer implements jobpb.JobServiceServer.
// It wraps the existing JobRepository, exposing job data and assignment
// operations as gRPC RPCs for internal consumption by the Matchmaker Service.
type JobServer struct {
	jobpb.UnimplementedJobServiceServer
	jobRepo repository.JobRepository
}

// NewJobServer creates a new gRPC JobService server.
func NewJobServer(jobRepo repository.JobRepository) *JobServer {
	return &JobServer{jobRepo: jobRepo}
}

// GetJob fetches a single job by ID.
func (s *JobServer) GetJob(ctx context.Context, req *jobpb.GetJobRequest) (*jobpb.JobResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}

	job, err := s.jobRepo.GetByID(ctx, req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "job not found: %v", err)
	}

	return jobToProto(job), nil
}

// ListOpenJobs returns unmatched open jobs in a time window.
func (s *JobServer) ListOpenJobs(ctx context.Context, req *jobpb.ListJobsRequest) (*jobpb.ListJobsResponse, error) {
	windowStart := time.Now().Add(-24 * time.Hour)
	if req.CreatedAfter != nil {
		windowStart = req.CreatedAfter.AsTime()
	}

	// Reuse matchmaker repo filter via a category-aware open jobs query
	var category *domain.ServiceCategory
	if req.Category != "" {
		cat := domain.ServiceCategory(req.Category)
		category = &cat
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 200
	}

	jobs, err := s.jobRepo.GetOpenJobs(ctx, "", category, limit, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list open jobs: %v", err)
	}

	// Filter to only unmatched + created after windowStart
	resp := &jobpb.ListJobsResponse{}
	for _, j := range jobs {
		if j.AssignedWorkerID != nil || j.CreatedAt.Before(windowStart) {
			continue
		}
		resp.Jobs = append(resp.Jobs, jobToProto(&j))
	}

	return resp, nil
}

// AssignWorker assigns a worker to a job atomically via the repository.
func (s *JobServer) AssignWorker(ctx context.Context, req *jobpb.AssignWorkerRequest) (*jobpb.AssignWorkerResponse, error) {
	if req.JobId == "" || req.WorkerId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id and worker_id are required")
	}

	err := s.jobRepo.AssignWorker(ctx, req.JobId, req.WorkerId)
	if err != nil {
		return &jobpb.AssignWorkerResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &jobpb.AssignWorkerResponse{Success: true}, nil
}

// --- Domain → Proto conversion helpers ---

func jobToProto(j *domain.Job) *jobpb.JobResponse {
	pb := &jobpb.JobResponse{
		JobId:           j.ID,
		Title:           j.Title,
		Category:        string(j.Category),
		Status:          string(j.Status),
		ClientId:        j.ClientID,
		Locality:        j.Locality,
		City:            j.City,
		PaymentRupees:   j.PaymentAmountRupees,
		IsUrgent:        j.IsUrgent,
		DurationMinutes: int32(j.EstimatedDurationMins),
		RequiredSkills:  j.RequiredSkills,
		CreatedAt:       timestamppb.New(j.CreatedAt),
		ScheduledDate:   timestamppb.New(j.ScheduledDate),
	}

	if j.Latitude != nil {
		pb.Latitude = *j.Latitude
	}
	if j.Longitude != nil {
		pb.Longitude = *j.Longitude
	}
	// Pre-compute H3 index if coordinates are available
	if pb.Latitude != 0 || pb.Longitude != 0 {
		pb.H3Index = string(h3.LatLngToCell(pb.Latitude, pb.Longitude, 9))
	}

	return pb
}
