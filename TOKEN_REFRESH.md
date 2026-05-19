# Token Refresh Implementation

## Overview
Implemented automatic token refresh mechanism to extend user sessions without requiring re-login.

## Token Expiry Times
- **Access Token**: 2 hours
- **Refresh Token**: 7 days
- **Auto-Refresh**: Triggered 5 minutes before access token expires

## Backend Changes

### Auth Service (`backend/auth/service.go`)
- Updated `generateToken()` to accept custom duration parameter
- Returns both token and expiry timestamp
- Added `RefreshToken()` method to generate new tokens using refresh token

### Auth Model (`backend/auth/model.go`)
- Updated `LoginResponse` to include:
  - `RefreshToken`: Long-lived token for refreshing access tokens
  - `ExpiresIn`: Unix timestamp of token expiry
- Added `RefreshTokenRequest` model

### Handler (`backend/handler/auth_handler.go`)
- Added `HandleRefreshToken()` endpoint
- Validates refresh token and returns new access/refresh tokens

### Router (`backend/routers/routers.go`)
- Added public route: `POST /refresh-token`
- No authentication required (uses refresh token instead)

## Frontend Changes

### Auth Context (`frontend/src/contexts/AuthContext.tsx`)
- Stores refresh token in localStorage
- Automatically schedules token refresh 5 minutes before expiry
- `refreshToken()` method to manually refresh tokens
- Cleanup on logout

### API Interceptor (`frontend/src/services/api.ts`)
- Detects 401 responses
- Automatically attempts token refresh
- Queues failed requests during refresh
- Retries original request with new token
- Falls back to login on refresh failure

## API Endpoints

### Login
```
POST /login
Request: {"username":"admin","password":"password"}
Response: {
  "token": "access_token",
  "refresh_token": "refresh_token",
  "username": "admin",
  "expires_in": 1234567890
}
```

### Refresh Token
```
POST /refresh-token
Request: {"refresh_token":"refresh_token"}
Response: {
  "token": "new_access_token",
  "refresh_token": "new_refresh_token",
  "username": "admin",
  "expires_in": 1234567890
}
```

## How It Works

### Initial Login
1. User logs in with credentials
2. Backend returns access token (2h) + refresh token (7d)
3. Frontend stores both tokens
4. Frontend schedules refresh for 1h 55m later

### Automatic Refresh
1. 5 minutes before access token expires, frontend automatically refreshes
2. Sends refresh token to `/refresh-token` endpoint
3. Backend validates and returns new tokens
4. Frontend updates stored tokens
5. Reschedules next refresh

### Manual API Call During Expiry
1. If access token expires during API call
2. API returns 401 Unauthorized
3. Frontend automatically refreshes token
4. Retries original request with new token
5. User doesn't notice the refresh

### Refresh Token Expiry
1. If refresh token expires (7 days)
2. User must log in again
3. New tokens are issued

## Security Features

- ✅ Refresh tokens stored in localStorage (same as access tokens)
- ✅ Automatic token refresh prevents session interruption
- ✅ Failed refresh redirects to login
- ✅ Request queue prevents race conditions during refresh
- ✅ Tokens expire automatically
- ✅ No sensitive data in tokens (JWT claims only contain username)

## Configuration

### Token Durations
Edit `backend/auth/service.go` in `Login()` method:
```go
token, expiresIn, err := s.generateToken(user.Username, 2*time.Hour)  // Access token
refreshToken, _, err := s.generateToken(user.Username, 7*24*time.Hour) // Refresh token
```

### Refresh Timing
Edit `frontend/src/contexts/AuthContext.tsx` in `scheduleTokenRefresh()`:
```typescript
refreshTimeoutRef.current = setTimeout(() => {
  refreshTokenAsync()
}, 2 * 60 * 60 * 1000 - 5 * 60 * 1000)  // 2h - 5m
```

## Testing

### Test Token Refresh
```bash
# 1. Login
curl -X POST http://localhost:1234/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"PASSWORD"}'

# 2. Use refresh token to get new access token
curl -X POST http://localhost:1234/refresh-token \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"REFRESH_TOKEN"}'

# 3. Use new access token
curl -X GET http://localhost:1234/devices \
  -H "Authorization: Bearer NEW_ACCESS_TOKEN"
```

## Benefits

- ✅ Users stay logged in for up to 7 days
- ✅ Automatic token refresh prevents session interruption
- ✅ Seamless user experience
- ✅ Improved security with short-lived access tokens
- ✅ Long-lived refresh tokens for convenience
