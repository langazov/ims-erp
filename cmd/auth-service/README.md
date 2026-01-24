# Auth Service

Authentication and authorization service for the ERP system.

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register a new user |
| POST | `/api/v1/auth/login` | Login and get tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout and invalidate tokens |
| POST | `/api/v1/auth/forgot-password` | Request password reset |
| POST | `/api/v1/auth/reset-password` | Reset password with token |
| GET | `/api/v1/auth/me` | Get current user info |
| PUT | `/api/v1/auth/me` | Update current user |
| POST | `/api/v1/auth/change-password` | Change password |

### RBAC

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/roles` | List all roles |
| POST | `/api/v1/roles` | Create a role |
| GET | `/api/v1/roles/:id` | Get role by ID |
| PUT | `/api/v1/roles/:id` | Update role |
| DELETE | `/api/v1/roles/:id` | Delete role |
| GET | `/api/v1/permissions` | List all permissions |
| POST | `/api/v1/roles/:id/permissions` | Assign permissions to role |
| DELETE | `/api/v1/roles/:id/permissions` | Remove permissions from role |

## Authentication Flow

```json
// Register
POST /api/v1/auth/register
{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "User Name",
  "role": "user"
}

// Login
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "securepassword"
}

// Response
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 3600,
  "tokenType": "Bearer"
}
```

## Running

```bash
go run cmd/auth-service/main.go
```

## Testing

```bash
make test
```
