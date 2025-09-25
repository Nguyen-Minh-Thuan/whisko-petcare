// Example of how to wire up the Unit of Work pattern in your main.go
// This shows the new approach alongside your existing code

package examples

/*
Example integration of Unit of Work pattern:

1. Initialize Unit of Work Factory:
   uowFactory := mongo.NewMongoUnitOfWorkFactory(mongoClient, database)

2. Create command handlers with UoW:
   createUserHandler := command.NewCreateUserWithUoWHandler(uowFactory, eventBus)
   updateUserProfileHandler := command.NewUpdateUserProfileWithUoWHandler(uowFactory, eventBus)
   updateUserContactHandler := command.NewUpdateUserContactWithUoWHandler(uowFactory, eventBus)
   deleteUserHandler := command.NewDeleteUserWithUoWHandler(uowFactory, eventBus)

3. Usage in service methods:
   func (s *UserService) CreateUser(ctx context.Context, cmd *command.CreateUser) error {
       return s.createUserHandler.Handle(ctx, cmd)
   }

4. Transaction handling is automatic within each handler:
   - Begin transaction
   - Execute business logic
   - Save to repositories
   - Publish events
   - Commit transaction (or rollback on error)

Benefits of this approach:
- Atomic operations across multiple repositories
- Automatic transaction management
- Consistent error handling
- Better separation of concerns
- Easier testing and mocking
*/
