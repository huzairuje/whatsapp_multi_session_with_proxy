# Debug Testing Guide - Session Logout Issue

## What I Added
Debug logging to track:
1. ✅ Login process and token storage
2. ✅ API requests and token presence
3. ✅ 401 errors and refresh attempts
4. ✅ Token refresh success/failure

## How to Test

### Step 1: Start Backend
```bash
cd backend
go run main.go
# Note the admin password from logs
```

### Step 2: Start Frontend
```bash
cd frontend
npm run dev
```

### Step 3: Open Browser DevTools
- Press `F12` or right-click → Inspect
- Go to **Console** tab
- Keep it open while testing

### Step 4: Login
1. Enter username: `admin`
2. Enter password: (from backend logs)
3. Click "Sign In"
4. **Watch the console** for logs like:
   ```
   [Login] Attempting login with username: admin
   [Login] Login successful
   [Login] Stored tokens: { token: "eyJ...", refreshToken: "eyJ...", username: "admin" }
   ```

### Step 5: Click "Manage Sessions"
1. Click the button
2. **Watch the console** for logs like:
   ```
   [API] Request to: /api/devices Token present: true
   [API] Request to: /api/bulk-send-status Token present: true
   ```

### Step 6: Check Network Tab
1. Go to **Network** tab in DevTools
2. Click "Manage Sessions" again
3. Look for requests to `/api/devices`
4. Click on the request
5. Check **Headers** section:
   - Should see: `Authorization: Bearer eyJ...`
6. Check **Response** section:
   - Should see device data, not error

## What to Look For

### ✅ Success Scenario
```
[Login] Login successful
[Login] Stored tokens: { token: "...", refreshToken: "...", username: "admin" }
[API] Request to: /api/devices Token present: true
[API] Request to: /api/bulk-send-status Token present: true
(page loads with data)
```

### ❌ Failure Scenario 1: Token Not Stored
```
[Login] Login successful
[Login] Stored tokens: { token: undefined, refreshToken: undefined, username: undefined }
[API] Request to: /api/devices Token present: false
[API] 401 Unauthorized on: /api/devices
[API] No refresh token, redirecting to login
```
**Fix**: AuthContext not storing tokens properly

### ❌ Failure Scenario 2: Token Not Sent
```
[Login] Login successful
[Login] Stored tokens: { token: "...", refreshToken: "...", username: "admin" }
[API] Request to: /api/devices Token present: false
[API] 401 Unauthorized on: /api/devices
```
**Fix**: Request interceptor not adding Authorization header

### ❌ Failure Scenario 3: Refresh Fails
```
[API] 401 Unauthorized on: /api/devices
[API] Refresh token present: true
[API] Attempting token refresh...
[API] Token refresh failed: Error: ...
[API] Redirecting to login due to refresh failure
```
**Fix**: Refresh endpoint not working

## Report Back With

Please share:
1. **Console logs** when you click "Manage Sessions"
2. **Network tab** - what's the response status for `/api/devices`?
3. **LocalStorage** - run in console:
   ```javascript
   console.log(localStorage)
   ```

This will help me identify exactly where the logout is happening!

## Quick Fixes to Try

### If tokens not stored:
Clear browser cache and localStorage:
```javascript
localStorage.clear()
// Then login again
```

### If Authorization header missing:
Check if browser is blocking it (unlikely but possible)

### If refresh fails:
Check backend logs for errors on `/refresh-token` endpoint

Let me know what you see in the console!
