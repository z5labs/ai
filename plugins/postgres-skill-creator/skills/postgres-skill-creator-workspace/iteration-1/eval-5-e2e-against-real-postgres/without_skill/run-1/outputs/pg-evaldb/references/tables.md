# Tables

## analytics.events

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('analytics.events_id_seq') | PK |
| user_id | bigint | YES |  | FK -> public.users.id; indexed (events_user_id_idx) |
| event_type | text | NO |  |  |
| payload | jsonb | NO | '{}'::jsonb |  |
| occurred_at | timestamp with time zone | NO | now() |  |

## public.order_items

| column | type | null | default | notes |
|---|---|---|---|---|
| order_id | bigint | NO |  | PK (composite); FK -> public.orders.id (composite with order_status) |
| line_no | integer | NO |  | PK (composite) |
| order_status | order_status | NO |  | enum; FK -> public.orders.status (composite with order_id) |
| product_id | bigint | NO |  | FK -> public.products.id |
| quantity | integer | NO |  | CHECK (quantity > 0) |

Composite FK: `(order_id, order_status) -> public.orders(id, status)` (constraint `order_items_order_fk`). Joins to `orders` should carry both columns to use the supporting `orders_id_status_uniq` unique index.

## public.orders

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('orders_id_seq') | PK |
| user_id | bigint | NO |  | FK -> public.users.id; indexed (orders_user_id_idx) |
| status | order_status | NO | 'pending'::order_status | enum; participates in unique (id, status); indexed with placed_at |
| placed_at | timestamp with time zone | NO | now() | indexed (status, placed_at) |

Has an extra UNIQUE `(id, status)` (`orders_id_status_uniq`) — required so `order_items` can declare its composite FK against this table.

## public.products

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('products_id_seq') | PK |
| sku | text | NO |  | UNIQUE (products_sku_key) |
| name | text | NO |  |  |
| price_cents | bigint | NO |  | CHECK (price_cents >= 0) |

## public.users

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('users_id_seq') | PK |
| email | text | NO |  | UNIQUE (users_email_key) |
| full_name | text | YES |  |  |
| created_at | timestamp with time zone | NO | now() |  |
