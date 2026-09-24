package utils

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Fatal("Could not hash password!")
		return "", err
	}

	return string(hashedPassword), nil
}

func ValidatePassword(userStoredPassword, providedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(userStoredPassword), []byte(providedPassword))

	return err == nil
}
