# Indexes

analytics.events: CREATE UNIQUE INDEX events_pkey ON analytics.events USING btree (id)
analytics.events: CREATE INDEX events_user_id_idx ON analytics.events USING btree (user_id)
public.order_items: CREATE UNIQUE INDEX order_items_pkey ON public.order_items USING btree (order_id, line_no)
public.orders: CREATE UNIQUE INDEX orders_id_status_uniq ON public.orders USING btree (id, status)
public.orders: CREATE UNIQUE INDEX orders_pkey ON public.orders USING btree (id)
public.orders: CREATE INDEX orders_status_placed_at_idx ON public.orders USING btree (status, placed_at)
public.orders: CREATE INDEX orders_user_id_idx ON public.orders USING btree (user_id)
public.products: CREATE UNIQUE INDEX products_pkey ON public.products USING btree (id)
public.products: CREATE UNIQUE INDEX products_sku_key ON public.products USING btree (sku)
public.users: CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email)
public.users: CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id)
