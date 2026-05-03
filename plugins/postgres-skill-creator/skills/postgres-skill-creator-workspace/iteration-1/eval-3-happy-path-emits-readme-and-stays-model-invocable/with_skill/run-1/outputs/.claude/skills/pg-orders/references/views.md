# Views

## public.order_summary

Pre-joined view of orders with the placing user and a roll-up of line items. Use this for the common "list orders with customer name and item count" pattern.

```sql
SELECT
    o.id              AS order_id,
    o.status,
    o.total_cents,
    o.created_at,
    u.id              AS user_id,
    u.email           AS user_email,
    u.name            AS user_name,
    COUNT(oi.product_id)        AS line_item_count,
    SUM(oi.quantity)            AS units_total
FROM public.orders        AS o
JOIN public.users         AS u  ON u.id = o.user_id
LEFT JOIN public.order_items AS oi ON oi.order_id = o.id
GROUP BY o.id, u.id;
```
