package pkg_test

import (
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/stretchr/testify/assert"
)

func TestGenerateOTP_ValidFormat_ReturnsOTP(t *testing.T) {
	otp, otpSlice, err := pkg.GenerateOTP()

	assert.NoError(t, err)
	assert.NotEmpty(t, otp)
	assert.Len(t, otp, 6)
	assert.Len(t, otpSlice, 6)
	assert.Equal(t, otpSlice[0]+otpSlice[1]+otpSlice[2]+otpSlice[3]+otpSlice[4]+otpSlice[5], otp)
}

func TestGenerateOTP_NumericOnly_ReturnsOnlyDigits(t *testing.T) {
	otp, _, err := pkg.GenerateOTP()

	assert.NoError(t, err)
	for _, char := range otp {
		assert.True(t, char >= '0' && char <= '9', "OTP should contain only digits, got: %c", char)
	}
}

func TestGenerateOTP_Range_Within100000To999999(t *testing.T) {
	otp, _, err := pkg.GenerateOTP()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, otp, "100000")
	assert.LessOrEqual(t, otp, "999999")
}

func TestGenerateOTP_Uniqueness_DifferentOTPs(t *testing.T) {
	otps := make(map[string]bool)

	// Generate multiple OTPs to check for uniqueness
	for i := 0; i < 100; i++ {
		otp, _, err := pkg.GenerateOTP()
		assert.NoError(t, err)

		// Check if this OTP was already generated
		if exists := otps[otp]; exists {
			t.Logf("Duplicate OTP found: %s at iteration %d", otp, i)
			// Note: This is not necessarily an error since random numbers can collide
			// but should be very rare for 100 iterations of 6-digit numbers
		}
		otps[otp] = true
	}

	// We should have generated 100 (or slightly fewer if collisions occurred) unique OTPs
	assert.Greater(t, len(otps), 95, "Should have at least 95 unique OTPs out of 100 generations")
}

func TestGenerateOTP_SliceConsistency_MatchesString(t *testing.T) {
	otp, otpSlice, err := pkg.GenerateOTP()

	assert.NoError(t, err)
	assert.Len(t, otpSlice, 6)

	// Reconstruct string from slice
	reconstructed := ""
	for _, digit := range otpSlice {
		reconstructed += digit
	}

	assert.Equal(t, otp, reconstructed, "OTP string should match slice reconstruction")
}

func TestGenerateOTP_MultipleCalls_DifferentResults(t *testing.T) {
	generatedOTPs := make([]string, 0, 10)

	// Generate 10 OTPs
	for i := 0; i < 10; i++ {
		otp, _, err := pkg.GenerateOTP()
		assert.NoError(t, err)
		generatedOTPs = append(generatedOTPs, otp)
	}

	// Check that not all OTPs are the same (extremely unlikely with random generation)
	firstOTP := generatedOTPs[0]
	allSame := true
	for _, otp := range generatedOTPs[1:] {
		if otp != firstOTP {
			allSame = false
			break
		}
	}

	assert.False(t, allSame, "Generated OTPs should not all be identical")
}

func TestGenerateOTP_EdgeCases_RapidGeneration(t *testing.T) {
	// Test rapid generation to ensure no timing issues
	for i := 0; i < 1000; i++ {
		otp, otpSlice, err := pkg.GenerateOTP()

		assert.NoError(t, err)
		assert.Len(t, otp, 6)
		assert.Len(t, otpSlice, 6)
		assert.NotEmpty(t, otp)

		// Verify it's a valid 6-digit number
		for _, char := range otp {
			assert.True(t, char >= '0' && char <= '9')
		}
	}
}

func TestGenerateOTP_OutputStructure_ConsistentFormat(t *testing.T) {
	// Test the structure of outputs
	for i := 0; i < 10; i++ {
		otp, otpSlice, err := pkg.GenerateOTP()

		assert.NoError(t, err)

		// Check string format
		assert.IsType(t, "", otp)
		assert.Len(t, otp, 6)

		// Check slice format
		assert.IsType(t, []string{}, otpSlice)
		assert.Len(t, otpSlice, 6)

		// Check each element in slice
		for _, digit := range otpSlice {
			assert.IsType(t, "", digit)
			assert.Len(t, digit, 1)
			assert.True(t, digit[0] >= '0' && digit[0] <= '9')
		}
	}
}

func TestGenerateOTP_NoErrors_CleanGeneration(t *testing.T) {
	// Test that no errors occur over many generations
	for i := 0; i < 500; i++ {
		otp, otpSlice, err := pkg.GenerateOTP()

		assert.NoError(t, err, "GenerateOTP should not return error at iteration %d", i)
		assert.NotEmpty(t, otp, "OTP should not be empty at iteration %d", i)
		assert.NotEmpty(t, otpSlice, "OTP slice should not be empty at iteration %d", i)
	}
}
