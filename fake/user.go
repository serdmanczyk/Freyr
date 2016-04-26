package fake

import "github.com/serdmanczyk/freyr/models"

type UserStore map[string]models.User

func (u UserStore) GetUser(email string) (models.User, error) {
	user, ok := u[email]
	if !ok {
		return models.User{}, nil
	}

	return user, nil
}

func (u UserStore) StoreUser(user models.User) error {
	u[user.Email] = user
	return nil
}
