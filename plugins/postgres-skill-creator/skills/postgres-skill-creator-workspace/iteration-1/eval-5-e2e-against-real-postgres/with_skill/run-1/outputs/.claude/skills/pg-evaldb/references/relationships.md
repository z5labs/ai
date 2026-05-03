# Foreign keys

```
analytics.events.user_id -> public.users.id
public.order_items.product_id -> public.products.id
public.order_items.order_id -> public.orders.id
public.order_items.order_status -> public.orders.status
public.orders.user_id -> public.users.id
```

`public.order_items` has a composite foreign key on `(order_id, order_status)` referencing `public.orders(id, status)` — both columns must match the parent row, so joins on a single column alone are not sufficient when reasoning about the integrity guarantee.
