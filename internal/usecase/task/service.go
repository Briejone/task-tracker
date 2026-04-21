package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/scheduler"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}
	var nextRunAt *time.Time
	if normalized.RepeatRule != nil && *normalized.RepeatRule != "" {
		if !scheduler.ValidateCron(*normalized.RepeatRule) {
			return nil, fmt.Errorf("%w: invalid cron expression", ErrInvalidInput)
		}
		next, err := scheduler.CalculateNextRun(*normalized.RepeatRule)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to calculate next run", ErrInvalidInput)
		}
		nextRunAt = &next
	}

	now := s.now()
	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		RepeatRule:  normalized.RepeatRule,
		NextRunAt:   nextRunAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}
// подумать
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != "" {
		existing.Title = strings.TrimSpace(input.Title)
	}
	if input.Description != "" {
		existing.Description = strings.TrimSpace(input.Description)
	}
	if input.Status != "" && input.Status.Valid() {
		existing.Status = input.Status
	}

	if input.RepeatRule != nil {
		if *input.RepeatRule == "" {
			existing.RepeatRule = nil
			existing.NextRunAt = nil
		} else {
			if !scheduler.ValidateCron(*input.RepeatRule) {
				return nil, fmt.Errorf("%w: invalid cron expression", ErrInvalidInput)
			}
			next, err := scheduler.CalculateNextRun(*input.RepeatRule)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to calculate next run", ErrInvalidInput)
			}
			rule := *input.RepeatRule
			existing.RepeatRule = &rule
			existing.NextRunAt = &next
		}
	}

	if existing.Title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if !existing.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	existing.UpdatedAt = s.now()

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	
	if input.RepeatRule != nil && *input.RepeatRule != "" {
		if !scheduler.ValidateCron(*input.RepeatRule) {
			return CreateInput{}, fmt.Errorf("%w: invalid cron expression", ErrInvalidInput)
		}
	}
	
	return input, nil
}
