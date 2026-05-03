# Tables

## public.order_items

| column | type | null | default | notes |
|---|---|---|---|---|
| order_id | bigint | NO |  | PK; FK → orders.id |
| product_id | bigint | NO |  | PK; FK → products.id |
| quantity | integer | NO | 1 |  |
| unit_price_cents | integer | NO |  |  |

## public.orders

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('orders_id_seq'::regclass) | PK |
| user_id | bigint | NO |  | FK → users.id |
| status | text | NO | 'pending' | order lifecycle: pending, paid, shipped, delivered, cancelled |
| total_cents | integer | NO | 0 |  |
| created_at | timestamptz | NO | now() |  |

## public.products

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('products_id_seq'::regclass) | PK |
| sku | text | NO |  | unique |
| name | text | NO |  |  |
| price_cents | integer | NO |  |  |
| created_at | timestamptz | NO | now() |  |

## public.users

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('users_id_seq'::regclass) | PK |
| email | text | NO |  | unique |
| name | text | YES |  |  |
| created_at | timestamptz | NO | now() |  |
| updated_at | timestamptz | NO | now() |  |
