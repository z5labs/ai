# Foreign-key relationships

```
public.orders.user_id              -> public.users.id
public.order_items.product_id      -> public.products.id
public.order_items.(order_id,
                    order_status)  -> public.orders.(id, status)   [composite FK: order_items_order_fk]
analytics.events.user_id           -> public.users.id              [cross-schema; nullable]
```

Notes:
- `public.users` is the central hub: referenced by both `public.orders` and `analytics.events`.
- The `order_items -> orders` foreign key is **composite on two columns** (`order_id`, `order_status`). Always join on both, e.g.
  `JOIN orders o ON o.id = oi.order_id AND o.status = oi.order_status`.
- `analytics.events.user_id` is **nullable**, so anonymous events exist; use a LEFT JOIN when computing per-user counts that should preserve users with zero events.
