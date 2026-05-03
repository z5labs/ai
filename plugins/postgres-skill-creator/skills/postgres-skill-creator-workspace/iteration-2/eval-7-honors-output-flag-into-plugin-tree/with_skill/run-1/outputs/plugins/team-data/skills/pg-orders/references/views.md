# Views — orders

## public.active_orders

Pre-joined view of non-cancelled orders with the purchasing user's email and name. Useful for operational dashboards and customer-facing order lookups.

```sql
SELECT
    o.id            AS order_id,
    o.status,
    o.total_amount,
    o.created_at    AS ordered_at,
    o.updated_at,
    u.id            AS user_id,
    u.email         AS user_email,
    u.name          AS user_name
FROM orders o
JOIN users u ON u.id = o.user_id
WHERE o.status <> 'cancelled';
```
