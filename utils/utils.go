/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package utils

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

//HashPassword hash a given password string
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Errorln(err)
	}
	return string(bytes)
}

//CheckPasswordHash compare a given hashed password and a plain string
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RandomIDGen generates a random string identifier.
func RandomIDGen(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// RandomInt  generates a random integer identifier.
func RandomInt(length int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// PointerBool returns a pointer to a bool.
func PointerBool(b bool) *bool {
	return &b
}
