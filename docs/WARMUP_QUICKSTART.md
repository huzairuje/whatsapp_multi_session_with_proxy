# Quick Start: WhatsApp Warm-up Feature

## What is Warm-up?

Warm-up gradually increases your daily sending limit from 5 messages/day to 1000 messages/day over time. This prevents Meta from flagging your account as spam.

**Without warm-up:** Send 1000 messages on day 1 → Account gets banned ❌  
**With warm-up:** Start with 5/day, increase gradually → Account stays safe ✅

## Setup (5 minutes)

### 1. Start the Application

```bash
# Terminal 1 - Backend
cd backend
go run main.go

# Terminal 2 - Frontend
cd frontend
npm run dev
```

### 2. Login to Dashboard

1. Open http://localhost:5173
2. Login with default credentials:
   - Username: `admin`
   - Password: `admin123`

### 3. Create Warm-up Configuration

**Option A: Via UI (Recommended)**

1. Click **Warm-up** in the sidebar
2. Click **New Warm-up** button
3. Fill in the form:
   - **Sender JID**: Your WhatsApp number (e.g., `6281234567890`)
   - **Starting Daily Limit**: `5` (for new accounts)
   - **Increment Amount**: `5` (add 5 messages each period)
   - **Increment Period**: `3` (days between increases)
   - **Maximum Daily Limit**: `1000` (safety cap)
   - **Enable**: ✓ Checked
4. Click **Create Configuration**

**Option B: Via API**

```bash
curl -X POST http://localhost:1234/warmup \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sender_jid": "6281234567890",
    "enabled": true,
    "daily_limit": 5,
    "increment_amount": 5,
    "increment_days": 3,
    "max_daily_limit": 1000
  }'
```

## Usage

### Send Bulk Messages

The warm-up limit is automatically enforced:

```bash
curl -X POST "http://localhost:1234/send-bulk?sender=6281234567890" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipients": [
      "6289876543210",
      "6289876543211",
      "6289876543212"
    ],
    "message": "Hello! This is a test message."
  }'
```

**Result:**
- Day 1-3: Only 5 messages sent (warm-up limit)
- Day 4-6: Only 10 messages sent
- Day 7-9: Only 15 messages sent
- And so on...

### Monitor Progress

**Via UI:**
1. Go to **Warm-up** page
2. Click **View Status** on your configuration
3. See:
   - Current Day: How many days since start
   - Current Limit: Today's allowed messages
   - Projected (30d): Estimated limit in 30 days

**Via API:**
```bash
curl "http://localhost:1234/warmup/status?sender=6281234567890" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Recommended Schedules

### New Account (Never Used)
```json
{
  "daily_limit": 5,
  "increment_amount": 5,
  "increment_days": 3,
  "max_daily_limit": 1000
}
```
- Reaches 1000/day in ~600 days (20 months)
- Safest option for brand new accounts

### Young Account (1-3 months old)
```json
{
  "daily_limit": 10,
  "increment_amount": 10,
  "increment_days": 5,
  "max_daily_limit": 1000
}
```
- Reaches 1000/day in ~500 days (16 months)
- Good for accounts with some history

### Established Account (6+ months old)
```json
{
  "daily_limit": 50,
  "increment_amount": 50,
  "increment_days": 7,
  "max_daily_limit": 1000
}
```
- Reaches 1000/day in ~140 days (4.5 months)
- For accounts with proven sending history

## Important Rules

### ✅ DO:
- Start with 1-5 messages/day for new accounts
- Be patient - gradual growth is key
- Monitor for any warnings from WhatsApp
- Keep max limit at 1000 or below
- Send only during business hours (8 AM - 10 PM)

### ❌ DON'T:
- Start with high daily limits (>10 for new accounts)
- Exceed 1000 messages/day
- Disable warm-up once started
- Send messages outside allowed hours
- Rush the warm-up process

## Troubleshooting

### "Daily limit reached" message
✅ **This is normal!** Warm-up is working correctly. Wait until tomorrow or next increment period.

### Messages not sending
1. Check if warm-up is enabled: `GET /warmup/status?sender=YOUR_NUMBER`
2. Verify current limit hasn't been reached
3. Check if within allowed hours (8 AM - 10 PM)
4. Review backend logs for errors

### Want to increase limit faster
1. Go to **Warm-up** page
2. Click **Edit** on your configuration
3. Increase `increment_amount` or decrease `increment_days`
4. Click **Update Configuration**

⚠️ **Warning:** Increasing too fast may trigger Meta's spam detection!

## Example Timeline

**Configuration:** Start at 5, +5 every 3 days, max 1000

| Day Range | Daily Limit | Total Sent (3 days) |
|-----------|-------------|---------------------|
| 1-3       | 5           | 15                  |
| 4-6       | 10          | 30                  |
| 7-9       | 15          | 45                  |
| 10-12     | 20          | 60                  |
| 13-15     | 25          | 75                  |
| ...       | ...         | ...                 |
| 595-597   | 995         | 2,985               |
| 598-600   | 1000        | 3,000               |

**Total time to reach 1000/day:** ~600 days (20 months)

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/warmup` | POST | Create configuration |
| `/warmup?sender=X` | GET | Get configuration |
| `/warmup/all` | GET | List all configurations |
| `/warmup?sender=X` | PUT | Update configuration |
| `/warmup?sender=X` | DELETE | Delete configuration |
| `/warmup/status?sender=X` | GET | Get current status |

## Support

- **Documentation:** See `WARMUP_GUIDE.md` for detailed info
- **Implementation:** See `WARMUP_IMPLEMENTATION.md` for technical details
- **Issues:** Check backend logs for error messages

## Next Steps

1. ✅ Create warm-up configuration for your sender
2. ✅ Send test bulk messages to verify limit enforcement
3. ✅ Monitor progress daily via UI or API
4. ✅ Adjust configuration if needed
5. ✅ Wait patiently for limits to increase

**Remember:** Patience is key! Rushing the warm-up process will get your account banned. Trust the system and let it work gradually.

---

**Status:** ✅ Ready to use  
**Last Updated:** 2026-05-19
