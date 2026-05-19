# Session Logout - Diagnostic Steps

## What You Need to Do

### Step 1: Open Browser DevTools
1. Press **F12** on your keyboard
2. Go to **Console** tab
3. Keep it open while testing

### Step 2: Clear Everything and Login Fresh
```javascript
// In browser console, run:
localStorage.clear()
sessionStorage.clear()
location.reload()
```

### Step 3: Login Again
- Username: `admin`
- Password: `GNn0geM51w0NxmW6`

### Step 4: Watch Console for These Logs
After login, you should see:
```
[Login] Attempting login with username: admin
[Login] Login successful
[Login] Stored tokens: { token: "eyJ...", refreshToken: "eyJ...", username: "admin" }
```

### Step 5: Click "Manage Sessions" Button
Watch the console for:
```
[API] Request to: /api/devices Token present: true
```

### Step 6: Check What Happens
- **If you see 401 error**: Token is not being sent or is invalid
- **If you see refresh attempt**: Token refresh is being triggered
- **If you get logged out**: Refresh is failing

## What to Report Back

Please tell me:
1. **What console logs do you see?** (Copy/paste them)
2. **Do you see `[API] Request to: /api/devices`?**
3. **Do you see `[API] 401 Unauthorized`?**
4. **Do you see `[API] Token refresh failed`?**
5. **What's the exact error message?**

## Quick Checks in Console

Run these commands in browser console:

```javascript
// Check if tokens are stored
console.log('Token:', localStorage.getItem('auth_token') ? 'YES' : 'NO')
console.log('Refresh:', localStorage.getItem('auth_refresh_token') ? 'YES' : 'NO')
console.log('Username:', localStorage.getItem('auth_username'))

// Check if API is accessible
fetch('/api/health-check').then(r => r.json()).then(console.log)
```

## Most Likely Issues

1. **Tokens not stored** → Login context issue
2. **Token not sent** → Request interceptor issue
3. **401 on every request** → Token validation issue
4. **Refresh fails** → Refresh endpoint issue

Once you provide the console logs, I can pinpoint the exact problem!
