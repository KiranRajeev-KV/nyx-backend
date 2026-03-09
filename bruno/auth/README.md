# Auth API Collection

This Bruno collection contains API tests for the authentication endpoints in the Nyx backend. It covers user registration, login, OTP verification, password management, and session management.

## 📁 Files Overview

| File | Purpose | Method | Auth Required |
|------|---------|---------|---------------|
| `register.bru` | User registration | POST | No |
| `verify-otp.bru` | Email verification with OTP | POST | No |
| `resend-otp.bru` | Resend OTP for email verification | POST | No |
| `login.bru` | User authentication | POST | No |
| `logout.bru` | User logout | GET | Yes |
| `refresh.bru` | Refresh access token | POST | Yes |
| `session.bru` | Check current session status | GET | Yes |
| `forgot-password.bru` | Initiate password reset | POST | No |
| `reset-password.bru` | Reset password with OTP | POST | No |

## 🚀 Quick Start

### 1. Setup Environment
- Set `BASE_URL` to `http://localhost:8080/api/v2` in environment files
- Ensure the backend server is running: `task dev`
- Ensure database is running: `task docker:up`

### 2. Complete Registration Flow
1. Run `register.bru` - Create a new user account
2. Run `verify-otp.bru` - Verify email with OTP (check logs/console for OTP)
3. If OTP expired, run `resend-otp.bru` to get a new OTP

### 3. Authentication Flow
1. Run `login.bru` - Authenticate with email and password
2. Run `session.bru` - Verify session is active
3. Run `refresh.bru` - Refresh access token when needed
4. Run `logout.bru` - End the session

### 4. Password Reset Flow
1. Run `forgot-password.bru` - Request password reset
2. Run `reset-password.bru` - Reset password with OTP (check logs for OTP)

## Field Validation Rules

### Registration (`register.bru`)
- **name**: 3-100 characters, required
- **email**: Valid email format, required
- **password**: 8-50 characters, required, must contain uppercase, lowercase, and numbers

### Login (`login.bru`)
- **email**: Valid email format, required
- **password**: 8-50 characters, required

### OTP Operations (`verify-otp.bru`, `reset-password.bru`)
- **otp**: 6 digits, required
- **password**: Only for reset-password, 8-50 characters with complexity requirements

### Password Reset Initiation (`forgot-password.bru`)
- **email**: Valid email format, required

## Authentication Flow

### Token Management
- **Access Token**: Short-lived token for API authentication
- **Refresh Token**: Long-lived token for refreshing access tokens
- Tokens are stored as HTTP-only cookies for security

### Session Management
- `GET /auth/session` - Returns current user session info
- `GET /auth/logout` - Invalidates session and clears cookies
- `POST /auth/refresh` - Refreshes access token using refresh token

## OTP System

### OTP Generation
- 6-digit numeric codes
- OTPs are logged to console during development
- OTPs expire after a configured time period

### OTP Usage
- **Email Verification**: Required after registration before login
- **Password Reset**: Required to reset forgotten passwords
- **Resend Option**: Users can request new OTP if expired

## Test Users

### Pre-configured Admin User
- **Email**: `admin@example.com`
- **Password**: `password123`
- Use this user for testing admin-only endpoints

### Test User Creation
Use `register.bru` with these test credentials:
- **Name**: `John Doe`
- **Email**: `john.doe@example.com`
- **Password**: `StrongPassword123`

## Response Format

### Success Responses
```json
{
  "message": "Operation completed successfully",
  "data": {
    // Response data varies by endpoint
  }
}
```

### Error Responses
```json
{
  "message": "Descriptive error message"
}
```

## Testing Scenarios

### Happy Path Testing
1. **Complete Registration**: Register → Verify OTP → Login → Check Session
2. **Password Reset**: Forgot Password → Reset with OTP → Login with new password
3. **Session Management**: Login → Refresh → Logout

### Error Testing
1. **Invalid Credentials**: Test login with wrong password
2. **Expired OTP**: Test with old OTP after expiration
3. **Unverified Email**: Attempt login without email verification
4. **Invalid Token**: Test endpoints with invalid/expired tokens

## Important Notes

### Security Considerations
- All passwords must meet complexity requirements
- OTPs are 6-digit numbers for development (check console logs)
- Tokens are stored as HTTP-only cookies
- Session tokens automatically expire

### Development Tips
- OTP codes are printed in the backend console during development
- Use the pre-configured admin user for quick testing
- Clear browser cookies or use incognito mode for clean testing
- Database seeding (`task db:seed`) may create additional test users

### Cookie Management
- Authentication uses HTTP-only cookies for security
- Refresh tokens have longer expiration than access tokens
- Logout clears all authentication cookies

### 📋 OTP Development Notes

During development, OTP codes are logged to the console. To find the OTP:

1. **Check Backend Logs**: Look for 6-digit numbers in the console output
2. **Email Verification**: OTP appears after registration
3. **Password Reset**: OTP appears after forgot-password request

The OTP format in the Bruno files uses placeholder values like `"412193"` - replace these with actual OTPs from logs during testing.