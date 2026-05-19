# Session Logout Issue - Fixed

## Problem
Clicking "Manage Session" or "Send Bulk Messages" buttons was causing logout/session loss.

## Root Cause
The API interceptor had a hardcoded refresh token endpoint URL:
```typescript
// WRONG - hardcoded URL
axios.post('http://localhost:1234/refresh-token', ...)
```

This bypassed the Vite proxy and failed to refresh tokens properly.

## Solution Applied
Updated to use the correct API path:
```typescript
// CORRECT - uses Vite proxy
axios.post('/api/refresh-token', ...)
```

## How It Works Now

### When You Click a Button
1. Frontend makes API call (e.g., `/api/devices`)
2. Request includes `Authorization: Bearer {token}`
3. If token is valid → request succeeds
4. If token expired → backend returns 401
5. Frontend automatically refreshes token via `/api/refresh-token`
6. Retries original request with new token
7. User stays logged in ✅

### Token Lifecycle
- **Login**: Get access token (2h) + refresh token (7d)
- **After 1h 55m**: Auto-refresh happens silently
- **On 401 error**: Manual refresh + retry
- **After 7 days**: Refresh token expires, user must login again

## Testing the Fix

### 1. Start Backend
```bash
cd backend
go run main.go
```

### 2. Start Frontend
```bash
cd frontend
npm run dev
```

### 3. Login
- Username: `admin`
- Password: `GNn0geM51w0NxmW6`

### 4. Test Buttons
- Click "Manage Session" → Should work ✅
- Click "Send Bulk Messages" → Should work ✅
- No logout should occur

### 5. Verify Token Refresh
Open browser DevTools → Application → LocalStorage:
- `auth_token` - Access token (changes every 2h)
- `auth_refresh_token` - Refresh token (7 days)
- `auth_username` - Username

## What Changed
- ✅ Fixed refresh token endpoint URL in API interceptor
- ✅ Now uses Vite proxy correctly
- ✅ Token refresh works automatically
- ✅ Session stays active during API calls

## If Still Having Issues

### Check 1: Verify Tokens in LocalStorage
```javascript
// In browser console
localStorage.getItem('auth_token')
localStorage.getItem('auth_refresh_token')
```

### Check 2: Check Network Tab
- Look for `/api/refresh-token` requests
- Should return 200 with new tokens
- Not 401 or 404

### Check 3: Check Backend Logs
```
[GIN] POST /refresh-token 200
```

### Check 4: Clear Cache and Retry
```bash
# Clear browser cache
# Delete localStorage
# Logout and login again
```
