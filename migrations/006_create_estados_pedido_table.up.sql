CREATE TABLE estados_pedido (
    id         SERIAL PRIMARY KEY,
    nombre     VARCHAR(30) NOT NULL UNIQUE,
    orden      SMALLINT    NOT NULL UNIQUE,
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP   NOT NULL DEFAULT NOW()
);
