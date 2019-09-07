/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"time"
)

var (
	// MinPasswordLength defines minimum password length
	MinPasswordLength = 6

	// ErrPasswordTooShort ...
	ErrPasswordTooShort = fmt.Errorf(
		"password must be at least %d characters long",
		MinPasswordLength,
	)
	// ErrUserNotFound ...
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidUserPassword ...
	ErrInvalidUserPassword = errors.New("invalid user password")
	// ErrCannotSetEmptyUsername ...
	ErrCannotSetEmptyUsername = errors.New("cannot set empty username")
	// ErrUserPasswordNotSet ...
	ErrUserPasswordNotSet = errors.New("user password not set")
	// ErrUsernameTaken ...
	ErrUsernameTaken = errors.New("username taken")
)

//UserExists returns true if user exists
func UserExists(username string) bool {
	_, err := FindUserByEmail(username)
	return err == nil
}

// FindUserByUsername looks up a user by username
func FindUserByEmail(email string) (*models.User, error) {
	// Usernames are case insensitive
	user := models.GetUserByFilter(bson.D{{"email", email}})
	// Not found
	if user.ID.IsZero() {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// CreateUser saves a new user to database
func CreateUser(roleID, username, password string) (*models.User, error) {
	return createUserCommon(roleID, username, password)
}

//SetPassword sets a user password
func SetPassword(user *models.User, password string) error {
	return setPasswordCommon(user, password)
}

// AuthUser authenticates user
func AuthUser(username, password string) (*models.User, error) {
	// Fetch the user
	user, err := FindUserByEmail(username)
	if err != nil {
		return nil, err
	}

	// Check that the password is set
	if user.Password == "" {
		return nil, ErrUserPasswordNotSet
	}

	// Verify the password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, ErrInvalidUserPassword
	}

	return user, nil
}

// UpdateUsername ...
func UpdateUsername(user *models.User, username string) error {
	if username == "" {
		return ErrCannotSetEmptyUsername
	}

	return updateUsernameCommon(user, username)
}

func createUserCommon(roleID, username, password string) (*models.User, error) {
	// Start with a user without a password
	user := &models.User{
		CreatedAt: time.Now().UTC(),
		Roles:     []string{roleID},
		Email:     strings.ToLower(username),
		Password:  "",
	}

	// If the password is being set already, create a bcrypt hash
	if password != "" {
		if len(password) < MinPasswordLength {
			return nil, ErrPasswordTooShort
		}
		user.Password = utils.HashPassword(password)
	}

	// Check the username is available
	if UserExists(user.Email) {
		return nil, ErrUsernameTaken
	}

	// Create the user
	user, err := models.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func setPasswordCommon(user *models.User, password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	// Create a bcrypt hash
	passwordHash := utils.HashPassword(password)
	// Set the password on the user object
	user.Password = passwordHash
	user, err := models.UpdateUser(user)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func updateUsernameCommon(user *models.User, username string) error {
	if username == "" {
		return ErrCannotSetEmptyUsername
	}
	return nil
	///return db.Model(user).UpdateColumn("username", strings.ToLower(username)).Error
}
