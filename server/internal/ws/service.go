package ws

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repository Repository) Service {
	return Service{
		repo: repository,
	}
}

func (s *Service) GetUsernameById(c context.Context, id string) (string, error) {
	return s.repo.queries.GetUsernameById(c, id)
}
