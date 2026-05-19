# WhatsApp Warm-up Feature Guide

## Overview

The Warm-up feature is designed to gradually increase the daily sending limit for WhatsApp sender accounts to avoid being flagged or banned by Meta's anti-spam system. Instead of sending 1000+ messages on day 1, accounts start with 1-5 messages/day and gradually increase over time.

## Why Warm-up is Important

Meta's WhatsApp system has sophisticated anti-spam detection that flags accounts sending too many messages too quickly. A new account sending 1000 messages on day 1 will almost certainly be flagged. The warm-up feature prevents this by:

- Starting with very low daily limits (1-5 messages)
- Gradually increasing limits over time
- Capping at a safe maximum (1000 messages/day)
- Simulating natural account growth patterns

## How It Works

### Warm-up Schedule Example

**Default Configuration:**
- Starting Daily Limit: 5 messages/day
- Increment: +5 messages every 3 days
- Maximum: 1000 messages/day

**Growth Timeline:**
- Days 1-3: 5 messages/day
- Days 4-6: 10 messages/day
- Days 7-9: 15 messages/day
- Days 10-12: 20 messages/day
- ... continues until reaching 1000 messages/day

### Configuration Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `sender_jid` | - | WhatsApp phone number (e.g., 6281234567890) |
| `enabled` | true | Enable/disable warm-up for this sender |
| `daily_limit` | 5 | Starting daily message limit |
| `increment_amount` | 5 | Messages to add each period |
| `increment_days` | 3 | Days between increments |
| `max_daily_limit` | 1000 | Maximum daily limit (safety cap) |

## API Endpoints

### Create Warm-up Configuration

```bash
POST /warmup
Content-Type: application/json
Authorization: Bearer <token>

{
  "sender_jid": "6281234567890",
  "enabled": true,
  "daily_limit": 5,
  "increment_amount": 5,
  "increment_days": 3,
  "max_daily_limit": 1000
}
```

### Get Warm-up Configuration

```bash
GET /warmup?sender=6281234567890
Authorization: Bearer <token>
```

### Get All Warm-up Configurations

```bash
GET /warmup/all
Authorization: Bearer <token>
```

### Get Warm-up Status

```bash
GET /warmup/status?sender=6281234567890
Authorization: Bearer <token>

Response:
{
  "enabled": true,
  "current_day": 15,
  "current_limit": 20,
  "max_limit": 1000,
  "start_date": "2026-05-04T10:00:00Z",
  "config": { ... }
}
```

### Update Warm-up Configuration

```bash
PUT /warmup?sender=6281234567890
Content-Type: application/json
Authorization: Bearer <token>

{
  "enabled": true,
  "daily_limit": 10,
  "increment_amount": 10,
  "increment_days": 2,
  "max_daily_limit": 1500
}
```

### Delete Warm-up Configuration

```bash
DELETE /warmup?sender=6281234567890
Authorization: Bearer <token>
```

## Frontend Usage

### Access Warm-up Manager

1. Navigate to **Warm-up** in the sidebar
2. View all active warm-up configurations
3. Create new warm-up for a sender
4. Edit existing configurations
5. View current status and projected growth

### Create New Warm-up

1. Click **New Warm-up** button
2. Enter sender phone number (without +)
3. Set starting daily limit (1-5 for new accounts)
4. Set increment amount and period
5. Set maximum daily limit (recommended: 1000)
6. Click **Create Configuration**

### Monitor Warm-up Progress

- **Current Day**: Days since warm-up started
- **Current Limit**: Today's allowed message count
- **Max Limit**: Safety cap (won't exceed this)
- **Projected (30d)**: Estimated limit in 30 days

## Best Practices

### For New Accounts

```json
{
  "daily_limit": 5,
  "increment_amount": 5,
  "increment_days": 3,
  "max_daily_limit": 1000
}
```

This reaches 1000 messages/day in approximately 600 days (20 months).

### For Established Accounts

```json
{
  "daily_limit": 50,
  "increment_amount": 50,
  "increment_days": 7,
  "max_daily_limit": 1000
}
```

This reaches 1000 messages/day in approximately 140 days (4.5 months).

### For High-Volume Accounts

```json
{
  "daily_limit": 100,
  "increment_amount": 100,
  "increment_days": 5,
  "max_daily_limit": 1000
}
```

This reaches 1000 messages/day in approximately 50 days.

## How Bulk Send Uses Warm-up

When you send bulk messages:

1. System checks if warm-up is enabled for the sender
2. Calculates current daily limit based on warm-up schedule
3. Uses warm-up limit instead of fixed config limit
4. Enforces the limit during bulk send operation
5. Stops sending when daily limit is reached

### Example

If a sender has:
- Warm-up enabled with current limit of 20 messages/day
- Tries to send 100 messages
- System will only send 20 messages and stop
- Remaining 80 messages will be marked as "daily limit reached"

## Database Schema

### warmup_configs Table

```sql
CREATE TABLE warmup_configs (
  id INTEGER PRIMARY KEY,
  sender_jid TEXT NOT NULL UNIQUE,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  current_day INTEGER NOT NULL DEFAULT 1,
  start_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  daily_limit INTEGER NOT NULL DEFAULT 5,
  increment_amount INTEGER NOT NULL DEFAULT 5,
  increment_days INTEGER NOT NULL DEFAULT 3,
  max_daily_limit INTEGER NOT NULL DEFAULT 1000,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)
```

## Troubleshooting

### Warm-up Not Applied to Bulk Send

1. Check if warm-up is **enabled** for the sender
2. Verify sender JID format (should be phone number without +)
3. Check current day calculation (should be > 0)
4. Review logs for warm-up service errors

### Current Limit Not Increasing

1. Verify `increment_days` has passed since start date
2. Check if `max_daily_limit` has been reached
3. Ensure warm-up is enabled
4. Check database for correct configuration

### Bulk Send Stops Early

1. This is expected behavior - warm-up limit is being enforced
2. Check current limit with `/warmup/status` endpoint
3. Wait for next increment period or manually increase limit
4. Or disable warm-up if you want to use config limit instead

## Integration with Anti-Ban Features

The warm-up feature works alongside other anti-ban protections:

- **Time Restrictions**: Only sends during allowed hours (8 AM - 10 PM)
- **Batch Pauses**: Takes breaks after N messages
- **Presence Simulation**: Sends typing indicators before messages
- **Random Delays**: Varies delay between messages
- **Rate Limit Detection**: Backs off on rate limit errors

All these features work together to keep accounts safe.

## Configuration Example

```yaml
# config.local.yaml
bulkSend:
  minDelay: 15000
  maxDelay: 45000
  batchSize: 10
  batchPauseMin: 300
  batchPauseMax: 600
  dailyLimit: 50  # Fallback if warm-up not configured
  typingDelayMin: 2000
  typingDelayMax: 5000
  enablePresenceSimulation: true
  allowedHourStart: 8
  allowedHourEnd: 22
  timezone: "Asia/Jakarta"
  enableTimeRestrictions: true
  errorBackoffMinutes: 30
  enableRecipientValidation: true
  validationCacheDuration: 24
  enableHealthCheck: true
  maxErrorRate: 0.3
```

## Support

For issues or questions about the warm-up feature:

1. Check the logs for error messages
2. Verify warm-up configuration in the UI
3. Review this guide for best practices
4. Check GitHub issues for similar problems
