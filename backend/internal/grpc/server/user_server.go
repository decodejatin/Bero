package grpcserver

import (
	"context"

	userpb "github.com/decodejatin/bero-backend/gen/pb/user"
	"github.com/decodejatin/bero-backend/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServer implements userpb.UserServiceServer.
// It wraps the existing UserRepository and MatchmakerRepository,
// exposing worker profile data via gRPC for internal service consumption.
type UserServer struct {
	userpb.UnimplementedUserServiceServer
	userRepo       repository.UserRepository
	matchmakerRepo repository.MatchmakerRepository
}

// NewUserServer creates a new gRPC UserService server.
func NewUserServer(userRepo repository.UserRepository, matchmakerRepo repository.MatchmakerRepository) *UserServer {
	return &UserServer{
		userRepo:       userRepo,
		matchmakerRepo: matchmakerRepo,
	}
}

// GetWorkerProfile fetches a single worker profile by user ID.
func (s *UserServer) GetWorkerProfile(ctx context.Context, req *userpb.GetWorkerRequest) (*userpb.WorkerProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	profile, err := s.userRepo.GetWorkerProfile(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "worker profile not found: %v", err)
	}

	return workerProfileToProto(profile), nil
}

// ListOnlineWorkers returns all workers currently marked as online,
// optionally filtered by H3 cells or category.
func (s *UserServer) ListOnlineWorkers(ctx context.Context, req *userpb.ListWorkersRequest) (*userpb.ListWorkersResponse, error) {
	var profiles []interface { /* domain.WorkerProfile */
	}

	if len(req.H3Cells) > 0 {
		// Spatial query: workers in specific H3 cells (O(1) lookup via DB index)
		domainProfiles, err := s.matchmakerRepo.GetOnlineWorkersInCells(ctx, req.H3Cells)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to query workers by cells: %v", err)
		}
		resp := &userpb.ListWorkersResponse{}
		for _, p := range domainProfiles {
			resp.Workers = append(resp.Workers, workerProfileToProto(&p))
		}
		return resp, nil
	}

	// Full scan: all online workers (filtered by skill category if provided)
	domainProfiles, err := s.matchmakerRepo.GetOnlineWorkers(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query online workers: %v", err)
	}

	_ = profiles
	resp := &userpb.ListWorkersResponse{}
	for _, p := range domainProfiles {
		// Optional: filter by category (skills contain the category keyword)
		if req.Category != "" {
			matched := false
			for _, skill := range p.Skills {
				if skill == req.Category {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		pb := workerProfileToProto(&p)
		resp.Workers = append(resp.Workers, pb)
	}

	return resp, nil
}
