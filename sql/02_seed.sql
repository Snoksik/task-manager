INSERT INTO users (name, email)
SELECT
    'Пользователь ' || n,
    'user' || n || '@example.com'
FROM generate_series(1, 10000) AS n;

INSERT into couriers (name, phone)
SELECT
    'Курьер ' || n,
    'phone ' || '7900' || lpad(n::text, 7, '0')
FROM generate_series(1, 1000) AS n;


    