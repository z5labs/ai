# Views

## public.active_users

Users who have placed at least one order.

```sql
 SELECT u.id,
    u.email,
    count(o.id) AS order_count
   FROM users u
     LEFT JOIN orders o ON o.user_id = u.id
  GROUP BY u.id, u.email
 HAVING count(o.id) > 0;
```
