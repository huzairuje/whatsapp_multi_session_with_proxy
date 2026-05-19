# Activity Logging Implementation

## Overview
Comprehensive activity logging system to track all important events in the WhatsApp multi-session system.

## Activity Types Tracked
- `session_connect` - Session connected
- `session_disconnect` - Session disconnected
- `session_logout` - Session logged out
- `message_sent` - Message sent successfully
- `message_failed` - Message send failed
- `bulk_send_start` - Bulk send started
- `bulk_send_complete` - Bulk send completed
- `bulk_send_error` - Bulk send error
- `rate_limit` - Rate limit detected
- `auto_login` - Auto login event
- `health_check` - Health check passed
- `health_check_failed` - Health check failed
- `qr_generated` - QR code generated
- `user_login` - User logged in
- `user_logout` - User logged out

## Backend Components

### Files Created
- `backend/activity/model.go` - Activity models and types
- `backend/activity/repository.go` - Database operations
- `backend/activity/service.go` - Business logic
- `backend/handler/activity_handler.go` - API handlers

### Files Updated
- `backend/handler/handler.go` - Added ActivityService
- `backend/boot/setup.go` - Initialize activity service
- `backend/routers/routers.go` - Added activity endpoints

## API Endpoints

### Protected Routes (require auth)
- `POST /activities/log` - Log new activity
- `GET /activities` - Get recent activities (query: limit)
- `GET /activities/sender` - Get activities by sender (query: sender, limit)
- `GET /activities/type` - Get activities by type (query: type, limit)
- `GET /activities/stats` - Get activity statistics

## Database Schema

### SQLite
```sql
CREATE TABLE activities (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  type TEXT NOT NULL,
  sender TEXT,
  user TEXT,
  message TEXT NOT NULL,
  details TEXT,
  status TEXT,
  error TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

### PostgreSQL
```sql
CREATE TABLE activities (
  id SERIAL PRIMARY KEY,
  type VARCHAR(100) NOT NULL,
  sender VARCHAR(255),
  user VARCHAR(255),
  message TEXT NOT NULL,
  details TEXT,
  status VARCHAR(50),
  error TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

## How to Use in Handlers

### Example: Log Session Connect
```go
// In HandleConnect handler, after successful connection:
_, err = h.ActivityService.LogActivity(
    activity.TypeSessionConnect,
    fmt.Sprintf("Session %s connected", senderString),
    senderString,
    "",
    "",
    "success",
    "",
)
if err != nil {
    log.Errorf("failed to log activity: %v", err)
}
```

### Example: Log Message Sent
```go
// In ServeSendText handler, after successful send:
_, err = h.ActivityService.LogActivity(
    activity.TypeMessageSent,
    fmt.Sprintf("Message sent to %s", requestBody.Recipient),
    senderString,
    "",
    requestBody.Message,
    "success",
    "",
)
```

### Example: Log Bulk Send Start
```go
// In ServeSendTextBulk handler, before starting bulk send:
_, err = h.ActivityService.LogActivity(
    activity.TypeBulkSendStart,
    fmt.Sprintf("Bulk send started for %d recipients", len(requestBody.Recipients)),
    senderString,
    "",
    fmt.Sprintf("Recipients: %d", len(requestBody.Recipients)),
    "started",
    "",
)
```

### Example: Log Rate Limit
```go
// When rate limit is detected:
_, err = h.ActivityService.LogActivity(
    activity.TypeRateLimit,
    "Rate limit detected, backing off",
    senderString,
    "",
    fmt.Sprintf("Backoff duration: %d minutes", config.Conf.BulkSend.ErrorBackoffMinutes),
    "warning",
    "",
)
```

### Example: Log User Login
```go
// In HandleLogin handler, after successful login:
_, err = h.ActivityService.LogActivity(
    activity.TypeUserLogin,
    fmt.Sprintf("User %s logged in", username),
    "",
    username,
    "",
    "success",
    "",
)
```

## Activity Stats Response

```json
{
  "total_activities": 1250,
  "activities_by_type": {
    "session_connect": 45,
    "message_sent": 980,
    "message_failed": 12,
    "bulk_send_start": 15,
    "rate_limit": 3
  },
  "recent_activities": [...],
  "sessions_connected": 45,
  "messages_sent": 980,
  "messages_failed": 12,
  "rate_limit_events": 3
}
```

## Integration Points

### Handlers to Update (Optional)
1. `HandleConnect` - Log session_connect
2. `HandleDisconnect` - Log session_disconnect
3. `Logout` - Log session_logout
4. `ServeSendText` - Already logs message (via message service)
5. `ServeSendTextBulk` - Log bulk_send_start, bulk_send_complete
6. `ServeAutoLogin` - Log auto_login
7. `HandleQR` - Log qr_generated
8. `HandleLogin` - Log user_login
9. `HandleLogout` - Log user_logout

### Frontend Integration
- Display recent activities on dashboard
- Show activity stats
- Filter activities by type or sender
- Real-time activity feed

## Build Status
✅ Backend compiles successfully
✅ All activity endpoints registered
✅ Database tables auto-created on startup
✅ Ready to integrate with handlers

## Next Steps
1. Add activity logging calls to specific handlers
2. Create frontend components to display activities
3. Add real-time activity updates (optional)
4. Add activity filtering and search (optional)
