package service

import (
	"context"
	"log"
	"time"

	"github.com/decodejatin/bero-backend/internal/matchmaker"
	"github.com/decodejatin/bero-backend/internal/matchmaker/h3"
	"github.com/decodejatin/bero-backend/internal/repository"
)

// MatchmakerService manages the matching engine lifecycle.
type MatchmakerService interface {
	// StartEngine starts the matching engine goroutine.
	StartEngine(ctx context.Context)

	// Trigger manually triggers a matching round.
	Trigger()

	// GetConfig returns current matching configuration.
	GetConfig() matchmaker.MatchConfig

	// UpdateConfig updates the matching configuration.
	UpdateConfig(cfg matchmaker.MatchConfig)

	// GetStatus returns the current engine status.
	GetStatus() matchmaker.EngineStatus

	// GetSupplyDensity returns worker density per H3 cell.
	GetSupplyDensity(ctx context.Context) ([]matchmaker.SupplyDensityInfo, error)

	// GetEngine exposes the underlying matching engine for gRPC server registration.
	GetEngine() *matchmaker.Engine
}

type matchmakerService struct {
	engine    *matchmaker.Engine
	matchRepo repository.MatchmakerRepository
	jobRepo   repository.JobRepository
}

// NewMatchmakerService creates a new matchmaker service.
func NewMatchmakerService(
	matchRepo repository.MatchmakerRepository,
	jobRepo repository.JobRepository,
	cfg matchmaker.MatchConfig,
) MatchmakerService {
	svc := &matchmakerService{
		matchRepo: matchRepo,
		jobRepo:   jobRepo,
	}

	// Create the engine with fetcher and assignment callbacks
	engine := matchmaker.NewEngine(
		cfg,
		svc.fetchWorkers,
		svc.fetchJobs,
		svc.onAssignment,
	)
	svc.engine = engine

	return svc
}

// fetchWorkers converts domain WorkerProfiles into MatchableWorkers.
// Uses H3 spatial indexing: converts H3IndexRes9 → lat/lng for proximity calculations.
func (s *matchmakerService) fetchWorkers(ctx context.Context) ([]matchmaker.MatchableWorker, error) {
	profiles, err := s.matchRepo.GetOnlineWorkers(ctx, nil)
	if err != nil {
		return nil, err
	}

	workers := make([]matchmaker.MatchableWorker, 0, len(profiles))
	for _, p := range profiles {
		// Skip workers who can't accept jobs (negative wallet)
		if !p.CanAcceptJobs() {
			continue
		}

		w := matchmaker.MatchableWorker{
			ID:        p.UserID,
			Skills:    p.Skills,
			RatingAvg: p.RatingAvg,
			IsOnline:  p.IsOnline,
		}

		// Convert H3 index to lat/lng for proximity calculations
		if p.H3IndexRes9 != nil && *p.H3IndexRes9 != "" {
			w.H3Index = *p.H3IndexRes9
			center, err := h3.CellToLatLng(h3.Cell(w.H3Index))
			if err == nil {
				w.Latitude = center.Lat
				w.Longitude = center.Lng
			}
		}

		workers = append(workers, w)
	}

	return workers, nil
}

// fetchJobs converts domain Jobs into MatchableJobs.
// Computes H3 cell index from job lat/lng for spatial queries.
func (s *matchmakerService) fetchJobs(ctx context.Context) ([]matchmaker.MatchableJob, error) {
	// Get open jobs from the last 24 hours (configurable window)
	windowStart := time.Now().Add(-24 * time.Hour)
	domainJobs, err := s.matchRepo.GetUnmatchedOpenJobs(ctx, windowStart)
	if err != nil {
		return nil, err
	}

	cfg := s.engine.Config()
	jobs := make([]matchmaker.MatchableJob, 0, len(domainJobs))
	for _, dj := range domainJobs {
		j := matchmaker.MatchableJob{
			ID:             dj.ID,
			RequiredSkills: dj.RequiredSkills,
			Category:       string(dj.Category),
			CreatedAt:      dj.CreatedAt,
		}

		if dj.Latitude != nil {
			j.Latitude = *dj.Latitude
		}
		if dj.Longitude != nil {
			j.Longitude = *dj.Longitude
		}

		// Compute H3 index for spatial queries
		if j.Latitude != 0 || j.Longitude != 0 {
			j.H3Index = string(h3.LatLngToCell(j.Latitude, j.Longitude, cfg.H3Resolution))
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

// onAssignment persists a matching assignment by assigning the worker to the job.
func (s *matchmakerService) onAssignment(ctx context.Context, assignment matchmaker.Assignment) error {
	log.Printf("[matchmaker] Assigning worker %s → job %s (weight: %.4f)",
		assignment.WorkerID, assignment.JobID, assignment.Weight)

	return s.jobRepo.AssignWorker(ctx, assignment.JobID, assignment.WorkerID)
}

// StartEngine starts the matching engine goroutine.
func (s *matchmakerService) StartEngine(ctx context.Context) {
	go s.engine.Run(ctx)
	log.Println("[matchmaker] Engine started in background")
}

// Trigger manually triggers a matching round.
func (s *matchmakerService) Trigger() {
	s.engine.Trigger()
}

// GetConfig returns current matching configuration.
func (s *matchmakerService) GetConfig() matchmaker.MatchConfig {
	return s.engine.Config()
}

// UpdateConfig updates the matching configuration.
func (s *matchmakerService) UpdateConfig(cfg matchmaker.MatchConfig) {
	s.engine.UpdateConfig(cfg)
}

// GetStatus returns the current engine status.
func (s *matchmakerService) GetStatus() matchmaker.EngineStatus {
	return s.engine.Status()
}

// GetSupplyDensity returns worker density per H3 cell.
func (s *matchmakerService) GetSupplyDensity(ctx context.Context) ([]matchmaker.SupplyDensityInfo, error) {
	workers, err := s.fetchWorkers(ctx)
	if err != nil {
		return nil, err
	}

	cfg := s.engine.Config()
	density := matchmaker.GetSupplyDensity(workers, cfg.H3Resolution)
	return density, nil
}

// GetEngine returns the underlying matching engine.
// Used by the gRPC MatcherServer for direct engine access.
func (s *matchmakerService) GetEngine() *matchmaker.Engine {
	return s.engine
}
