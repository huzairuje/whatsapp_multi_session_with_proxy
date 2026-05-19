# Authentication Implementation Summary

## Backend Changes

### New Auth Package (`backend/auth/`)
- **model.go**: User model, LoginRequest, LoginResponse, ChangePasswordRequest
- **service.go**: Authentication service with JWT token generation/validation, password hashing with bcrypt
- **repository.go**: Database layer for users table (SQLite and PostgreSQL support)
- **middleware.go**: JWT authentication middleware for protecting routes

### Database Updates
- **database/sqlite.go**: Added `GetRawSqliteDB()` function to get raw SQL connection for auth
- **database/postgresql.go**: Added `GetRawPostgresDB()` function for PostgreSQL

### Boot/Setup Updates
- **boot/setup.go**: 
  - Initialize auth service on startup
  - Create users table automatically
  - Generate random admin password on first run (logged to console)
  - Pass auth service to handler and router

### Handler Updates
- **handler/handler.go**: Updated to include AuthService
- **handler/auth_handler.go**: New file with login and change-password endpoints

### Router Updates
- **routers/routers.go**: 
  - Public routes: `/login`, `/health-check`
  - Protected routes: All other endpoints require Bearer token
  - Auth middleware applied to protected routes

### Dependencies Added
- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `golang.org/x/crypto` - Password hashing with bcrypt

## Frontend Changes

### New Auth Infrastructure
- **contexts/AuthContext.tsx**: 
  - Auth context provider with login/logout/changePassword functions
  - Token and username stored in localStorage
  - useAuth hook for accessing auth state

- **pages/Login.tsx**: 
  - Login page with username/password form
  - Error handling
  - Default credentials info display

- **components/ProtectedRoute.tsx**: 
  - Route wrapper that redirects to login if not authenticated

### Updated Files
- **main.tsx**: Wrapped app with AuthProvider
- **App.tsx**: Added login route and protected routes
- **services/api.ts**: 
  - Added auth token interceptor to all requests
  - Added 401 response handler to redirect to login
- **components/layout/Navbar.tsx**: 
  - Display current username
  - Added logout button

## How It Works

### First Startup
1. Backend initializes and creates `users` table
2. Admin user created with randomized password
3. Password printed to console logs (save this!)
4. Example: `Password: aBcDeFgHiJkLmNoP`

### Login Flow
1. User navigates to `/login`
2. Enters username (admin) and password
3. Frontend sends POST to `/api/login`
4. Backend validates credentials and returns JWT token
5. Token stored in localStorage
6. User redirected to dashboard

### Protected Routes
1. All API requests include `Authorization: Bearer {token}` header
2. Backend middleware validates token
3. Invalid/expired tokens trigger 401 response
4. Frontend redirects to login on 401

### Logout
1. User clicks logout button
2. Token removed from localStorage
3. User redirected to login page

## Configuration

### JWT Secret
Currently set to `"your-secret-key-change-this"` in `boot/setup.go`
**IMPORTANT**: Change this to a secure random string in production!

### Password Requirements
- Minimum 8 characters recommended
- Hashed with bcrypt (cost factor 10)
- Never stored in plain text

## Testing

### Backend
```bash
cd backend
go mod tidy
go run main.go
# Check logs for admin password
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### API Testing
```bash
# Login
curl -X POST http://localhost:1234/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"YOUR_PASSWORD"}'

# Use returned token for other requests
curl -X GET http://localhost:1234/api/devices \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Security Notes

1. **JWT Secret**: Change in production
2. **HTTPS**: Use HTTPS in production
3. **Token Expiry**: Currently 24 hours, adjust as needed
4. **Password Policy**: Consider adding password strength requirements
5. **Rate Limiting**: Consider adding login attempt rate limiting
6. **CORS**: Configure CORS properly for production

## Next Steps (Optional)

1. Add password strength validation
2. Implement refresh tokens
3. Add user management (create/delete users)
4. Add role-based access control (RBAC)
5. Add login attempt rate limiting
6. Add audit logging for auth events
7. Add 2FA support
