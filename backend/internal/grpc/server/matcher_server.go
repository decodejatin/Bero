package grpcserver

import (
	"context"
	"log"
	"time"

	matcherpb "github.com/decodejatin/bero-backend/gen/pb/matcher"
	"github.com/decodejatin/bero-backend/internal/matchmaker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MatcherServer implements matcherpb.MatcherServiceServer.
// It wraps the matchmaker Engine, exposing matching operations and real-time
// streaming of assignment events for downstream consumers (notifications, audit).
type MatcherServer struct {
	matcherpb.UnimplementedMatcherServiceServer
	engine *matchmaker.Engine
}

// NewMatcherServer creates a new gRPC MatcherService server.
func NewMatcherServer(engine *matchmaker.Engine) *MatcherServer {
	return &MatcherServer{engine: engine}
}

// TriggerRound manually triggers a matching round.
func (s *MatcherServer) TriggerRound(ctx context.Context, _ *matcherpb.TriggerRequest) (*matcherpb.TriggerResponse, error) {
	status0 := s.engine.Status()
	s.engine.Trigger()
	return &matcherpb.TriggerResponse{
		Triggered: true,
		WindowId:  status0.WindowID + 1,
	}, nil
}

// GetStatus returns current engine state.
func (s *MatcherServer) GetStatus(ctx context.Context, _ *matcherpb.StatusRequest) (*matcherpb.StatusResponse, error) {
	st := s.engine.Status()
	return &matcherpb.StatusResponse{
		Running:            st.Running,
		TotalRounds:        st.TotalRounds,
		TotalMatches:       st.TotalMatches,
		WindowId:           st.WindowID,
		CircuitState:       st.CircuitState,
		LastShadowDeltaPct: st.LastShadowDelta,
		LastRunAt:          timestamppb.New(st.LastRunAt),
	}, nil
}

// GetConfig returns current matching configuration.
func (s *MatcherServer) GetConfig(ctx context.Context, _ *matcherpb.ConfigRequest) (*matcherpb.MatchConfigResponse, error) {
	cfg := s.engine.Config()
	return matchConfigToProto(cfg), nil
}

// UpdateConfig updates matching configuration.
func (s *MatcherServer) UpdateConfig(ctx context.Context, req *matcherpb.UpdateConfigRequest) (*matcherpb.MatchConfigResponse, error) {
	if req.Config == nil {
		return nil, status.Error(codes.InvalidArgument, "config is required")
	}
	cfg := protoToMatchConfig(req.Config)
	if cfg.MaxDistanceKm <= 0 {
		return nil, status.Error(codes.InvalidArgument, "max_distance_km must be positive")
	}
	if cfg.WindowDurationSeconds < 5 {
		return nil, status.Error(codes.InvalidArgument, "window_duration_s must be >= 5")
	}
	s.engine.UpdateConfig(cfg)
	return matchConfigToProto(cfg), nil
}

// StreamAssignments streams real-time assignment events to the caller.
// The engine's Results channel is consumed and each assignment is emitted
// as a separate AssignmentEvent message.
func (s *MatcherServer) StreamAssignments(_ *matcherpb.StreamRequest, stream matcherpb.MatcherService_StreamAssignmentsServer) error {
	log.Printf("[grpc] StreamAssignments: client connected")

	results := s.engine.Results()
	for {
		select {
		case result, ok := <-results:
			if !ok {
				// Engine stopped — stream ends
				log.Printf("[grpc] StreamAssignments: engine stopped, closing stream")
				return nil
			}
			for _, a := range result.Assignments {
				event := &matcherpb.AssignmentEvent{
					WorkerId:   a.WorkerID,
					JobId:      a.JobID,
					Weight:     a.Weight,
					WindowId:   result.WindowID,
					AssignedAt: timestamppb.New(time.Now()),
				}
				if err := stream.Send(event); err != nil {
					log.Printf("[grpc] StreamAssignments: send error: %v", err)
					return err
				}
			}
		case <-stream.Context().Done():
			log.Printf("[grpc] StreamAssignments: client disconnected")
			return nil
		}
	}
}

// --- Proto ↔ Domain conversions ---

func matchConfigToProto(cfg matchmaker.MatchConfig) *matcherpb.MatchConfigResponse {
	return &matcherpb.MatchConfigResponse{
		AlphaProximity:     cfg.AlphaProximity,
		AlphaReputation:    cfg.AlphaReputation,
		AlphaSkillMatch:    cfg.AlphaSkillMatch,
		AlphaWaitTime:      cfg.AlphaWaitTime,
		WindowDurationS:    int32(cfg.WindowDurationSeconds),
		MaxDistanceKm:      cfg.MaxDistanceKm,
		MinWeightThreshold: cfg.MinWeightThreshold,
		H3Resolution:       int32(cfg.H3Resolution),
		KNearestNeighbors:  int32(cfg.KNearestNeighbors),
		EnablePruning:      cfg.EnablePruning,
		LambdaDecay:        cfg.LambdaDecay,
		SwitchingCost:      cfg.SwitchingCost,
		EnableStability:    cfg.EnableStability,
		EnableShadowMode:   cfg.EnableShadowMode,
	}
}

func protoToMatchConfig(pb *matcherpb.MatchConfigResponse) matchmaker.MatchConfig {
	return matchmaker.MatchConfig{
		AlphaProximity:        pb.AlphaProximity,
		AlphaReputation:       pb.AlphaReputation,
		AlphaSkillMatch:       pb.AlphaSkillMatch,
		AlphaWaitTime:         pb.AlphaWaitTime,
		WindowDurationSeconds: int(pb.WindowDurationS),
		MaxDistanceKm:         pb.MaxDistanceKm,
		MinWeightThreshold:    pb.MinWeightThreshold,
		H3Resolution:          int(pb.H3Resolution),
		KNearestNeighbors:     int(pb.KNearestNeighbors),
		EnablePruning:         pb.EnablePruning,
		LambdaDecay:           pb.LambdaDecay,
		SwitchingCost:         pb.SwitchingCost,
		EnableStability:       pb.EnableStability,
		EnableShadowMode:      pb.EnableShadowMode,
	}
}
