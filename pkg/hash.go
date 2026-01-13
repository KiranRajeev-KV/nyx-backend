package pkg

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	hashedPassword := string(hashed)
	return hashedPassword, nil
}

func CompareHash(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
