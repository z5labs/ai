---
name: pg-evaldb
description: Query the live evaldb PostgreSQL 17 database. Two schemas (public, analytics) covering users, products, orders, and order_items with a composite FK on (order_id, status), plus an analytics.events table with jsonb payloads. Use when answering questions about evaldb data, schema, or writing SQL against it.
---

# pg-evaldb

Read-only knowledge for working with the live `evaldb` PostgreSQL 17 database fixture. The schema below was introspected from the live server. If it drifts, regenerate this skill (regeneration wipes and rewrites the entire `pg-evaldb/` directory).

## Connection

The five libpq env vars are expected to be exported in the user's shell:

```
PGHOST=127.0.0.1
PGPORT=46065
PGUSER=evaluser
PGPASSWORD=evalpass
PGDATABASE=evaldb
```

There is no host-side `psql`. Use `podman` with the official postgres image and `--network=host` (also exported as `PG_DOCKER_ARGS=--network=host`):

```bash
podman run --rm --network=host -e PGPASSWORD="$PGPASSWORD" \
  docker.io/library/postgres:16 \
  psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" \
  -c "<your SQL here>"
```

Pass `PGPASSWORD` via `-e`, never embed it in a SQL string or commit it.

## Schemas

Two non-system schemas:

- `public` - application tables (`users`, `products`, `orders`, `order_items`)
- `analytics` - `events` (jsonb payloads, FK to `users`)

## Custom types

- `public.order_status` - enum: `pending`, `paid`, `shipped`, `cancelled`

See `references/tables.md` for full column-level table details, indexes, constraints, and foreign keys.

## Things to watch out for

- `order_items` has a composite FK `(order_id, order_status) -> orders(id, status)`. When you insert or update an `order_items` row you must carry the parent order's current `status` - you cannot just reference `order_id`. The matching unique constraint on the parent is `orders_id_status_uniq` on `(id, status)`.
- `orders.status` defaults to `'pending'::order_status`; cast string literals to the enum when comparing.
- All timestamp columns are `timestamp with time zone` and default to `now()`.
- `analytics.events.user_id` is nullable (anonymous events allowed); the FK is still enforced when present.
- `analytics.events.payload` is `jsonb NOT NULL DEFAULT '{}'::jsonb` - use `->`, `->>`, and `@>` operators, not string ops.
- `products.price_cents` has `CHECK (price_cents >= 0)` and is stored in cents (bigint), not dollars.
- `order_items.quantity` has `CHECK (quantity > 0)`.

## Sample queries

Top buyers by paid-order count:

```sql
SELECT u.id, u.email, COUNT(*) AS paid_orders
FROM public.users u
JOIN public.orders o ON o.user_id = u.id
WHERE o.status = 'paid'
GROUP BY u.id, u.email
ORDER BY paid_orders DESC;
```

Order totals (cents) including line items:

```sql
SELECT o.id, o.status, SUM(p.price_cents * oi.quantity) AS total_cents
FROM public.orders o
JOIN public.order_items oi ON oi.order_id = o.id
JOIN public.products p ON p.id = oi.product_id
GROUP BY o.id, o.status
ORDER BY o.id;
```

Page-view paths from analytics:

```sql
SELECT payload->>'path' AS path, COUNT(*) AS views
FROM analytics.events
WHERE event_type = 'page_view'
GROUP BY payload->>'path'
ORDER BY views DESC;
```
