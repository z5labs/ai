-- Seed schema for the postgres-skill-creator e2e eval suite.
--
-- Designed so that a single introspection run exercises every code path the
-- generator's reference-file emission has to handle:
--   - Tables across two schemas (public + analytics) so schema-qualification
--     is real, not theoretical.
--   - Foreign keys that span schemas so relationships.md has cross-schema arrows.
--   - A composite (multi-column) FK so the unnest WITH ORDINALITY pairing in
--     introspect.sh's foreign_keys query is exercised — a same-named-column
--     two-column key would silently degrade to a cross-product on a buggy
--     introspection rewrite.
--   - A view so views.md is generated.
--   - An enum used as a column type so enums.md is generated.
--   - Non-PK indexes so indexes.md has user-defined entries (not just the PK
--     indexes Postgres makes automatically).
--   - Comments on a table and a column so the comment-merging path produces
--     non-empty content.
--   - Seeded rows in known counts so the grader's "run query.sh against the
--     fixture" assertion has stable expected values.

CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'cancelled');

CREATE TABLE public.users (
  id          bigserial PRIMARY KEY,
  email       text NOT NULL UNIQUE,
  full_name   text,
  created_at  timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE public.users IS 'End users of the storefront.';
COMMENT ON COLUMN public.users.email IS 'Lower-cased on insert by the application layer.';

CREATE TABLE public.products (
  id          bigserial PRIMARY KEY,
  sku         text NOT NULL UNIQUE,
  name        text NOT NULL,
  price_cents bigint NOT NULL CHECK (price_cents >= 0)
);

CREATE TABLE public.orders (
  id          bigserial PRIMARY KEY,
  user_id     bigint NOT NULL REFERENCES public.users(id),
  status      order_status NOT NULL DEFAULT 'pending',
  placed_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX orders_user_id_idx ON public.orders(user_id);
CREATE INDEX orders_status_placed_at_idx ON public.orders(status, placed_at);

-- order_items has a (order_id, line_no) PK and a composite FK back to a
-- composite UNIQUE on orders so the introspection query has to pair the two
-- columns correctly. The "natural" route would be a single-column FK on
-- order_id, but that wouldn't exercise the unnest WITH ORDINALITY path.
ALTER TABLE public.orders ADD CONSTRAINT orders_id_status_uniq UNIQUE (id, status);

CREATE TABLE public.order_items (
  order_id     bigint NOT NULL,
  line_no      int    NOT NULL,
  order_status order_status NOT NULL,
  product_id   bigint NOT NULL REFERENCES public.products(id),
  quantity     int    NOT NULL CHECK (quantity > 0),
  PRIMARY KEY (order_id, line_no),
  CONSTRAINT order_items_order_fk
    FOREIGN KEY (order_id, order_status) REFERENCES public.orders(id, status)
);

CREATE TABLE analytics.events (
  id         bigserial PRIMARY KEY,
  user_id    bigint REFERENCES public.users(id),
  event_type text NOT NULL,
  payload    jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX events_user_id_idx ON analytics.events(user_id);

CREATE VIEW public.active_users AS
  SELECT u.id, u.email, count(o.id) AS order_count
  FROM public.users u
  LEFT JOIN public.orders o ON o.user_id = u.id
  GROUP BY u.id, u.email
  HAVING count(o.id) > 0;

COMMENT ON VIEW public.active_users IS 'Users who have placed at least one order.';

INSERT INTO public.users (email, full_name) VALUES
  ('alice@example.com', 'Alice Anderson'),
  ('bob@example.com',   'Bob Brown'),
  ('carol@example.com', 'Carol Carter');

INSERT INTO public.products (sku, name, price_cents) VALUES
  ('SKU-001', 'Widget',   500),
  ('SKU-002', 'Gadget',  1500),
  ('SKU-003', 'Gizmo',   2500);

INSERT INTO public.orders (user_id, status) VALUES
  (1, 'paid'),
  (1, 'shipped'),
  (2, 'pending');

INSERT INTO public.order_items (order_id, line_no, order_status, product_id, quantity) VALUES
  (1, 1, 'paid',    1, 2),
  (1, 2, 'paid',    2, 1),
  (2, 1, 'shipped', 3, 1),
  (3, 1, 'pending', 1, 1);

INSERT INTO analytics.events (user_id, event_type, payload) VALUES
  (1, 'page_view', '{"path":"/"}'),
  (1, 'page_view', '{"path":"/products/1"}'),
  (2, 'page_view', '{"path":"/"}');
