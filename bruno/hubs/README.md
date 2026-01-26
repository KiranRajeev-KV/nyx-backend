# Hubs API Collection

This Bruno collection contains API tests for the hubs CRUD endpoints in the Nyx backend.

## 📁 Files Overview

| File | Purpose | Method | Auth Required |
|------|---------|---------|---------------|
| `fetch-all-hubs.bru` | Get all hubs | GET | No |
| `fetch-hub-by-id.bru` | Get hub by ID | GET | No |
| `create-hub.bru` | Create new hub | POST | Yes (Admin) |
| `update-hub.bru` | Update existing hub | PATCH | Yes (Admin) |
| `delete-hub.bru` | Delete hub | DELETE | Yes (Admin) |

## 🚀 Quick Start

### 1. Setup Environment
- Set `BASE_URL` to `http://localhost:8080/api/v1` in `nyx-env.bru`
- Ensure the backend server is running: `task dev`
- Ensure database is running: `task docker:up`

### 2. Test Public Endpoints
1. Run `fetch-all-hubs.bru` - Should return list of hubs
2. Run `fetch-hub-by-id.bru` - Replace placeholder UUID with a valid hub ID

### 3. Test Admin Endpoints (Auth Required)
1. First authenticate using auth collection (`login.bru` as admin)
2. Run `create-hub.bru` - Create a test hub
3. Copy the returned hub ID for subsequent tests
4. Run `fetch-hub-by-id.bru` - Test with the new hub ID
5. Run `update-hub.bru` - Update the hub
6. Run `delete-hub.bru` - Clean up the test hub

## Field Validation Rules

### Required Fields (Create)
- **name**: 3-100 characters
- **address**: 5-200 characters  
- **contact**: 5-50 characters

### Optional Fields (All Operations)
- **longitude**: 1-50 characters (no format validation)
- **latitude**: 1-50 characters (no format validation)

## Authentication

- **Public endpoints**: `GET /hubs/` and `GET /hubs/:id`
- **Admin endpoints**: All other operations require valid admin token
- Use the auth collection to get authentication tokens first

## Important Notes

- **Delete Protection**: Hubs cannot be deleted if they have associated items
- **Partial Updates**: Use `PATCH` with only the fields you want to update
- **Response Format**: All responses follow consistent JSON format with `message` and optional `data` fields

### 📋 UUID Testing Notes

The Bruno files use placeholder UUIDs. To test with real data:

1. **Create a hub first** using `create-hub.bru`
2. **Copy the hub ID** from the response (e.g., `"id": "019bbc79-208b-7105-95ae-d29aa8f6ecd8"`)
3. **Replace placeholder UUID** in `fetch-hub-by-id.bru`, `update-hub.bru`, and `delete-hub.bru`

The placeholder UUIDs in the files are: `019bbc79-208b-7105-95ae-d29aa8f6ecd8`