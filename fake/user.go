package fake

import "github.com/serdmanczyk/freyr/models"

// UserStore implements the models.UserStore interface for use in testing libraries
// that use a models.UserStore.
type UserStore map[string]models.User

// GetUser returns the user data for the given user email.
func (u UserStore) GetUser(email string) (models.User, error) {
	user, ok := u[email]
	if !ok {
		return models.User{}, nil
	}

	return user, nil
}

// StoreUser inserts/updates data for the given user.
func (u UserStore) StoreUser(user models.User) error {
	u[user.Email] = user
	return nil
}
