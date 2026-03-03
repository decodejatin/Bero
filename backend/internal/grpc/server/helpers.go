package grpcserver

import (
	userpb "github.com/decodejatin/bero-backend/gen/pb/user"
	"github.com/decodejatin/bero-backend/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// workerProfileToProto converts a domain.WorkerProfile to the proto wire type.
func workerProfileToProto(p *domain.WorkerProfile) *userpb.WorkerProfileResponse {
	pb := &userpb.WorkerProfileResponse{
		UserId:       p.UserID,
		Skills:       p.Skills,
		IsOnline:     p.IsOnline,
		RatingAvg:    p.RatingAvg,
		RatingCount:  int32(p.RatingCount),
		WalletMicros: p.WalletBalanceMicros,
		Tier:         string(p.Tier),
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
	}

	if p.H3IndexRes9 != nil {
		pb.H3IndexRes9 = *p.H3IndexRes9
	}

	// Populate user fields if relation is preloaded
	if p.User.PhoneNumber != "" {
		pb.PhoneNumber = p.User.PhoneNumber
	}
	if p.User.FullName != nil {
		pb.FullName = *p.User.FullName
	}

	return pb
}
