# Indexes

public.order_items: CREATE UNIQUE INDEX order_items_pkey ON public.order_items USING btree (id)
public.order_items: CREATE INDEX order_items_order_id_idx ON public.order_items USING btree (order_id)
public.order_items: CREATE INDEX order_items_product_id_idx ON public.order_items USING btree (product_id)
public.orders: CREATE UNIQUE INDEX orders_pkey ON public.orders USING btree (id)
public.orders: CREATE INDEX orders_user_id_idx ON public.orders USING btree (user_id)
public.orders: CREATE INDEX orders_status_idx ON public.orders USING btree (status)
public.orders: CREATE INDEX orders_created_at_idx ON public.orders USING btree (created_at DESC)
public.products: CREATE UNIQUE INDEX products_pkey ON public.products USING btree (id)
public.users: CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id)
public.users: CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email)
