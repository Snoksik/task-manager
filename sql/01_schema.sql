DROP TABLE IF EXISTS 
order_status_history,
deliveries,
orders,
couriers, 
users CASCADE;

CREATE TABLE users(
    id bigserial PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at timestamptz DEFAULT now() NOT NULL 
);


CREATE TABLE couriers(
    id bigserial PRIMARY KEY,
    name text not NULL,
    phone text not NULL UNIQUE,
    is_active boolean not NULL DEFAULT true,
    created_at timestamptz not null DEFAULT now()

);

CREATE TABLE orders(
    id bigserial PRIMARY KEY,
    user_id bigint not null REFERENCES users(id),
    status text not null CHECK (status IN (
    'new',
    'confirmed',
    'assigned',
    'in_delivery',
    'delivered',
    'cancelled'
)),
    address text not null,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE deliveries(
    ID BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id),
    courier_id BIGINT NOT NULL references couriers(id),
    status TEXT NOT NULL check (status in(
    'assigned',
    'picked_up',
    'in_transit',
    'delivered',
    'cancelled'
    )), 
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE order_status_history(
    ID BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    status text not null check (status in(
    'new',
    'confirmed',
    'assigned',
    'in_delivery',
    'delivered',
    'cancelled'
    )),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);