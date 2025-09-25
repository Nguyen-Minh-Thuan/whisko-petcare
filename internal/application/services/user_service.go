package services

import (
	"context"
	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/internal/infrastructure/projection"
)

// UserService orchestrates user operations
type UserService struct {
	// Command handlers (using Unit of Work)
	createUserHandler        *command.CreateUserWithUoWHandler
	updateUserProfileHandler *command.UpdateUserProfileWithUoWHandler
	updateUserContactHandler *command.UpdateUserContactWithUoWHandler
	deleteUserHandler        *command.DeleteUserWithUoWHandler

	// Query handlers
	getUserHandler     *query.GetUserHandler
	listUsersHandler   *query.ListUsersHandler
	searchUsersHandler *query.SearchUsersHandler
}

func NewUserService(
	createUserHandler *command.CreateUserWithUoWHandler,
	updateUserProfileHandler *command.UpdateUserProfileWithUoWHandler,
	updateUserContactHandler *command.UpdateUserContactWithUoWHandler,
	deleteUserHandler *command.DeleteUserWithUoWHandler,
	getUserHandler *query.GetUserHandler,
	listUsersHandler *query.ListUsersHandler,
	searchUsersHandler *query.SearchUsersHandler,
) *UserService {
	return &UserService{
		createUserHandler:        createUserHandler,
		updateUserProfileHandler: updateUserProfileHandler,
		updateUserContactHandler: updateUserContactHandler,
		deleteUserHandler:        deleteUserHandler,
		getUserHandler:           getUserHandler,
		listUsersHandler:         listUsersHandler,
		searchUsersHandler:       searchUsersHandler,
	}
}

// Command operations
func (s *UserService) CreateUser(ctx context.Context, cmd command.CreateUser) error {
	return s.createUserHandler.Handle(ctx, &cmd)
}

func (s *UserService) UpdateUserProfile(ctx context.Context, cmd command.UpdateUserProfile) error {
	return s.updateUserProfileHandler.Handle(ctx, &cmd)
}

func (s *UserService) UpdateUserContact(ctx context.Context, cmd command.UpdateUserContact) error {
	return s.updateUserContactHandler.Handle(ctx, &cmd)
}

func (s *UserService) DeleteUser(ctx context.Context, cmd command.DeleteUser) error {
	return s.deleteUserHandler.Handle(ctx, &cmd)
}

// Query operations
func (s *UserService) GetUser(ctx context.Context, query query.GetUser) (*projection.UserReadModel, error) {
	return s.getUserHandler.Handle(ctx, query)
}

func (s *UserService) ListUsers(ctx context.Context, query query.ListUsers) ([]*projection.UserReadModel, error) {
	return s.listUsersHandler.Handle(ctx, query)
}

func (s *UserService) SearchUsers(ctx context.Context, query query.SearchUsers) ([]*projection.UserReadModel, error) {
	return s.searchUsersHandler.Handle(ctx, query)
}
