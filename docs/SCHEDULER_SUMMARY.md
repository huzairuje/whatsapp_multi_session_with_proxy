# Scheduler System - Implementation Summary

## Overview

Successfully implemented a smart scheduler that:
1. **Automatically schedules bulk sends >10 recipients** across multiple days
2. **Respects awaken hours** (8 AM - 10 PM, configurable)
3. **Rotates message variants** to avoid spam detection
4. **Integrates with warm-up limits** for safe sending

---

## Files Created (3 files)

### Backend
```
backend/
├── scheduler/
│   ├── model.go              - Scheduled job data models
│   └── service.go            - Database operations and scheduling logic
└── worker/
    └── scheduler_worker.go   - Background worker to process scheduled jobs
```

---

## Files Modified (5 files)

### Backend
1. **commandhandler/commandhandler.go** - Added SchedulerService
2. **handler/handler.go** - Added scheduler import, updated bulk send logic
3. **handler/scheduler_handler.go** - Created scheduler endpoints
4. **routers/routers.go** - Added 3 scheduler routes
5. **boot/setup.go** - Initialize scheduler service and start worker

---

## How It Works

### Automatic Scheduling (>10 Recipients)
```
User sends bulk to 100 recipients
↓
System detects >10 recipients
↓
Creates scheduled job in database
↓
Returns job ID and schedule
↓
Background worker processes job gradually
↓
Respects warm-up limits + time restrictions
↓
Rotates message variants
```

### Manual Sending (≤10 Recipients)
```
User sends bulk to 5 recipients
↓
System detects ≤10 recipients
↓
Sends immediately with anti-ban delays
```

---

## API Endpoints

### Create Scheduled Job
```bash
POST /scheduler/jobs
{
  "sender_jid": "6281234567890",
  "recipients": ["628...", "628...", ...],
  "message_variants": ["Hi {{name}}!", "Hello {{name}}!", "Hey {{name}}!"],
  "template_id": 1
}
```

### Get Scheduled Job
```bash
GET /scheduler/jobs?id=1
```

### Get Pending Jobs
```bash
GET /scheduler/jobs/pending
```

---

## Database Schema

```sql
CREATE TABLE scheduled_jobs (
  id INTEGER PRIMARY KEY,
  sender_jid TEXT NOT NULL,
  template_id INTEGER,
  recipients TEXT NOT NULL,
  message_variants TEXT,
  total_messages INTEGER NOT NULL,
  sent_messages INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  scheduled_for TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
```

---

## Key Features

### 1. Automatic Scheduling
- Bulk sends >10 recipients automatically scheduled
- Spreads across multiple days based on warm-up limits
- Respects time restrictions (8 AM - 10 PM)

### 2. Message Variant Rotation
- Provide multiple message variations
- System rotates through variants
- Each recipient gets different message
- Avoids spam pattern detection

### 3. Background Worker
- Runs every minute
- Checks for pending jobs
- Processes jobs respecting limits
- Updates job status automatically

### 4. Integration with Warm-up
- Uses warm-up daily limits
- Falls back to config limits if no warm-up
- Automatically pauses when limit reached
- Resumes next day

---

## Example Usage

### Scenario: Send to 100 Recipients

**Request:**
```bash
POST /send-bulk?sender=6281234567890
{
  "recipients": ["628...", "628...", ...], // 100 recipients
  "message_variants": [
    "Hi {{name}}, check out our offer!",
    "Hello {{name}}, special deal for you!",
    "Hey {{name}}, don't miss this!"
  ]
}
```

**Response:**
```json
{
  "message": "bulk send scheduled",
  "job_id": 1,
  "recipients": 100,
  "scheduled_for": "2026-05-19T08:00:00Z",
  "note": "messages will be sent gradually across multiple days"
}
```

**What Happens:**
- Day 1: Send 5 messages (warm-up limit)
- Day 2: Send 5 messages
- Day 3: Send 5 messages
- Day 4: Send 10 messages (limit increased)
- Continue until all 100 sent

**Message Rotation:**
- Recipient 1: "Hi John, check out our offer!"
- Recipient 2: "Hello Jane, special deal for you!"
- Recipient 3: "Hey Bob, don't miss this!"
- Recipient 4: "Hi Alice, check out our offer!"
- And so on...

---

## Benefits

### Account Safety
- Gradual sending prevents flagging
- Time restrictions avoid suspicious patterns
- Message rotation avoids spam detection

### Automation
- Set it and forget it
- Background worker handles everything
- Automatic retry on next day

### Flexibility
- Works with warm-up limits
- Works with templates
- Works with time restrictions

---

## Build Status

```
Backend: ✅ Builds successfully
Worker:  ✅ Started automatically on boot
```

---

## Complete Feature Set

### Warm-up (5 → 1000 messages/day)
✅ Gradual limit increase
✅ Per-sender configuration
✅ Real-time monitoring

### Templates (Personalization)
✅ Variable substitution
✅ Per-recipient customization
✅ Preview functionality

### Scheduler (Smart Distribution)
✅ Automatic scheduling >10 recipients
✅ Time restriction enforcement
✅ Message variant rotation
✅ Background processing

---

**Status:** ✅ Complete and Ready for Testing  
**Implementation Date:** May 19, 2026
