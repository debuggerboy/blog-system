package service

import (
	"blog-system/internal/models"
	"blog-system/internal/repository"
	"errors"
)

type PostService struct {
	postRepo *repository.PostRepository
	userRepo *repository.UserRepository
}

func NewPostService() *PostService {
	return &PostService{
		postRepo: repository.NewPostRepository(),
		userRepo: repository.NewUserRepository(),
	}
}

func (s *PostService) CreatePost(req *models.PostRequest, authorID int64) (*models.Post, error) {
	post := &models.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
		Status:   req.Status,
		Pinned:   false, // Default to not pinned
	}

	if post.Status == "" {
		post.Status = "published"
	}

	err := s.postRepo.Create(post)
	if err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(post.ID)
}

func (s *PostService) GetPostByID(id int64) (*models.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *PostService) ListPosts(limit, offset int, userID int64, role string) ([]models.Post, error) {
	var filterStatus *string
	var filterAuthorID *int64

	if role != "admin" && role != "supervisor" {
		published := "published"
		filterStatus = &published
	}

	if role == "user" {
		filterAuthorID = &userID
	}

	return s.postRepo.List(limit, offset, filterAuthorID, filterStatus)
}

func (s *PostService) GetLatestPosts(limit int) ([]models.Post, error) {
	return s.postRepo.GetLatestPosts(limit)
}

func (s *PostService) GetUserPosts(authorID int64) ([]models.Post, error) {
	return s.postRepo.ListByAuthor(authorID, 100, 0)
}

func (s *PostService) UpdatePost(id int64, req *models.PostRequest, userID int64, role string) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	if role != "admin" && role != "supervisor" && post.AuthorID != userID {
		return nil, errors.New("unauthorized to edit this post")
	}

	post.Title = req.Title
	post.Content = req.Content
	post.Status = req.Status

	err = s.postRepo.Update(post)
	if err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(post.ID)
}

func (s *PostService) DeletePost(id int64, userID int64, role string) error {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}

	if role != "admin" && role != "supervisor" && post.AuthorID != userID {
		return errors.New("unauthorized to delete this post")
	}

	return s.postRepo.Delete(id)
}

func (s *PostService) TogglePostStatus(id int64, userID int64, role string) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	if role != "admin" && role != "supervisor" && post.AuthorID != userID {
		return nil, errors.New("unauthorized to modify this post")
	}

	if post.Status == "published" {
		post.Status = "disabled"
	} else {
		post.Status = "published"
	}

	err = s.postRepo.Update(post)
	if err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(post.ID)
}
