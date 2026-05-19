# Session Logout Diagnostic Guide

## Problem
Clicking "Manage Sessions" or "Send Bulk Messages" buttons causes logout/redirect to login.

## Root Cause Analysis

The issue is likely one of these:

### 1. Token Not Being Sent with Requests
The API interceptor should add the token to every request:
```typescript
// In api.ts - request interceptor
const token = localStorage.getItem('auth_token')
if (token) {
  config.headers.Authorization = `Bearer ${token}`
}
```

### 2. Token Refresh Failing
When a 401 occurs, the interceptor tries to refresh:
```typescript
const response = await axios.post('/api/refresh-token', {
  refresh_token: refreshToken,
})
```

### 3. Timing Issue
The auth context might not have loaded the token from localStorage before API calls are made.

## Quick Diagnostic Steps

### Step 1: Check Browser Console
1. Open DevTools (F12)
2. Go to Console tab
3. Look for any error messages
4. Check Network tab for failed requests

### Step 2: Check LocalStorage
In browser console, run:
```javascript
console.log('Token:', localStorage.getItem('auth_token'))
console.log('Refresh:', localStorage.getItem('auth_refresh_token'))
console.log('Username:', localStorage.getItem('auth_username'))
```

All three should have values after login.

### Step 3: Check Network Requests
1. Open DevTools Network tab
2. Click "Manage Sessions"
3. Look for requests to `/api/devices`
4. Check if Authorization header is present
5. Check response status (should be 200, not 401)

### Step 4: Check Refresh Endpoint
In browser console:
```javascript
fetch('/api/refresh-token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ 
    refresh_token: localStorage.getItem('auth_refresh_token')
  })
}).then(r => r.json()).then(console.log)
```

Should return new tokens, not an error.

## Solution: Add Debug Logging

I'll add console logging to help identify the issue. This will show:
1. When tokens are stored/retrieved
2. When API calls are made
3. When refresh happens
4. Any errors

Would you like me to:
1. Add debug logging to the frontend?
2. Check the backend logs for 401 errors?
3. Test the full flow step by step?

Let me know what you see in the browser console and Network tab, and I can pinpoint the exact issue.
