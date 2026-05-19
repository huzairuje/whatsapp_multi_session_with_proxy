# Complete Troubleshooting Guide - Session Logout Issue

## The Problem
Clicking dashboard buttons causes logout/redirect to login page.

## Root Cause Analysis

### Possible Causes (in order of likelihood)

1. **Token not being sent with requests**
   - API interceptor not working
   - Token not stored in localStorage

2. **API returning 401 (Unauthorized)**
   - Token expired
   - Token invalid
   - Backend not recognizing token

3. **Token refresh failing**
   - Refresh endpoint not accessible
   - Refresh token invalid
   - CORS issue

4. **Frontend routing issue**
   - Page reload losing auth context
   - Navigation clearing localStorage

## Step-by-Step Diagnostic

### Phase 1: Verify Backend is Running

```bash
# In terminal, check if backend is running
curl http://localhost:1234/health-check

# Should return:
# {"message":"server is alive and ok!"}
```

### Phase 2: Test Login Endpoint

```bash
curl -X POST http://localhost:1234/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"GNn0geM51w0NxmW6"}'

# Should return tokens like:
# {"token":"eyJ...","refresh_token":"eyJ...","username":"admin","expires_in":1234567890}
```

### Phase 3: Test Protected Endpoint

```bash
# Get token from login response above
TOKEN="eyJ..."

curl -X GET http://localhost:1234/devices \
  -H "Authorization: Bearer $TOKEN"

# Should return device list, NOT 401 error
```

### Phase 4: Browser Console Checks

Open DevTools (F12) → Console tab and run:

```javascript
// 1. Check localStorage
console.log('=== LocalStorage ===')
console.log('Token:', localStorage.getItem('auth_token')?.substring(0, 50) + '...')
console.log('Refresh:', localStorage.getItem('auth_refresh_token')?.substring(0, 50) + '...')
console.log('Username:', localStorage.getItem('auth_username'))

// 2. Test API directly
console.log('\n=== Testing API ===')
fetch('/api/health-check')
  .then(r => r.json())
  .then(d => console.log('Health check:', d))
  .catch(e => console.error('Health check failed:', e))

// 3. Test with token
const token = localStorage.getItem('auth_token')
fetch('/api/devices', {
  headers: { 'Authorization': `Bearer ${token}` }
})
  .then(r => r.json())
  .then(d => console.log('Devices:', d))
  .catch(e => console.error('Devices failed:', e))
```

## Common Issues & Fixes

### Issue 1: "Token not stored"
**Symptom**: localStorage shows undefined for token

**Fix**:
```javascript
// Clear and re-login
localStorage.clear()
location.href = '/login'
```

### Issue 2: "401 Unauthorized on every request"
**Symptom**: Console shows `[API] 401 Unauthorized on: /api/devices`

**Fix**: 
- Check if backend is running: `curl http://localhost:1234/health-check`
- Check if token is valid: `curl -H "Authorization: Bearer TOKEN" http://localhost:1234/devices`
- Restart backend: `cd backend && go run main.go`

### Issue 3: "Refresh token failed"
**Symptom**: Console shows `[API] Token refresh failed`

**Fix**:
- Test refresh endpoint: 
```bash
curl -X POST http://localhost:1234/refresh-token \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"REFRESH_TOKEN"}'
```
- If it fails, restart backend

### Issue 4: "CORS error"
**Symptom**: Console shows CORS error

**Fix**:
- Make sure frontend is running on port 3555 (or configured port)
- Make sure backend is on port 1234
- Check Vite proxy config in `frontend/vite.config.ts`

## Quick Fix Checklist

- [ ] Backend running: `cd backend && go run main.go`
- [ ] Frontend running: `cd frontend && npm run dev`
- [ ] Browser console shows no errors
- [ ] localStorage has token after login
- [ ] `/api/health-check` returns 200
- [ ] `/api/devices` with token returns 200
- [ ] No 401 errors in Network tab

## If Still Not Working

Please provide:
1. **Backend logs** (last 20 lines)
2. **Browser console logs** (when clicking button)
3. **Network tab** (show /api/devices request)
4. **localStorage contents** (from console)

Then I can identify the exact issue!
