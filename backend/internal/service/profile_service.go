package service

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
)

// ProfileService defines profile business logic
type ProfileService interface {
	GetProfile(ctx context.Context, userID string) (*ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, fullName string, email *string, address *string) (*ProfileResponse, error)
	SetUserType(ctx context.Context, userID string, userType string) error
	GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error)
	UpdateWorkerSkills(ctx context.Context, userID string, skills []string) error
	GetMyRatings(ctx context.Context, userID string) ([]RatingHistoryItem, error)
}

// ProfileResponse returned when getting profile
type ProfileResponse struct {
	ID          string  `json:"id"`
	PhoneNumber string  `json:"phone_number"`
	FullName    *string `json:"full_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	UserType    string  `json:"user_type"`
	Address     *string `json:"address,omitempty"`
	KycStatus   string  `json:"kyc_status"`
}

// UserStatsResponse returned for stats endpoint
type UserStatsResponse struct {
	JobsPosted    int64   `json:"jobs_posted"`
	JobsCompleted int64   `json:"jobs_completed"`
	TotalSpent    float64 `json:"total_spent"`
	TotalEarned   float64 `json:"total_earned"`
	AvgRating     float64 `json:"avg_rating"`
}

// RatingHistoryItem represents a single rating entry for the user
type RatingHistoryItem struct {
	JobID      string `json:"job_id"`
	JobTitle   string `json:"job_title"`
	OtherParty string `json:"other_party"`
	Rating     int    `json:"rating"`
	Review     string `json:"review"`
	IsGiven    bool   `json:"is_given"` // true = I gave this, false = I received this
	CreatedAt  string `json:"created_at"`
}

type profileService struct {
	userRepo repository.UserRepository
	jobRepo  repository.JobRepository
}

// NewProfileService creates a new profile service
func NewProfileService(userRepo repository.UserRepository, jobRepo repository.JobRepository) ProfileService {
	return &profileService{
		userRepo: userRepo,
		jobRepo:  jobRepo,
	}
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*ProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Try to get client profile for address
	var address *string
	clientProfile, err := s.userRepo.GetClientProfile(ctx, userID)
	if err == nil && clientProfile != nil {
		address = clientProfile.DefaultAddress
	}

	return &ProfileResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		Email:       user.Email,
		UserType:    string(user.UserType),
		Address:     address,
		KycStatus:   string(user.AadhaarKycStatus),
	}, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, userID string, fullName string, email *string, address *string) (*ProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update user fields
	if fullName != "" {
		user.FullName = &fullName
	}
	if email != nil {
		user.Email = email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Update address in client profile if user is a client
	if address != nil && user.UserType == domain.UserTypeClient {
		clientProfile, err := s.userRepo.GetClientProfile(ctx, userID)
		if err == nil && clientProfile != nil {
			clientProfile.DefaultAddress = address
			s.userRepo.UpdateClientProfile(ctx, clientProfile)
		}
	}

	return s.GetProfile(ctx, userID)
}

func (s *profileService) SetUserType(ctx context.Context, userID string, userType string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.UserType = domain.UserType(userType)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Create corresponding profile
	if userType == "WORKER" {
		workerProfile := &domain.WorkerProfile{
			UserID: userID,
			Skills: []string{},
			Tier:   domain.WorkerTierBronze,
		}
		return s.userRepo.CreateWorkerProfile(ctx, workerProfile)
	} else if userType == "CLIENT" {
		clientProfile := &domain.ClientProfile{
			UserID: userID,
		}
		return s.userRepo.CreateClientProfile(ctx, clientProfile)
	}

	return nil
}

func (s *profileService) GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats := &UserStatsResponse{}

	if user.UserType == domain.UserTypeClient {
		jobsPosted, totalSpent, avgRating, err := s.jobRepo.GetClientStats(ctx, userID)
		if err != nil {
			return nil, err
		}
		stats.JobsPosted = jobsPosted
		stats.TotalSpent = totalSpent
		stats.AvgRating = avgRating
	} else if user.UserType == domain.UserTypeWorker {
		jobsCompleted, totalEarned, err := s.jobRepo.GetWorkerStats(ctx, userID)
		if err != nil {
			return nil, err
		}
		stats.JobsCompleted = jobsCompleted
		stats.TotalEarned = totalEarned
		// Get worker rating from profile
		workerProfile, err := s.userRepo.GetWorkerProfile(ctx, userID)
		if err == nil && workerProfile != nil {
			stats.AvgRating = workerProfile.RatingAvg
		}
	}

	return stats, nil
}

func (s *profileService) UpdateWorkerSkills(ctx context.Context, userID string, skills []string) error {
	profile, err := s.userRepo.GetWorkerProfile(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return errors.New("worker profile not found")
	}
	profile.Skills = skills
	return s.userRepo.UpdateWorkerProfile(ctx, profile)
}

func (s *profileService) GetMyRatings(ctx context.Context, userID string) ([]RatingHistoryItem, error) {
	rows, err := s.jobRepo.GetUserRatings(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]RatingHistoryItem, len(rows))
	for i, r := range rows {
		items[i] = RatingHistoryItem{
			JobID:      r.JobID,
			JobTitle:   r.JobTitle,
			OtherParty: r.OtherName,
			Rating:     r.Rating,
			Review:     r.Review,
			IsGiven:    r.IsGiven,
			CreatedAt:  r.CreatedAt,
		}
	}
	return items, nil
}
