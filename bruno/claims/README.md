# Claims API Collection

This Bruno collection contains API tests for the claims management endpoints in the Nyx backend. It covers claim creation, retrieval, and processing for the lost-and-found system.

## 📁 Files Overview

| File | Purpose | Method | Auth Required |
|------|---------|---------|---------------|
| `create-claim.bru` | Submit a claim for a found item | POST | Yes |
| `get-all-claims-by-me.bru` | Get current user's claims | GET | Yes |
| `get-claims-by-item-id.bru` | Get all claims for a specific item | GET | No |

## 🚀 Quick Start

### 1. Setup Environment
- Set `BASE_URL` to `http://localhost:8080/api/v2` in environment files
- Ensure the backend server is running: `task dev`
- Ensure database is running: `task docker:up`

### 2. Test Claims Workflow
1. First authenticate using auth collection (`login.bru`)
2. Create a FOUND item using items collection (`create-item.bru`)
3. Copy the returned item ID for claim testing
4. Run `create-claim.bru` - Submit a claim for the item
5. Run `get-all-claims-by-me.bru` - View your claims
6. Run `get-claims-by-item-id.bru` - View claims for an item

## Claim System Overview

### Claim Types and Rules
- **Purpose**: Allow users to claim ownership of FOUND items
- **Eligibility**: Only FOUND items with OPEN or PENDING_CLAIM status can be claimed
- **Restrictions**: Users cannot claim their own items
- **Uniqueness**: Each user can only claim an item once

### Claim Status Values
- **PENDING**: Claim submitted and awaiting admin review
- **APPROVED**: Claim approved by admin, item marked as RESOLVED
- **REJECTED**: Claim rejected by admin

### Item Status Workflow
- **OPEN** → **PENDING_CLAIM**: When first claim is submitted
- **PENDING_CLAIM** → **OPEN**: When all claims are rejected
- **PENDING_CLAIM** → **RESOLVED**: When any claim is approved

## Field Validation Rules

### Required Fields (Create Claim)
- **item_id** (string, UUID): ID of the FOUND item to claim
- **proof_text** (string, 10-1000 chars): Detailed proof of ownership description

### Optional Fields (Create Claim)
- **proof_image_url** (string, URL): URL to proof image or document

## Business Rules

### Claim Eligibility
- Only FOUND items can be claimed (not LOST items)
- Items must have OPEN or PENDING_CLAIM status
- Users cannot claim their own items
- Each user can only submit one claim per item
- Rejected users cannot claim the same item again

### Claim Processing
- Claims require admin approval/rejection
- Approved claims resolve the item (status becomes RESOLVED)
- Rejected claims may keep item open if other claims exist
- Admin processing updates both claim and item statuses

## Authentication & Authorization

### User Endpoints
- `POST /claims/` - Submit claim (requires authentication)
- `GET /claims/me` - View your own claims (requires authentication)

### Public Endpoints
- `GET /claims/item/:id` - View claims for an item (no authentication required)

### Admin Endpoints
- Admin functionality exists but requires separate admin endpoints (not in this collection)

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
1. **Create FOUND Item**: Report a found item with hub
2. **Submit Claim**: Claim the item with proof
3. **View Claims**: Check your claims and item claims
4. **Multiple Claims**: Test multiple users claiming same item

### Error Testing
1. **Invalid Item**: Claim non-existent item
2. **Wrong Type**: Attempt to claim a LOST item
3. **Own Item**: Try to claim your own item
4. **Duplicate Claim**: Submit multiple claims for same item
5. **Invalid Status**: Claim already resolved item

## Important Notes

### Claim Workflow
- Claims automatically update item status to PENDING_CLAIM
- Claim processing affects item availability
- Multiple claims can exist for the same item
- First approved claim resolves the item

### Proof Requirements
- Text proof is required and must be detailed
- Image proof is optional but recommended
- Proof helps admins validate ownership claims

### Status Management
- Claims start with PENDING status
- Admin approval sets item to RESOLVED
- Admin rejection may return item to OPEN status
- Rejected users cannot claim again

### 📋 UUID Testing Notes

The Bruno files use placeholder UUIDs. To test with real data:

1. **Create a FOUND item first** using items collection (`create-item.bru`)
2. **Copy the item ID** from the response (e.g., `"id": "019c239a-8423-70e2-a7fd-e60a7da60a77"`)
3. **Replace placeholder UUID** in `create-claim.bru` and `get-claims-by-item-id.bru`

The placeholder UUIDs in the files are:
- `019c239a-8423-70e2-a7fd-e60a7da60a77` (item ID for claim testing)

### Missing Admin Endpoints

This collection includes user-facing claim endpoints. Admin endpoints for claim processing (`PATCH /claims/:id`) exist in the API but may require separate admin authentication setup and are not included in this user collection.