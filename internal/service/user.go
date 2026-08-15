package service

import (
	"blog-system/internal/models"
	"blog-system/internal/repository"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *UserService) GetUserByID(id int64) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) ListUsers() ([]models.User, error) {
	return s.userRepo.List()
}

func (s *UserService) UpdateUser(id int64, req *models.UserUpdateRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if username is taken by another user
	if req.Username != user.Username {
		existing, err := s.userRepo.FindByUsername(req.Username)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("username already taken")
		}
	}

	// Check if email is taken by another user
	if req.Email != user.Email {
		existing, err := s.userRepo.FindByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("email already registered")
		}
	}

	user.Username = req.Username
	user.Email = req.Email
	user.Role = req.Role
	user.Status = req.Status

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdatePassword(id int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(id, string(hashedPassword))
}

func (s *UserService) DeleteUser(id int64) error {
	// Don't allow deleting the last admin
	users, err := s.userRepo.List()
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.Role == "admin" {
		adminCount := 0
		for _, u := range users {
			if u.Role == "admin" {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return errors.New("cannot delete the last admin user")
		}
	}

	// Delete refresh tokens first
	err = s.userRepo.DeleteUserRefreshTokens(id)
	if err != nil {
		return err
	}

	return s.userRepo.Delete(id)
}
