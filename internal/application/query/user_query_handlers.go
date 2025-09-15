package query

import (
	"context"
	"strings"
	"whisko-petcare/internal/infrastructure/projection"
	"whisko-petcare/pkg/errors"
)

// Queries
type GetUser struct {
	UserID string `json:"user_id"`
}

type ListUsers struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type SearchUsers struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// GetUserHandler handles user retrieval from read model
type GetUserHandler struct {
	userProjection projection.UserProjection
}

func NewGetUserHandler(userProjection projection.UserProjection) *GetUserHandler {
	return &GetUserHandler{
		userProjection: userProjection,
	}
}

func (h *GetUserHandler) Handle(ctx context.Context, query GetUser) (*projection.UserReadModel, error) {
	if strings.TrimSpace(query.UserID) == "" {
		return nil, errors.NewValidationError("user ID is required")
	}

	user, err := h.userProjection.GetByID(ctx, query.UserID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	return user, nil
}

// ListUsersHandler handles user listing from read model
type ListUsersHandler struct {
	userProjection projection.UserProjection
}

func NewListUsersHandler(userProjection projection.UserProjection) *ListUsersHandler {
	return &ListUsersHandler{
		userProjection: userProjection,
	}
}

func (h *ListUsersHandler) Handle(ctx context.Context, query ListUsers) ([]*projection.UserReadModel, error) {
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	users, err := h.userProjection.List(ctx, query.Limit, query.Offset)
	if err != nil {
		return nil, errors.NewInternalError("failed to get users")
	}

	return users, nil
}

// SearchUsersHandler handles user search from read model
type SearchUsersHandler struct {
	userProjection projection.UserProjection
}

func NewSearchUsersHandler(userProjection projection.UserProjection) *SearchUsersHandler {
	return &SearchUsersHandler{
		userProjection: userProjection,
	}
}

func (h *SearchUsersHandler) Handle(ctx context.Context, query SearchUsers) ([]*projection.UserReadModel, error) {
	users, err := h.userProjection.Search(ctx, query.Name, query.Email)
	if err != nil {
		return nil, errors.NewInternalError("failed to search users")
	}

	return users, nil
}
