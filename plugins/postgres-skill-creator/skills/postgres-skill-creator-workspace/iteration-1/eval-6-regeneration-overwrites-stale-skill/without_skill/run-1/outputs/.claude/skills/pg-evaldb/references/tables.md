# evaldb tables

Introspected from the live PostgreSQL 17 fixture. Two non-system schemas: `public` and `analytics`.

## Custom types

### public.order_status (enum)

Values, in declared order: `pending`, `paid`, `shipped`, `cancelled`.

---

## public.users

| column      | type                       | nullable | default                              |
|-------------|----------------------------|----------|--------------------------------------|
| id          | bigint                     | not null | `nextval('users_id_seq'::regclass)`  |
| email       | text                       | not null |                                      |
| full_name   | text                       | yes      |                                      |
| created_at  | timestamp with time zone   | not null | `now()`                              |

- PK: `users_pkey` on `(id)`
- Unique: `users_email_key` on `(email)`
- Referenced by: `orders.user_id`, `analytics.events.user_id`

Row count at generation time: 3.

## public.products

| column        | type   | nullable | default                                 |
|---------------|--------|----------|-----------------------------------------|
| id            | bigint | not null | `nextval('products_id_seq'::regclass)`  |
| sku           | text   | not null |                                         |
| name          | text   | not null |                                         |
| price_cents   | bigint | not null |                                         |

- PK: `products_pkey` on `(id)`
- Unique: `products_sku_key` on `(sku)`
- Check: `products_price_cents_check` -> `price_cents >= 0`
- Referenced by: `order_items.product_id`

Row count at generation time: 3.

## public.orders

| column     | type                       | nullable | default                                |
|------------|----------------------------|----------|----------------------------------------|
| id         | bigint                     | not null | `nextval('orders_id_seq'::regclass)`   |
| user_id    | bigint                     | not null |                                        |
| status     | public.order_status        | not null | `'pending'::order_status`              |
| placed_at  | timestamp with time zone   | not null | `now()`                                |

- PK: `orders_pkey` on `(id)`
- Unique: `orders_id_status_uniq` on `(id, status)` - exists specifically to back the composite FK from `order_items`
- Index: `orders_status_placed_at_idx` on `(status, placed_at)`
- Index: `orders_user_id_idx` on `(user_id)`
- FK: `orders_user_id_fkey` -> `users(id)`
- Referenced by: `order_items` via composite FK on `(id, status)`

Row count at generation time: 3.

## public.order_items

| column         | type                  | nullable | default |
|----------------|-----------------------|----------|---------|
| order_id       | bigint                | not null |         |
| line_no        | integer               | not null |         |
| order_status   | public.order_status   | not null |         |
| product_id     | bigint                | not null |         |
| quantity       | integer               | not null |         |

- PK: `order_items_pkey` on `(order_id, line_no)`
- Check: `order_items_quantity_check` -> `quantity > 0`
- FK: `order_items_order_fk` -> `orders(id, status)` via columns `(order_id, order_status)` (composite)
- FK: `order_items_product_id_fkey` -> `products(id)`

Row count at generation time: 4.

## analytics.events

| column        | type                       | nullable | default                                          |
|---------------|----------------------------|----------|--------------------------------------------------|
| id            | bigint                     | not null | `nextval('analytics.events_id_seq'::regclass)`   |
| user_id       | bigint                     | yes      |                                                  |
| event_type    | text                       | not null |                                                  |
| payload       | jsonb                      | not null | `'{}'::jsonb`                                    |
| occurred_at   | timestamp with time zone   | not null | `now()`                                          |

- PK: `events_pkey` on `(id)`
- Index: `events_user_id_idx` on `(user_id)`
- FK: `events_user_id_fkey` -> `public.users(id)` (nullable)

Row count at generation time: 3.
