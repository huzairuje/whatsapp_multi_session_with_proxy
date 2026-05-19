# Quick Start Guide - Authentication

## 🚀 Starting the Application

### 1. Start Backend
```bash
cd backend
go run main.go
```

**IMPORTANT**: On first startup, look for this in the logs:
```
========================================
ADMIN USER CREATED
Username: admin
Password: aBcDeFgHiJkLmNoP
PLEASE SAVE THIS PASSWORD - IT WILL NOT BE SHOWN AGAIN
========================================
```

**Save this password immediately!** You'll need it to login.

### 2. Start Frontend
```bash
cd frontend
npm run dev
```

### 3. Login
1. Open browser to `http://localhost:5173` (or the port shown by Vite)
2. You'll be redirected to `/login`
3. Enter:
   - Username: `admin`
   - Password: (the one from backend logs)
4. Click "Sign In"
5. You'll be redirected to the dashboard

## 🔐 Authentication Features

### Login
- Simple username/password authentication
- JWT token-based (24-hour expiry)
- Token stored in localStorage
- Auto-redirect to login if not authenticated

### Protected Routes
All routes except `/login` and `/health-check` require authentication:
- Dashboard
- Sessions
- Bulk Send
- Recipients
- Templates
- Settings

### Logout
Click the "Logout" button in the top-right navbar to sign out.

### Change Password
Use the `/api/change-password` endpoint (can be added to Settings page):
```bash
curl -X POST http://localhost:1234/api/change-password \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "current_password",
    "new_password": "new_password"
  }'
```

## 🗄️ Database

### SQLite (Default)
Users table created in `examplestore.db`

### PostgreSQL
If `postgres.enablePostgres: true` in config, users table created in PostgreSQL database.

## 🔧 Configuration

### Change JWT Secret (IMPORTANT for Production!)
Edit `backend/boot/setup.go` line 56:
```go
authService := auth.NewService(rawDB.(*sql.DB), "your-secret-key-change-this", isPostgres)
```

Change `"your-secret-key-change-this"` to a secure random string.

### Token Expiry
Edit `backend/auth/service.go` line 109:
```go
"exp": time.Now().Add(24 * time.Hour).Unix(),
```

Change `24 * time.Hour` to your desired expiry time.

## 🧪 Testing

### Test Login API
```bash
# Login
curl -X POST http://localhost:1234/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"YOUR_PASSWORD"}'

# Response:
# {"token":"eyJhbGc...","username":"admin"}
```

### Test Protected Endpoint
```bash
# Without token (should fail)
curl -X GET http://localhost:1234/api/devices

# With token (should succeed)
curl -X GET http://localhost:1234/api/devices \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Test Health Check (Public)
```bash
curl -X GET http://localhost:1234/api/health-check
```

## 🐛 Troubleshooting

### "Invalid credentials" error
- Check username is exactly `admin` (lowercase)
- Verify password from backend startup logs
- Check backend logs for errors

### "Missing authorization header" error
- Token not being sent with request
- Check browser localStorage for `auth_token`
- Try logging out and back in

### Redirected to login immediately after login
- Token validation failing
- Check JWT secret matches between token generation and validation
- Check token hasn't expired
- Check browser console for errors

### Can't find admin password
- Stop backend
- Delete `examplestore.db` (SQLite) or drop users table (PostgreSQL)
- Restart backend - new password will be generated

## 📝 Notes

- Default admin password is randomly generated on first startup
- Password is hashed with bcrypt (never stored in plain text)
- JWT tokens expire after 24 hours
- All API endpoints (except login and health-check) require authentication
- Frontend automatically redirects to login on 401 responses
