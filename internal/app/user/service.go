package user

import (
	model "anhnq/api-core/internal/models"
	repository "anhnq/api-core/internal/repositories"
)

type Service struct {
	repo repository.UserRepository
}

func NewService(r repository.UserRepository) *Service {
	return &Service{repo: r}
}

func (s *Service) GetAll() ([]model.User, error) {
	return s.repo.FindAll()
}

func (s *Service) GetByID(id string) (model.User, error) {
	return s.repo.FindByID(id)
}

func (s *Service) Create(u model.User) (model.User, error) {
	return s.repo.Create(u)
}

func (s *Service) Update(id string, u model.User) (model.User, error) {
	return s.repo.Update(id, u)
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}
