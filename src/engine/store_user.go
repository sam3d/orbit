package engine

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// User is any user who has access to a system.
type User struct {
	ID       string    `json:"id"` // Auto generated
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Password [60]byte  `json:"password"` // Bcrypt hashed field
	Email    string    `json:"email"`
	Profile  []byte    `json:"profile"`  // Image data in a byte slice
	Sessions []Session `json:"sessions"` // The session array
}

// Session is a user session that has a unique token that identifies it for the
// purpose of authenticating a user.
type Session struct {
	Token string `json:"token"`
}

// Users is a list of users.
type Users []User

// UserConfig configures the user generation process.
type UserConfig struct {
	Name     string
	Username string
	Password string
	Email    string
	Profile  []byte
}

// Generate creates a unique user that can be added to the store.
func (u *Users) Generate(config UserConfig) (*User, error) {
	// Ensure that the fields are present.
	if config.Name == "" ||
		config.Username == "" ||
		config.Password == "" ||
		config.Email == "" {
		return nil, ErrMissingFields
	}

	// Check for duplicates.
	for _, user := range *u {
		if user.Username == config.Username {
			return nil, ErrUsernameTaken
		}
		if user.Email == config.Email {
			return nil, ErrEmailTaken
		}
	}

	// Hash the password.
	rawHashed, err := bcrypt.GenerateFromPassword([]byte(config.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("could not hash password")
	}
	var hashed [60]byte
	copy(hashed[:], rawHashed)

	// Create the user, append and return it.
	newUser := User{
		ID:       u.GenerateID(),
		Name:     config.Name,
		Username: config.Username,
		Password: hashed,
		Email:    config.Email,
		Profile:  config.Profile,
	}

	return &newUser, nil
}

// GenerateID returns an available ID from the user. It will keep autogenerating
// until one is found, so this can take unlimited time (but in practice, pretty
// much never will).
func (u *Users) GenerateID() string {
search:
	for {
		b := make([]byte, 8)
		rand.Read(b)
		id := hex.EncodeToString(b)

		for _, user := range *u {
			if user.ID == id {
				continue search
			}
		}

		return id
	}
}

// GenerateSession will create and return a user session based off of the user
// provided.
func (u User) GenerateSession() Session {
	// Generate a random token.
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)

	// Create and return the session object.
	return Session{
		Token: token,
	}
}

// ValidatePassword will take in a plaintext password and return whether or not
// it is valid.
func (u User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password[:], []byte(password))
	return err == nil
}

// FindByID will search for the user in the given list of users. It returns the
// index of that user, or -1 if that does not exist.
func (u *Users) FindByID(id string) (int, *User) {
	for i, user := range *u {
		if user.ID == id {
			return i, &user
		}
	}
	return -1, nil
}

// Find will search a user by username, email address, or ID.
func (u *Users) Find(id string) *User {
	for _, user := range *u {
		if user.ID == id || user.Email == id || user.Username == id {
			return &user
		}
	}
	return nil
}

// Remove removes the user with the specified ID from the slice.
func (u *Users) Remove(id string) error {
	i, _ := u.FindByID(id)
	if i == -1 {
		return ErrNotFound
	}
	*u = append((*u)[:i], (*u)[i+1:]...)
	return nil
}
