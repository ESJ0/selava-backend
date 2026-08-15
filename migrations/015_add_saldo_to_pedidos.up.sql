ALTER TABLE pedidos
ADD COLUMN saldo NUMERIC(10,2) NOT NULL DEFAULT 0;

-- Conserva el saldo real de los pedidos que ya existian antes de esta migracion.
UPDATE pedidos AS pedido
SET saldo = GREATEST(
    pedido.total - COALESCE(
        (SELECT SUM(pago.monto) FROM pagos AS pago WHERE pago.pedido_id = pedido.id),
        0
    ),
    0
);

ALTER TABLE pedidos
ADD CONSTRAINT pedidos_saldo_no_negativo CHECK (saldo >= 0);
