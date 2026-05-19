# Practical Examples: Warm-up + Template System

## Real-World Usage Scenarios

This guide provides practical examples of using the warm-up and template systems together for safe, personalized bulk messaging.

---

## Scenario 1: New Business Launch

**Situation:** You're launching a new business and want to notify 500 potential customers.

**Problem:** Sending 500 identical messages on day 1 will get your account banned.

**Solution:** Warm-up + Templates

### Step 1: Set Up Warm-up (Day 1)
```bash
curl -X POST http://localhost:1234/warmup \
  -H "Authorization: Bearer <token>" \
  -d '{
    "sender_jid": "6281234567890",
    "enabled": true,
    "daily_limit": 5,
    "increment_amount": 5,
    "increment_days": 3,
    "max_daily_limit": 1000
  }'
```

### Step 2: Create Template (Day 1)
```bash
curl -X POST http://localhost:1234/templates \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Business Launch",
    "description": "Announce new business to potential customers",
    "content": "Hi {{name}}, we are excited to announce the launch of {{business}}! As a valued {{customer_type}}, you get {{discount}}% off your first order. Visit: {{website}}"
  }'
```

### Step 3: Prepare Recipient Data
```json
{
  "recipients": [
    {
      "phone": "6289876543210",
      "variables": {
        "name": "John",
        "business": "Acme Store",
        "customer_type": "early supporter",
        "discount": "20",
        "website": "acme.com"
      }
    },
    {
      "phone": "6289876543211",
      "variables": {
        "name": "Jane",
        "business": "Acme Store",
        "customer_type": "VIP member",
        "discount": "25",
        "website": "acme.com"
      }
    }
  ]
}
```

### Step 4: Send Daily (Days 1-100)
```bash
# Day 1-3: Send 5 messages/day
# Day 4-6: Send 10 messages/day
# Day 7-9: Send 15 messages/day
# Continue until all 500 customers notified
```

**Timeline:**
- Days 1-3: 15 messages sent (5/day × 3 days)
- Days 4-6: 30 messages sent (10/day × 3 days)
- Days 7-9: 45 messages sent (15/day × 3 days)
- By day 100: All 500 customers reached safely

**Result:** ✅ Account stays safe, each customer gets personalized message

---

## Scenario 2: E-commerce Order Confirmations

**Situation:** You run an online store and need to send order confirmations.

**Problem:** Sending identical "Order confirmed" messages looks like spam.

**Solution:** Template with order-specific variables

### Template
```json
{
  "name": "Order Confirmation",
  "content": "Hi {{customer_name}}, your order #{{order_id}} has been confirmed! Total: ${{amount}}. Estimated delivery: {{delivery_date}}. Track your order: {{tracking_url}}"
}
```

### Usage
```bash
curl -X POST "http://localhost:1234/send-bulk?sender=6281234567890" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "template_id": 2,
    "recipients": [
      {
        "phone": "6289876543210",
        "variables": {
          "customer_name": "John Doe",
          "order_id": "ORD-12345",
          "amount": "99.99",
          "delivery_date": "May 25, 2026",
          "tracking_url": "track.acme.com/ORD-12345"
        }
      }
    ]
  }'
```

**Result:** Each customer gets unique order details, no spam flagging

---

## Scenario 3: Appointment Reminders

**Situation:** Medical clinic needs to send appointment reminders to 200 patients.

**Problem:** Identical reminders sent to many people = spam.

**Solution:** Personalized template with patient-specific details

### Template
```json
{
  "name": "Appointment Reminder",
  "content": "Dear {{patient_name}}, this is a reminder for your appointment with Dr. {{doctor_name}} on {{date}} at {{time}}. Location: {{clinic_address}}. Reply CONFIRM to confirm or RESCHEDULE to change."
}
```

### Warm-up Configuration
```json
{
  "daily_limit": 20,
  "increment_amount": 10,
  "increment_days": 5,
  "max_daily_limit": 200
}
```

### Daily Sending
```bash
# Day 1-5: Send 20 reminders/day
# Day 6-10: Send 30 reminders/day
# Day 11-15: Send 40 reminders/day
# Continue until all patients notified
```

**Result:** All patients get personalized reminders, account stays safe

---

## Scenario 4: Marketing Campaign

**Situation:** Run a promotional campaign for 1000 customers.

**Problem:** Need to send quickly but safely.

**Solution:** Aggressive warm-up + multiple message variations

### Multiple Templates (Avoid Pattern Detection)

**Template 1: Discount Focus**
```
Hi {{name}}! Get {{discount}}% off {{product}} today only! Use code: {{code}}
```

**Template 2: Urgency Focus**
```
{{name}}, hurry! Only {{hours}} hours left for {{discount}}% off {{product}}! Code: {{code}}
```

**Template 3: Benefit Focus**
```
Hey {{name}}! Save big on {{product}} - {{discount}}% off with code {{code}}. Don't miss out!
```

### Warm-up Strategy
```json
{
  "daily_limit": 50,
  "increment_amount": 50,
  "increment_days": 7,
  "max_daily_limit": 1000
}
```

### Rotation Strategy
```javascript
// Rotate templates to avoid pattern detection
const templates = [1, 2, 3]
const recipients = [...] // 1000 recipients

recipients.forEach((recipient, index) => {
  const templateId = templates[index % 3]
  sendMessage(templateId, recipient)
})
```

**Result:** 1000 customers reached in ~20 days with varied messages

---

## Scenario 5: Event Invitations

**Situation:** Invite 300 people to a corporate event.

**Problem:** Mass invitations look like spam.

**Solution:** Personalized invitations with RSVP tracking

### Template
```json
{
  "name": "Event Invitation",
  "content": "Dear {{name}}, you're invited to {{event_name}} on {{date}} at {{venue}}. As our {{relationship}}, we'd love to have you join us. RSVP: {{rsvp_link}}"
}
```

### Segmented Sending
```javascript
// VIP guests (send first)
const vipGuests = [
  { phone: "628...", variables: { relationship: "valued partner" } }
]

// Regular guests (send after)
const regularGuests = [
  { phone: "628...", variables: { relationship: "esteemed guest" } }
]

// Send VIPs first, then regular guests
```

**Result:** Personalized invitations, proper etiquette, no spam flags

---

## Scenario 6: Customer Support Follow-ups

**Situation:** Follow up with 100 customers who contacted support.

**Problem:** Generic follow-ups feel impersonal.

**Solution:** Template with ticket-specific details

### Template
```json
{
  "name": "Support Follow-up",
  "content": "Hi {{customer_name}}, this is a follow-up on ticket #{{ticket_id}}. Your issue regarding {{issue_summary}} has been {{status}}. {{resolution_details}}. Need more help? Reply to this message."
}
```

### Usage
```javascript
const tickets = [
  {
    phone: "6289876543210",
    variables: {
      customer_name: "John",
      ticket_id: "TKT-001",
      issue_summary: "login problem",
      status: "resolved",
      resolution_details: "Password reset link sent to your email"
    }
  }
]
```

**Result:** Personalized follow-ups, improved customer satisfaction

---

## Scenario 7: Educational Course Updates

**Situation:** Notify 500 students about course updates.

**Problem:** Bulk notifications to students = spam risk.

**Solution:** Personalized updates with student-specific info

### Template
```json
{
  "name": "Course Update",
  "content": "Hi {{student_name}}, update for {{course_name}}: {{update_message}}. Your current progress: {{progress}}%. Next assignment due: {{due_date}}. Questions? Reply here."
}
```

### Warm-up for Educational Account
```json
{
  "daily_limit": 30,
  "increment_amount": 20,
  "increment_days": 5,
  "max_daily_limit": 500
}
```

**Result:** All students notified with personalized progress info

---

## Scenario 8: Real Estate Property Alerts

**Situation:** Alert 200 buyers about new properties matching their criteria.

**Problem:** Generic property alerts ignored or flagged as spam.

**Solution:** Personalized alerts based on buyer preferences

### Template
```json
{
  "name": "Property Alert",
  "content": "Hi {{buyer_name}}, new property in {{location}} matches your criteria! {{bedrooms}}BR, {{bathrooms}}BA, ${{price}}. {{special_features}}. View: {{property_url}}"
}
```

### Personalization
```javascript
const buyers = [
  {
    phone: "6289876543210",
    variables: {
      buyer_name: "John",
      location: "Downtown Jakarta",
      bedrooms: "3",
      bathrooms: "2",
      price: "500,000",
      special_features: "Pool, gym, parking",
      property_url: "realestate.com/prop123"
    }
  }
]
```

**Result:** Relevant alerts, higher engagement, no spam flags

---

## Best Practices Summary

### 1. Always Use Templates
❌ **Bad:** Send identical message to 100 people  
✅ **Good:** Use template with variables for personalization

### 2. Respect Warm-up Limits
❌ **Bad:** Try to send 1000 messages on day 1  
✅ **Good:** Start with 5/day, increase gradually

### 3. Vary Your Messages
❌ **Bad:** Use same template for everyone  
✅ **Good:** Create 2-3 template variations, rotate them

### 4. Personalize Meaningfully
❌ **Bad:** Only use {{name}}  
✅ **Good:** Use multiple relevant variables (name, order_id, date, etc.)

### 5. Monitor and Adjust
❌ **Bad:** Set it and forget it  
✅ **Good:** Check warm-up status daily, adjust if needed

### 6. Time Your Messages
❌ **Bad:** Send at 2 AM  
✅ **Good:** Send during business hours (8 AM - 10 PM)

### 7. Test Before Bulk Send
❌ **Bad:** Send to 1000 people immediately  
✅ **Good:** Test with 5-10 people first, verify delivery

---

## Troubleshooting Common Issues

### Issue 1: Messages Not Sending
**Cause:** Daily limit reached  
**Solution:** Check warm-up status, wait for next day or increase limit

### Issue 2: Variables Not Replaced
**Cause:** Variable names don't match  
**Solution:** Ensure template uses `{{name}}` and data provides `name`

### Issue 3: Account Flagged
**Cause:** Sent too many identical messages  
**Solution:** Use templates with more variables, slow down sending

### Issue 4: Low Delivery Rate
**Cause:** Recipients blocking or reporting spam  
**Solution:** Improve message relevance, reduce frequency

---

## Success Metrics

### Track These Metrics:
1. **Delivery Rate:** % of messages delivered successfully
2. **Read Rate:** % of messages read by recipients
3. **Response Rate:** % of recipients who reply
4. **Block Rate:** % of recipients who block you
5. **Warm-up Progress:** Current daily limit vs. target

### Target Benchmarks:
- Delivery Rate: >95%
- Read Rate: >60%
- Response Rate: >10%
- Block Rate: <1%
- Warm-up: Steady increase without flags

---

## Conclusion

The warm-up + template system enables safe, personalized bulk messaging when used correctly:

1. **Start slow** (5 messages/day for new accounts)
2. **Personalize everything** (use templates with variables)
3. **Monitor progress** (check status daily)
4. **Adjust as needed** (increase limits gradually)
5. **Stay patient** (rushing = account ban)

**Remember:** The goal is long-term account health, not short-term message volume.

---

**Last Updated:** May 19, 2026
