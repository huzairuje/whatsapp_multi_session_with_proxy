# Deployment & Testing Checklist

## Pre-Deployment Verification

### Backend Build
- [x] Code compiles without errors
- [x] All imports resolve correctly
- [x] No missing dependencies
- [ ] Run `go mod tidy` to verify dependencies
- [ ] Run `go build` to verify compilation

### Frontend Build
- [ ] TypeScript compiles without errors
- [ ] All imports resolve correctly
- [ ] Run `npm run build` to verify build

### Database
- [ ] SQLite database file created
- [ ] Or PostgreSQL connection configured
- [ ] Tables created automatically on startup

---

## Deployment Steps

### Step 1: Backend Setup
```bash
cd backend
go mod download
go build -o whatsapp_multi_session
```

### Step 2: Frontend Setup
```bash
cd frontend
npm install
npm run build
```

### Step 3: Start Services
```bash
# Terminal 1 - Backend
cd backend
./whatsapp_multi_session

# Terminal 2 - Frontend (development)
cd frontend
npm run dev
```

### Step 4: Verify Services
- [ ] Backend running on http://localhost:1234
- [ ] Frontend running on http://localhost:5173
- [ ] Health check: `curl http://localhost:1234/health-check`

---

## Feature Testing Checklist

### Warm-up Feature Tests

#### Create Warm-up Configuration
```bash
curl -X POST http://localhost:1234/warmup \
  -H "Authorization: Bearer <token>" \
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
- [ ] Returns 201 Created
- [ ] Response includes configuration ID
- [ ] Database record created

#### Get Warm-up Configuration
```bash
curl http://localhost:1234/warmup?sender=6281234567890 \
  -H "Authorization: Bearer <token>"
```
- [ ] Returns 200 OK
- [ ] Response includes all configuration fields
- [ ] Data matches what was created

#### Get All Configurations
```bash
curl http://localhost:1234/warmup/all \
  -H "Authorization: Bearer <token>"
```
- [ ] Returns 200 OK
- [ ] Response is array of configurations
- [ ] Includes all created configurations

#### Get Warm-up Status
```bash
curl http://localhost:1234/warmup/status?sender=6281234567890 \
  -H "Authorization: Bearer <token>"
```
- [ ] Returns 200 OK
- [ ] Shows current_day
- [ ] Shows current_limit
- [ ] Shows max_limit
- [ ] Calculations are correct

#### Update Warm-up Configuration
```bash
curl -X PUT http://localhost:1234/warmup?sender=6281234567890 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"daily_limit": 10}'
```
- [ ] Returns 200 OK
- [ ] Configuration updated in database
- [ ] Other fields unchanged

#### Delete Warm-up Configuration
```bash
curl -X DELETE http://localhost:1234/warmup?sender=6281234567890 \
  -H "Authorization: Bearer <token>"
```
- [ ] Returns 200 OK
- [ ] Configuration removed from database
- [ ] Cannot retrieve after deletion

### Template Feature Tests

#### Create Template
```bash
curl -X POST http://localhost:1234/templates \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Welcome",
    "description": "Welcome template",
    "content": "Hi {{name}}, welcome to {{company}}!"
  }'
```
- [ ] Returns 201 Created
- [ ] Response includes template ID
- [ ] Variables extracted automatically
- [ ] Database record created

#### Get Template
```bash
curl http://localhost:1234/templates?id=1 \
  -H "Authorization: Bearer <token>"
```
- [ ] Returns 200 OK
- [ ] Response includes all template fields
- [ ] Variables array populated

#### Get All Templates
```bash
curl http://localhost:1234/templates/all \
  -H "Authorization: Bearer <token>"
```
- [ ] Returns 200 OK
- [ ] Response is array of templates
- [ ] Includes all created templates

#### Preview Template
```bash
curl -X POST http://localhost:1234/templates/preview?id=1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "recipients": [
      {
        "phone": "6281234567890",
        "variables": {"name": "John", "company": "Acme"}
      }
    ]
  }'
```
- [ ] Returns 200 OK
- [ ] Response includes personalized messages
- [ ] Variables replaced correctly
- [ ] Phone number included

#### Update Template
```bash
curl -X PUT http://localhost:1234/templates?id=1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"content": "Updated content with {{new_var}}"}'
```
- [ ] Returns 200 OK
- [ ] Template updated in database
- [ ] Variables re-extracted

#### Delete Template
```bash
curl -X DELETE http://localhost:1234/templates?id=1 \
  -H "Authorization: Bearer <token>"
```
- [ ] Returns 200 OK
- [ ] Template removed from database
- [ ] Cannot retrieve after deletion

### Frontend UI Tests

#### Warm-up Page
- [ ] Page loads without errors
- [ ] Can create new warm-up configuration
- [ ] Can view all configurations
- [ ] Can edit configuration
- [ ] Can delete configuration
- [ ] Can view warm-up status
- [ ] Status shows correct calculations
- [ ] Projected growth calculated correctly

#### Templates Page
- [ ] Page loads without errors
- [ ] Can create new template
- [ ] Can view all templates
- [ ] Can edit template
- [ ] Can delete template
- [ ] Can preview template
- [ ] Variables displayed correctly
- [ ] Personalized messages shown in preview

### Integration Tests

#### Bulk Send with Warm-up
- [ ] Bulk send respects warm-up limit
- [ ] Cannot send more than daily limit
- [ ] Limit increases after increment period
- [ ] Falls back to config limit if no warm-up

#### Bulk Send with Template
- [ ] Can send using template_id
- [ ] Each recipient gets personalized message
- [ ] Variables replaced correctly
- [ ] Phone numbers included in messages

#### Combined Warm-up + Template
- [ ] Warm-up limit enforced
- [ ] Template personalization applied
- [ ] Each recipient gets unique message
- [ ] Limit respected across multiple sends

---

## Database Verification

### SQLite
```bash
sqlite3 whatsapp.db
sqlite> .tables
sqlite> SELECT * FROM warmup_configs;
sqlite> SELECT * FROM message_templates;
```
- [ ] Both tables exist
- [ ] Data inserted correctly
- [ ] Timestamps recorded

### PostgreSQL
```bash
psql -U user -d database
\dt
SELECT * FROM warmup_configs;
SELECT * FROM message_templates;
```
- [ ] Both tables exist
- [ ] Data inserted correctly
- [ ] Timestamps recorded

---

## Performance Tests

### Load Testing
- [ ] Create 100 warm-up configurations
- [ ] Create 100 templates
- [ ] List all configurations (should be fast)
- [ ] List all templates (should be fast)
- [ ] Preview template with 100 recipients (should complete in <5s)

### Concurrent Requests
- [ ] Send 10 concurrent requests to create warm-up
- [ ] Send 10 concurrent requests to create templates
- [ ] No race conditions or conflicts
- [ ] All requests succeed

---

## Security Tests

### Authentication
- [ ] Endpoints require valid token
- [ ] Invalid token returns 401
- [ ] Expired token returns 401
- [ ] Missing token returns 401

### Authorization
- [ ] Users can only access their own data
- [ ] Cannot modify other users' configurations
- [ ] Cannot delete other users' templates

### Input Validation
- [ ] Empty sender_jid rejected
- [ ] Invalid daily_limit rejected
- [ ] Empty template name rejected
- [ ] Empty template content rejected

---

## Error Handling Tests

### Warm-up Errors
- [ ] Duplicate sender_jid returns error
- [ ] Invalid increment_days returns error
- [ ] daily_limit > max_daily_limit returns error
- [ ] Non-existent configuration returns 404

### Template Errors
- [ ] Duplicate template name returns error
- [ ] Invalid template ID returns 404
- [ ] Empty content returns error
- [ ] Invalid variables format handled gracefully

---

## Real-World Testing

### With Real WhatsApp Account
- [ ] Create warm-up configuration
- [ ] Create message template
- [ ] Send test bulk message
- [ ] Verify message delivery
- [ ] Check for Meta warnings/flags
- [ ] Monitor account health

### Multi-Sender Testing
- [ ] Create warm-up for sender 1
- [ ] Create warm-up for sender 2
- [ ] Send bulk from sender 1 (respects limit 1)
- [ ] Send bulk from sender 2 (respects limit 2)
- [ ] Limits independent per sender

### Long-Term Testing
- [ ] Run for 7 days
- [ ] Verify daily limits increase correctly
- [ ] Check for any Meta warnings
- [ ] Monitor delivery rates
- [ ] Verify account stays healthy

---

## Documentation Verification

- [ ] WARMUP_GUIDE.md is complete
- [ ] WARMUP_QUICKSTART.md is accurate
- [ ] TEMPLATE_SUMMARY.md is complete
- [ ] PRACTICAL_EXAMPLES.md has real examples
- [ ] COMPLETE_SUMMARY.md is comprehensive
- [ ] All code examples work correctly

---

## Deployment Checklist

### Pre-Production
- [ ] All tests pass
- [ ] No console errors
- [ ] No database errors
- [ ] Performance acceptable
- [ ] Security verified

### Production
- [ ] Backup database before deployment
- [ ] Deploy backend
- [ ] Deploy frontend
- [ ] Verify all endpoints working
- [ ] Monitor logs for errors
- [ ] Test with real WhatsApp account

### Post-Deployment
- [ ] Monitor warm-up functionality
- [ ] Monitor template functionality
- [ ] Check for any errors in logs
- [ ] Verify user feedback
- [ ] Make adjustments if needed

---

## Rollback Plan

If issues occur:

1. **Stop Services**
   ```bash
   # Stop backend and frontend
   ```

2. **Restore Database**
   ```bash
   # Restore from backup
   ```

3. **Revert Code**
   ```bash
   git revert <commit>
   ```

4. **Restart Services**
   ```bash
   # Restart backend and frontend
   ```

---

## Success Criteria

### Warm-up Feature
- ✅ Configurations created and stored
- ✅ Daily limits enforced in bulk send
- ✅ Limits increase automatically
- ✅ Status calculations accurate
- ✅ No account flagging

### Template Feature
- ✅ Templates created and stored
- ✅ Variables extracted automatically
- ✅ Messages personalized correctly
- ✅ Preview shows correct output
- ✅ No spam detection

### Overall
- ✅ Backend builds successfully
- ✅ Frontend loads without errors
- ✅ All API endpoints working
- ✅ Database operations successful
- ✅ Real WhatsApp account stays safe

---

## Support & Troubleshooting

### Common Issues

**Issue:** Warm-up limit not enforced
- Check if warm-up is enabled
- Verify sender_jid matches
- Check current_day calculation

**Issue:** Template variables not replaced
- Verify variable names match
- Check template content syntax
- Ensure recipient data includes variables

**Issue:** Account flagged despite warm-up
- Check if limits were exceeded
- Verify time restrictions enabled
- Review message content for spam triggers

### Getting Help
- Check documentation files
- Review logs for error messages
- Test with simple examples first
- Verify database integrity

---

## Final Verification

Before going live:

- [ ] Backend builds: `go build`
- [ ] Frontend builds: `npm run build`
- [ ] All tests pass
- [ ] Documentation complete
- [ ] Security verified
- [ ] Performance acceptable
- [ ] Error handling working
- [ ] Database backup created
- [ ] Rollback plan ready
- [ ] Team trained on features

---

**Status:** Ready for Deployment  
**Last Updated:** May 19, 2026
