# Message Template Approval Guide

Complete guide for getting SMS (DLT) and WhatsApp message templates approved.

---

## Part 1: DLT SMS Template Approval

### Template Categories

| Category | Use Case | Approval Time |
|----------|----------|---------------|
| **Service Implicit** | OTP, Alerts, Transactions | 1-2 days |
| **Service Explicit** | Reminders, Notifications | 2-3 days |
| **Promotional** | Marketing, Offers | 3-5 days |

### Template Rules

1. **Variables**: Use `{#var#}` for dynamic content
2. **Length**: Max 160 characters per SMS (or 306 for 2-part)
3. **Language**: Hindi templates need separate registration
4. **Sender ID**: Must match your registered header

### Step-by-Step Template Registration

#### Step 1: Login to DLT Portal
- Jio: https://trueconnect.jio.com
- Use your registered credentials

#### Step 2: Go to Content Template
- Menu → Template → Content Template Registration

#### Step 3: Fill Template Details

**For each template, fill:**

| Field | Description |
|-------|-------------|
| Template Name | Unique identifier (e.g., "cold_storage_otp") |
| Template Type | Transactional / Promotional |
| Content Type | Text / Unicode (for Hindi) |
| Header | Your registered Sender ID |
| Template Message | Your message with {#var#} placeholders |

---

### Cold Storage Templates - Ready to Use

Copy these exact templates for DLT registration:

#### Template 1: OTP Verification
```
Template Name: cold_storage_otp
Type: Transactional (Service Implicit)
Header: COLDST

Message:
Dear Customer, Your OTP for Cold Storage login is {#var#}. Valid for 10 minutes. Do not share. - COLDST
```
**Character Count**: 98

---

#### Template 2: Payment Received
```
Template Name: cold_storage_payment_received
Type: Transactional (Service Implicit)
Header: COLDST

Message:
Dear {#var#}, payment of Rs.{#var#} received at Cold Storage. Balance: Rs.{#var#}. Thank you! - COLDST
```
**Character Count**: 102

---

#### Template 3: Payment Cleared (Zero Balance)
```
Template Name: cold_storage_payment_cleared
Type: Transactional (Service Implicit)
Header: COLDST

Message:
Dear {#var#}, payment of Rs.{#var#} received. Your account is now clear. Thank you for your payment! - COLDST
```
**Character Count**: 106

---

#### Template 4: Item In (Storage)
```
Template Name: cold_storage_item_in
Type: Transactional (Service Implicit)
Header: COLDST

Message:
Dear {#var#}, {#var#} items received at Cold Storage. Thock: {#var#}. Total stored: {#var#}. Thank you! - COLDST
```
**Character Count**: 108

---

#### Template 5: Item Out (Pickup)
```
Template Name: cold_storage_item_out
Type: Transactional (Service Implicit)
Header: COLDST

Message:
Dear {#var#}, {#var#} items picked up. Gate Pass: {#var#}. Remaining: {#var#} items. Thank you! - COLDST
```
**Character Count**: 100

---

#### Template 6: Payment Reminder
```
Template Name: cold_storage_payment_reminder
Type: Transactional (Service Explicit)
Header: COLDST

Message:
Dear {#var#}, your pending balance at Cold Storage is Rs.{#var#}. Please clear dues at earliest. Thank you! - COLDST
```
**Character Count**: 112

---

#### Template 7: Bulk/Promotional Message
```
Template Name: cold_storage_promotional
Type: Promotional
Header: COLDST

Message:
Dear {#var#}, {#var#}. Visit Cold Storage for best rates. Contact: {#var#}. Reply STOP to opt out. - COLDST
```
**Character Count**: 105

---

### Template Approval Checklist

Before submitting, verify:

- [ ] Template name is unique
- [ ] All variables use `{#var#}` format
- [ ] Message ends with `- HEADER` (your sender ID)
- [ ] Character count is under 160 (for single SMS)
- [ ] No special characters that may cause issues
- [ ] Spelling and grammar are correct

### Common Rejection Reasons

| Reason | Fix |
|--------|-----|
| "Variable format incorrect" | Use `{#var#}` not `{var}` or `{{var}}` |
| "Header mismatch" | End message with `- YOURSENDERID` |
| "Promotional content in transactional" | Remove marketing language |
| "Template too similar" | Make content more unique |
| "Invalid characters" | Remove emojis, special symbols |

---

## Part 2: WhatsApp Template Approval

### Template Categories

| Category | Use Case | Cost | Approval |
|----------|----------|------|----------|
| **Utility** | Transactions, Updates | Lower | 1-2 days |
| **Authentication** | OTP, Login | Lower | 1 day |
| **Marketing** | Promotions, Offers | Higher | 2-3 days |

### WhatsApp Template Rules

1. **Variables**: Use `{{1}}`, `{{2}}`, etc.
2. **No URLs** in first message to new users
3. **Must provide value** to customer
4. **Opt-out** required for marketing templates
5. **Language**: Must match selected language code

---

### AiSensy Template Creation

#### Step 1: Login to AiSensy Dashboard
https://app.aisensy.com

#### Step 2: Go to Templates
Menu → Templates → Create New Template

#### Step 3: Fill Template Details

---

### Cold Storage WhatsApp Templates - Ready to Use

#### Template 1: OTP Authentication
```
Template Name: cold_storage_otp
Category: Authentication
Language: English (en)

Body:
Your Cold Storage verification code is {{1}}. Valid for 10 minutes. Do not share this code with anyone.

Footer: Cold Storage Management
```

---

#### Template 2: Payment Received
```
Template Name: payment_received
Category: Utility
Language: English (en)

Body:
Dear {{1}},

We have received your payment of ₹{{2}}.

Remaining balance: ₹{{3}}

Thank you for your payment!

Footer: Cold Storage Management
```

---

#### Template 3: Payment Cleared
```
Template Name: payment_cleared
Category: Utility
Language: English (en)

Body:
Dear {{1}},

We have received your payment of ₹{{2}}.

Your account is now clear with zero balance.

Thank you for your prompt payment!

Footer: Cold Storage Management
```

---

#### Template 4: Item Storage Confirmation
```
Template Name: item_stored
Category: Utility
Language: English (en)

Body:
Dear {{1}},

Your items have been stored successfully.

📦 Items received: {{2}}
🏷️ Thock Number: {{3}}
📊 Total items in storage: {{4}}

Thank you for choosing Cold Storage!

Footer: Cold Storage Management
```

---

#### Template 5: Item Pickup Confirmation
```
Template Name: item_pickup
Category: Utility
Language: English (en)

Body:
Dear {{1}},

Your items have been picked up successfully.

📦 Items picked: {{2}}
🎫 Gate Pass: {{3}}
📊 Remaining in storage: {{4}}

Thank you!

Footer: Cold Storage Management
```

---

#### Template 6: Payment Reminder
```
Template Name: payment_reminder
Category: Utility
Language: English (en)

Body:
Dear {{1}},

This is a reminder about your pending balance.

💰 Amount Due: ₹{{2}}

Please clear your dues at your earliest convenience.

For queries, contact us at the storage facility.

Footer: Cold Storage Management

Buttons:
- Call to Action: "Call Us" → phone_number
```

---

#### Template 7: Bulk Promotional
```
Template Name: promotional_message
Category: Marketing
Language: English (en)

Body:
Dear {{1}},

{{2}}

Visit Cold Storage for the best rates in town!

📞 Contact: {{3}}

Reply STOP to unsubscribe.

Footer: Cold Storage Management

Buttons:
- Quick Reply: "Learn More"
- Quick Reply: "Not Interested"
```

---

### Interakt Template Creation

#### Step 1: Login to Interakt Dashboard
https://app.interakt.shop

#### Step 2: Create Template
Settings → WhatsApp Templates → Create New

#### Step 3: Same templates as above
Use the same content, just change variable format if needed.

---

### WhatsApp Template Approval Tips

#### DO:
- ✅ Use clear, professional language
- ✅ Provide value to the customer
- ✅ Include business name in footer
- ✅ Use proper variable placeholders
- ✅ Keep messages concise

#### DON'T:
- ❌ Use threatening language ("Pay now or else...")
- ❌ Include misleading information
- ❌ Use ALL CAPS
- ❌ Add too many emojis
- ❌ Copy templates from other businesses

### Common WhatsApp Rejection Reasons

| Reason | Fix |
|--------|-----|
| "Template not clear" | Add more context, make purpose obvious |
| "Variable misuse" | Variables should be dynamic, not static text |
| "Missing opt-out" | Add "Reply STOP to unsubscribe" for marketing |
| "URL in template" | Remove URLs or use button instead |
| "Duplicate template" | Make content unique |

---

## Part 3: Template Mapping in Code

After templates are approved, map them in the system:

### DLT Template IDs (Fast2SMS)

After approval, you'll get Template IDs. Add to Fast2SMS:

| Template Name | Template ID | Use In Code |
|---------------|-------------|-------------|
| cold_storage_otp | 110XXXX01 | OTP sending |
| cold_storage_payment_received | 110XXXX02 | Payment notification |
| cold_storage_item_in | 110XXXX03 | Item in notification |
| cold_storage_item_out | 110XXXX04 | Item out notification |
| cold_storage_payment_reminder | 110XXXX05 | Payment reminders |

### WhatsApp Template Names (AiSensy/Interakt)

| Template Name | Use In Code |
|---------------|-------------|
| cold_storage_otp | OTP sending |
| payment_received | Payment notification |
| payment_cleared | Payment cleared notification |
| item_stored | Item in notification |
| item_pickup | Item out notification |
| payment_reminder | Payment reminders |
| promotional_message | Bulk promotions |

---

## Part 4: Testing Templates

### Test SMS Template

1. Go to Fast2SMS Dashboard
2. Send Test SMS
3. Select your DLT template
4. Enter test values for variables
5. Send to your number
6. Verify message received correctly

### Test WhatsApp Template

1. Go to AiSensy/Interakt Dashboard
2. Send Test Message
3. Select approved template
4. Enter variable values
5. Send to your WhatsApp number
6. Verify message and formatting

---

## Approval Timeline Summary

| Item | Time Required |
|------|---------------|
| DLT Portal Registration | 2-7 days |
| Sender ID (Header) Approval | 1-3 days |
| DLT SMS Templates | 1-3 days each |
| WhatsApp Business Verification | 2-7 days |
| WhatsApp Templates | 1-3 days each |

**Total: 2-3 weeks** (if done in parallel)

---

## Support

| Issue | Contact |
|-------|---------|
| DLT Template Rejected | Jio: truconnect.support@jio.com |
| Fast2SMS Issues | support@fast2sms.com |
| WhatsApp Template Rejected | AiSensy: support@aisensy.com |
| Meta Verification Issues | Through AiSensy/Interakt support |
