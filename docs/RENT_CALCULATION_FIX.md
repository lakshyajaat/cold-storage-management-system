# Rent Calculation Fix Documentation

**Date:** 2025-12-22
**Version:** 1.5.x

---

## Problem Statement

Rent was being calculated incorrectly across multiple pages. The system was charging customers based on `expected_quantity` (items they brought) instead of `room_entries.quantity` (items actually stored in rooms).

**Example:**
- Customer brings 300 items (expected_quantity = 300)
- Only 250 items are stored in rooms (room_entries.quantity = 250)
- **Old behavior:** Charged rent for 300 items
- **New behavior:** Charged rent for 250 items (correct)

---

## Files Fixed

### Frontend Templates

| File | Issue | Fix |
|------|-------|-----|
| `customer_pdf_export.html` | Used `expected_quantity` | Now fetches `/api/room-entries` and calculates from stored quantity |
| `rent_management.html` | Used `entry.actual_quantity \|\| entry.expected_quantity` | Now uses `thockStoredQty` map from room_entries |
| `rent.html` | Used `thock.actual_quantity` | Now uses `thockStoredQty` map from room_entries |
| `loding_invoice.html` | Used `actual_quantity \|\| expected_quantity` | Now uses `thockStoredQty` map from room_entries |

### Backend

| File | Issue | Fix |
|------|-------|-----|
| `internal/services/customer_portal_service.go` | Hardcoded fallback `160.0` for rent_per_item | Changed to `0.0` (use API only) |

---

## Implementation Pattern

All fixed pages now follow this pattern:

```javascript
// 1. Fetch room_entries alongside other data
const [entriesResponse, roomEntriesResponse] = await Promise.all([
    fetch('/api/entries', { headers: { 'Authorization': `Bearer ${token}` } }),
    fetch('/api/room-entries', { headers: { 'Authorization': `Bearer ${token}` } })
]);

// 2. Build map of thock_number -> stored quantity
const thockStoredQty = {};
roomEntries.forEach(re => {
    if (!thockStoredQty[re.thock_number]) {
        thockStoredQty[re.thock_number] = 0;
    }
    thockStoredQty[re.thock_number] += re.quantity || 0;
});

// 3. Calculate rent from stored quantity
const storedQty = thockStoredQty[entry.thock_number] || 0;
const rent = storedQty * rentPerItem;
```

---

## Display Format

Quantity is now displayed as `expected -> stored` when different:

- **Example:** `300 -> 250` (expected 300, stored 250)
- If same, shows just the number: `250`

---

## Dashboard Updates

### Admin Dashboard (`dashboard_admin.html`)

**System Overview Section:**
| Stat | Description |
|------|-------------|
| Total Users | All registered users |
| Total Customers | All customers |
| Total Entries | All entries (all time) |
| Room Entries | All room entries (all time) |
| **Total Bags** | Sum of all room_entries.quantity (all time) |

**Today's Summary Section:**
| Stat | Description |
|------|-------------|
| Today's Entries | Entries created today |
| Today's Room Entries | Room entries created today |
| Active Users | Currently active users |
| **Today's Bags** | Sum of today's room_entries.quantity |

### Employee Dashboard (`dashboard_employee.html`)

**Today's Summary Section:**
- My Entries Today
- Total Entries
- Room Entries
- **Today's Bags** (fixed - was showing all-time total)

### Admin Report (`admin_report.html`)

**Summary Stats:**
- Total Entries
- Room Entries
- **Total Qty** (sum for selected date range)
- **Today's Qty** (sum for today only)
- Gate Passes
- Total Payments
- Qty Changes
- Total Edits

---

## Rent Per Item Configuration

Rent rate is loaded from the system settings API:

```javascript
const settingsResponse = await fetch('/api/settings/rent_per_item', {
    headers: { 'Authorization': `Bearer ${token}` }
});
const rentSetting = await settingsResponse.json();
const rentPerItem = parseFloat(rentSetting.setting_value);
```

**Important:** No hardcoded fallback values. If API fails, rent rate is `0`.

---

## Testing Checklist

- [ ] Customer Export page shows correct rent based on room_entries
- [ ] Account Management shows correct outstanding amounts
- [ ] Rent Payment page calculates correctly
- [ ] Loading Invoice uses room_entries quantities
- [ ] Admin dashboard shows both Today's Bags and Total Bags
- [ ] Employee dashboard shows Today's Bags correctly
- [ ] Admin report shows Today's Qty and Total Qty separately

---

## Related Thock Number Format

The `thock_number` format is: `[seed/sell number]/[expected quantity]`

**Example:** `0001/300`
- `0001` = seed/sell identifier
- `300` = expected quantity customer brought

Rent is calculated from `room_entries.quantity`, NOT the expected quantity in the thock number.

---

## Verification

To verify rent calculation is correct:

1. Go to `/customer-export`
2. Select a customer
3. Check "Storage Details" section shows room entries
4. Verify rent = (stored quantity) x (rent per item)
5. Compare with Account Management (`/rent-management`) - should match

---

**Last Updated:** 2025-12-22
