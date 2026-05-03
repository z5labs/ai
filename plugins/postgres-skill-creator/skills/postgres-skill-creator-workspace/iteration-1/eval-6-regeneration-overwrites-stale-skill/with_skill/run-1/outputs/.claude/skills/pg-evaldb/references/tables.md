# Tables

## analytics.events

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('analytics.events_id_seq'::regclass) | PK |
| user_id | bigint | YES |  | FK → public.users.id |
| event_type | text | NO |  |  |
| payload | jsonb | NO | '{}'::jsonb |  |
| occurred_at | timestamp with time zone | NO | now() |  |

## public.order_items

| column | type | null | default | notes |
|---|---|---|---|---|
| order_id | bigint | NO |  | PK; FK → public.orders.id |
| line_no | integer | NO |  | PK |
| order_status | order_status | NO |  | enum order_status; FK → public.orders.status |
| product_id | bigint | NO |  | FK → public.products.id |
| quantity | integer | NO |  |  |

## public.orders

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('orders_id_seq'::regclass) | PK |
| user_id | bigint | NO |  | FK → public.users.id |
| status | order_status | NO | 'pending'::order_status | enum order_status |
| placed_at | timestamp with time zone | NO | now() |  |

## public.products

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('products_id_seq'::regclass) | PK |
| sku | text | NO |  | unique |
| name | text | NO |  |  |
| price_cents | bigint | NO |  |  |

## public.users

End users of the storefront.

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('users_id_seq'::regclass) | PK |
| email | text | NO |  | unique; Lower-cased on insert by the application layer. |
| full_name | text | YES |  |  |
| created_at | timestamp with time zone | NO | now() |  |
