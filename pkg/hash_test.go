package pkg_test

import (
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/stretchr/testify/assert"
)

func TestHash_ValidPassword_ReturnsHash(t *testing.T) {
	password := "testPassword123"

	hash, err := pkg.Hash(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	assert.GreaterOrEqual(t, len(hash), 60) // bcrypt hashes are typically 60 chars
}

func TestHash_EmptyPassword_ReturnsHash(t *testing.T) {
	password := ""

	hash, err := pkg.Hash(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestHash_VariousPasswords_ReturnsDifferentHashes(t *testing.T) {
	password1 := "password123"
	password2 := "password456"
	password3 := "password123" // Same as password1

	hash1, err1 := pkg.Hash(password1)
	hash2, err2 := pkg.Hash(password2)
	hash3, err3 := pkg.Hash(password3)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	assert.NotEqual(t, hash1, hash2) // Different passwords should have different hashes
	assert.NotEqual(t, hash1, hash3) // Same password should have different hashes (due to salt)
}

func TestCompareHash_ValidMatch_ReturnsNoError(t *testing.T) {
	password := "testPassword123"
	hash, _ := pkg.Hash(password)

	err := pkg.CompareHash(hash, password)

	assert.NoError(t, err)
}

func TestCompareHash_InvalidMatch_ReturnsError(t *testing.T) {
	password := "testPassword123"
	wrongPassword := "wrongPassword456"
	hash, _ := pkg.Hash(password)

	err := pkg.CompareHash(hash, wrongPassword)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hashedPassword is not the hash of the given password")
}

func TestCompareHash_InvalidHash_ReturnsError(t *testing.T) {
	password := "testPassword123"
	invalidHash := "invalid_hash_string"

	err := pkg.CompareHash(invalidHash, password)

	assert.Error(t, err)
}

func TestCompareHash_EmptyHash_ReturnsError(t *testing.T) {
	password := "testPassword123"

	err := pkg.CompareHash("", password)

	assert.Error(t, err)
}

func TestCompareHash_EmptyPassword_ReturnsError(t *testing.T) {
	password := "testPassword123"
	hash, _ := pkg.Hash(password)

	err := pkg.CompareHash(hash, "")

	assert.Error(t, err)
}

func TestHash_CompareHash_RoundTrip(t *testing.T) {
	testPasswords := []string{
		"simple",
		"complexPassword123!@#",
		"verylongpasswordthatexceedstypicallengthsandmaybeeventhrowsomeerrors",
		"🔒🔑🛡️",
		"password\nwith\tnewlines",
	}

	for _, password := range testPasswords {
		t.Run("password_"+password, func(t *testing.T) {
			hash, err := pkg.Hash(password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)

			err = pkg.CompareHash(hash, password)
			assert.NoError(t, err)

			err = pkg.CompareHash(hash, password+"wrong")
			assert.Error(t, err)
		})
	}
}
