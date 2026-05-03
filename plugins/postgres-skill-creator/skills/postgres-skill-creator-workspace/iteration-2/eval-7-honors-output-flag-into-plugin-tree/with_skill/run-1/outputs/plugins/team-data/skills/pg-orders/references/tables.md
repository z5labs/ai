# Table Reference — orders

## public.order_items

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('order_items_id_seq') | PK |
| order_id | bigint | NO | | FK → orders.id |
| product_id | bigint | NO | | FK → products.id |
| quantity | integer | NO | | |
| unit_price | numeric(12,2) | NO | | Price at time of purchase |

## public.orders

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('orders_id_seq') | PK |
| user_id | bigint | NO | | FK → users.id |
| status | text | NO | 'pending' | e.g. pending, confirmed, shipped, cancelled |
| total_amount | numeric(12,2) | NO | | Sum of line items |
| created_at | timestamptz | NO | now() | |
| updated_at | timestamptz | NO | now() | |

## public.products

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('products_id_seq') | PK |
| name | text | NO | | |
| description | text | YES | | |
| price | numeric(12,2) | NO | | Current list price |
| stock_quantity | integer | NO | 0 | Units in stock |
| created_at | timestamptz | NO | now() | |

## public.users

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('users_id_seq') | PK |
| email | text | NO | | UNIQUE |
| name | text | NO | | |
| created_at | timestamptz | NO | now() | |
| updated_at | timestamptz | NO | now() | |
