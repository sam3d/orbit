package engine

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// User is any user who has access to a system.
type User struct {
	ID       string   `json:"id"` // Auto generated
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Password [60]byte `json:"password"` // Bcrypt hashed field
	Email    string   `json:"email"`
}

// Users is a list of users.
type Users []User

// New will add a new user to the list of users and return the resulting user
// and an error.
func (u *Users) New(name, username, password, email string) (*User, error) {
	// Check for duplicates.
	for _, user := range *u {
		if user.Username == username {
			return nil, errors.New("username already exists on the store")
		}
		if user.Email == email {
			return nil, errors.New("email already exists on the store")
		}
	}

	// Hash the password.
	rawHashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("could not hash password")
	}
	var hashed [60]byte
	copy(hashed[:], rawHashed)

	// Create the user, append and return it.
	newUser := User{
		ID:       u.GenerateID(),
		Name:     name,
		Username: username,
		Password: hashed,
		Email:    email,
	}
	*u = append(*u, newUser)
	return &newUser, nil
}

// GenerateID returns an available ID from the user. It will keep autogenerating
// until one is found, so this can take unlimited time (but in practice, never
// will).
func (u *Users) GenerateID() string {
	for {
		b := make([]byte, 16)
		rand.Read(b)
		id := hex.EncodeToString(b)

		collision := false
		for _, user := range *u {
			if user.ID == id {
				collision = true
				break
			}
		}
		if !collision {
			return id
		}
	}
}
