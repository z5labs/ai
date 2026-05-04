# Views

## public.order_summary

Pre-joined roll-up of per-order totals with customer information and line-item counts, useful for reporting and dashboarding queries.

```sql
SELECT
    o.id         AS order_id,
    o.status,
    o.total_amount,
    o.created_at AS ordered_at,
    u.id         AS user_id,
    u.email      AS user_email,
    u.name       AS user_name,
    COUNT(oi.id) AS line_item_count
FROM orders o
JOIN users u ON u.id = o.user_id
LEFT JOIN order_items oi ON oi.order_id = o.id
GROUP BY o.id, o.status, o.total_amount, o.created_at, u.id, u.email, u.name;
```
