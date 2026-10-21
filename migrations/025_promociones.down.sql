DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM pedido_promociones) THEN
  RAISE EXCEPTION 'no se puede revertir: se perderian promociones historicas aplicadas';
 END IF;
END $$;
DROP TABLE pedido_promociones;
DROP TABLE promociones;
