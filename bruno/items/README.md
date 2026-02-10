# Items API Collection

This Bruno collection contains API tests for the lost-and-found items endpoints in the Nyx backend. It covers item creation, retrieval, updates, deletion, and status management for both LOST and FOUND items.

## 📁 Files Overview

| File | Purpose | Method | Auth Required |
|------|---------|---------|---------------|
| `fetch-all-items.bru` | Get all items (with optional type filter) | GET | No |
| `create-item.bru` | Create new item (LOST or FOUND) | POST | Yes |
| `fetch-item-by-id.bru` | Get item by ID | GET | No |
| `fetch-all-items-by-user-id.bru` | Get current user's items | GET | Yes |
| `update-item-by-id.bru` | Update item details | PATCH | Yes (Owner only) |
| `delete-item-by-id.bru` | Delete item (soft delete) | DELETE | Yes (Owner only) |
| `update-item-status-by-id.bru` | Update item status | PATCH | Yes (Admin/Owner) |

## 🚀 Quick Start

### 1. Setup Environment
- Set `BASE_URL` to `http://localhost:8080/api/v1` in environment files
- Ensure the backend server is running: `task dev`
- Ensure database is running: `task docker:up`

### 2. Test Public Endpoints
1. Run `fetch-all-items.bru` - Get all items or filter by type
2. Run `fetch-item-by-id.bru` - Get specific item details

### 3. Test Authenticated Endpoints
1. First authenticate using auth collection (`login.bru`)
2. Run `create-item.bru` - Create LOST or FOUND item
3. Copy the returned item ID for subsequent tests
4. Run `fetch-all-items-by-user-id.bru` - View your items
5. Run `update-item-by-id.bru` - Update item details
6. Run `update-item-status-by-id.bru` - Change item status
7. Run `delete-item-by-id.bru` - Soft delete item

## Item Types and Rules

### LOST Items
- **Purpose**: Report lost items
- **Hub**: Not required (must be null)
- **Anonymous**: Can be true or false
- **Visibility**: Publicly visible to help finders locate owners

### FOUND Items
- **Purpose**: Report found items
- **Hub**: Required (must specify which hub has the item)
- **Anonymous**: Can be true or false
- **Visibility**: Publicly visible to help owners find their items

## Field Validation Rules

### Required Fields (Create)
- **name**: 3-100 characters, item title/name
- **description**: 10-1000 characters, detailed description
- **type**: Must be "LOST" or "FOUND"
- **is_anonymous**: Boolean, hide reporter identity if true

### Optional Fields (All Operations)
- **location**: 5-200 characters, location description
- **time_at**: RFC3339 timestamp, when item was lost/found
- **latitude**: String, GPS latitude coordinate
- **longitude**: String, GPS longitude coordinate
- **hub_id**: UUID (required for FOUND items, null for LOST)

## Item Status Values

- **OPEN**: Item is active and visible to public
- **PENDING_CLAIM**: Item has a claim pending review
- **RESOLVED**: Item has been returned to owner
- **ARCHIVED**: Item is no longer active (hidden from public)

## Authentication & Authorization

### Public Endpoints
- `GET /items/` - Browse all items with optional type filter
- `GET /items/:id` - View specific item details

### User Endpoints
- `POST /items/` - Create items (requires authentication)
- `GET /items/me` - View your own items
- `PATCH /items/:id` - Update your own items
- `DELETE /items/:id` - Delete your own items

### Admin Endpoints
- `PATCH /items/:id/status` - Update any item status

## Query Parameters

### Filter by Type
```bash
GET /items/?type=LOST     # Get only lost items
GET /items/?type=FOUND    # Get only found items
GET /items/               # Get all items (both types)
```

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
1. **Create LOST Item**: Report a lost item without hub
2. **Create FOUND Item**: Report a found item with hub
3. **Browse Items**: Filter by type, view details
4. **Manage Items**: Update, change status, delete
5. **User Items**: View items you created

### Error Testing
1. **Invalid Item Type**: Test with type other than LOST/FOUND
2. **Missing Hub**: Create FOUND item without hub_id
3. **Unauthorized**: Update/delete other user's items
4. **Invalid UUID**: Test with malformed item IDs

## Important Notes

### Item Ownership
- Users can only update/delete their own items
- Admins can update status of any item
- Anonymous items hide reporter identity from public

### Hub Requirements
- FOUND items must specify a valid hub_id
- LOST items must not have a hub_id (null)
- Hub must exist before creating FOUND item

### Soft Delete
- Items are soft deleted (marked as deleted, not physically removed)
- Deleted items are no longer visible in public listings
- Item history is preserved for audit purposes

### Image Handling
- Items support image uploads (separate endpoint)
- Image URLs are redacted based on visibility rules
- Public items show appropriate image details

### 📋 UUID Testing Notes

The Bruno files use placeholder UUIDs. To test with real data:

1. **Create an item first** using `create-item.bru`
2. **Copy the item ID** from the response (e.g., `"id": "019bbcd3-d654-76bf-98dc-121f7487b2e1"`)
3. **Replace placeholder UUID** in `fetch-item-by-id.bru`, `update-item-by-id.bru`, `delete-item-by-id.bru`, and `update-item-status-by-id.bru`

The placeholder UUIDs in the files are:
- `019bbcd3-d654-76bf-98dc-121f7487b2e1` (update item)
- `019bbcd3-00ca-7f9a-a642-ddbe4def995f` (delete item)
- `019bbd0a-b3fd-7187-9a30-2b8376540304` (update status)
- `019bbc84-9648-79f9-8984-22c437280952` (fetch by ID)