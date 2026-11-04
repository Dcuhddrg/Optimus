package utils

import "golang.org/x/crypto/bcrypt"

// BcryptHash hashes a password with bcrypt, as GVA's utils.BcryptHash.
func BcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// BcryptCheck reports whether password matches hash, as GVA's utils.BcryptCheck.
func BcryptCheck(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
