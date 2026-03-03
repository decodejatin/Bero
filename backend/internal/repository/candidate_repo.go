package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

// CandidateRepository handles spatial + skill-filtered worker queries
// optimized for the matching pipeline. Uses PostGIS GiST index + GIN
// index on skills for sub-30ms candidate retrieval.
type CandidateRepository interface {
	// FindCandidateWorkers returns available workers near a job location,
	// filtered by skill match, ordered by distance ASC + reputation DESC.
	FindCandidateWorkers(
		ctx context.Context,
		jobLat, jobLon, radiusMeters float64,
		category string,
		requiredSkills []string,
		limit int,
	) ([]domain.WorkerCandidate, error)
}

type candidateRepository struct {
	db *gorm.DB
}

// NewCandidateRepository creates a new candidate repository.
func NewCandidateRepository(db *gorm.DB) CandidateRepository {
	return &candidateRepository{db: db}
}

func (r *candidateRepository) FindCandidateWorkers(
	ctx context.Context,
	jobLat, jobLon, radiusMeters float64,
	category string,
	requiredSkills []string,
	limit int,
) ([]domain.WorkerCandidate, error) {
	if limit <= 0 {
		limit = 20
	}

	// Build the skill filter clause dynamically.
	// If requiredSkills is provided, use jsonb overlap operator (?|)
	// which checks if the worker's skills array contains ANY of the required skills.
	// This leverages the GIN index on worker_profiles.skills for performance.
	skillFilter := ""
	args := []sql.NamedArg{
		sql.Named("lat", jobLat),
		sql.Named("lon", jobLon),
		sql.Named("radius", radiusMeters),
		sql.Named("lim", limit),
	}

	if len(requiredSkills) > 0 {
		// Use PostgreSQL array overlap: skills::jsonb ?| array['PLUMBING','ELECTRICAL']
		skillFilter = "AND wp.skills::jsonb ?| @skills"
		args = append(args, sql.Named("skills", pqTextArray(requiredSkills)))
	} else if category != "" {
		// Fallback: match category as a single skill
		skillFilter = "AND wp.skills::jsonb ? @category"
		args = append(args, sql.Named("category", category))
	}

	query := fmt.Sprintf(`
		SELECT
			wp.user_id                                              AS worker_id,
			COALESCE(u.full_name, '')                               AS name,
			wp.latitude,
			wp.longitude,
			ST_Distance(
				wp.location::geography,
				ST_SetSRID(ST_MakePoint(@lon, @lat), 4326)::geography
			)                                                       AS distance_meters,
			COALESCE(wp.h3_index, '')                               AS h3_index,
			wp.rating_avg,
			COALESCE(wp.rating_count, 0)                            AS rating_count,
			wp.tier
		FROM worker_profiles wp
		JOIN users u ON u.id = wp.user_id
		WHERE wp.is_available = TRUE
		  AND wp.location IS NOT NULL
		  AND ST_DWithin(
				wp.location::geography,
				ST_SetSRID(ST_MakePoint(@lon, @lat), 4326)::geography,
				@radius
			)
		  %s
		ORDER BY distance_meters ASC, wp.rating_avg DESC
		LIMIT @lim
	`, skillFilter)

	// Convert named args to interface slice for GORM
	gormArgs := make([]interface{}, len(args))
	for i, a := range args {
		gormArgs[i] = a
	}

	var candidates []domain.WorkerCandidate
	result := r.db.WithContext(ctx).Raw(query, gormArgs...).Scan(&candidates)
	if result.Error != nil {
		return nil, fmt.Errorf("find candidate workers: %w", result.Error)
	}
	if candidates == nil {
		candidates = []domain.WorkerCandidate{}
	}
	return candidates, nil
}

// pqTextArray formats a Go string slice into a PostgreSQL text array literal.
// E.g., ["PLUMBING", "ELECTRICAL"] → ARRAY['PLUMBING','ELECTRICAL']
func pqTextArray(ss []string) string {
	if len(ss) == 0 {
		return "ARRAY[]::text[]"
	}
	result := "ARRAY["
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += "'" + s + "'"
	}
	result += "]"
	return result
}
