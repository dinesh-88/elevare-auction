# Settlement MVP Runbook

1. Create auction with 2 lots and place bids.
2. End both lots so invoices are issued.
3. Open `/admin/auctions/{auctionId}/settlement`.
4. Verify summary cards match invoice list totals.
5. Mark one invoice paid with a payment reference.
6. Verify:
   - invoice status changes to `PAID`
   - `paid_at` is set
   - total paid increases
   - outstanding decreases
7. Cancel one unpaid invoice with a reason.
8. Verify:
   - invoice status changes to `CANCELLED`
   - cancelled invoice is excluded from hammer/fees/tax/total due aggregates
   - counts include cancelled invoices
9. Click `Export CSV`.
10. Verify CSV rows match settlement table and amounts.

## SQL Checks

```sql
SELECT id, invoice_number, payment_status, paid_at, payment_reference, cancelled_at, cancel_reason
FROM invoices
WHERE tenant_id = '<tenant_id>' AND auction_id = '<auction_id>'
ORDER BY issued_at DESC;

SELECT action, entity_type, entity_id, metadata, created_at
FROM audit_logs
WHERE tenant_id = '<tenant_id>'
  AND action IN ('INVOICE_MARKED_PAID', 'INVOICE_CANCELLED', 'SETTLEMENT_CSV_EXPORTED')
ORDER BY created_at DESC;
```
