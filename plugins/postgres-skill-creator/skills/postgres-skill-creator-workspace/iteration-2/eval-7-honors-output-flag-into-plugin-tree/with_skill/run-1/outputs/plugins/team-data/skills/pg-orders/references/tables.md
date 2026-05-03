# Tables

## public.order_items

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('order_items_id_seq') | PK |
| order_id | bigint | NO | | FK → orders.id |
| product_id | bigint | NO | | FK → products.id |
| quantity | integer | NO | | |
| unit_price | numeric(10,2) | NO | | |
| created_at | timestamptz | NO | now() | |

## public.orders

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('orders_id_seq') | PK |
| user_id | bigint | NO | | FK → users.id |
| status | text | NO | 'pending' | |
| total_amount | numeric(10,2) | NO | 0.00 | |
| created_at | timestamptz | NO | now() | |
| updated_at | timestamptz | NO | now() | |

## public.products

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('products_id_seq') | PK |
| name | text | NO | | |
| description | text | YES | | |
| price | numeric(10,2) | NO | | |
| stock_quantity | integer | NO | 0 | |
| created_at | timestamptz | NO | now() | |
| updated_at | timestamptz | NO | now() | |

## public.users

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval('users_id_seq') | PK |
| email | text | NO | | unique |
| name | text | NO | | |
| created_at | timestamptz | NO | now() | |
| updated_at | timestamptz | NO | now() | |
