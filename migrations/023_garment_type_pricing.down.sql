-- No se permite perder instantaneas de pedidos creados con SEL-107.
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM prendas WHERE precio_aplicado IS NOT NULL)
       OR EXISTS (SELECT 1 FROM servicios WHERE precio_base IS NULL) THEN
        RAISE EXCEPTION 'No se puede revertir SEL-107 despues de registrar prendas o servicios nuevos; conservar datos y usar una migracion correctiva';
    END IF;
END $$;
ALTER TABLE servicios ALTER COLUMN precio_base SET NOT NULL;
COMMENT ON COLUMN servicios.precio_base IS NULL;
ALTER TABLE prendas DROP COLUMN precio_aplicado;
DROP TRIGGER tipo_prenda_precio_requerido ON tipos_prenda;
DROP FUNCTION exigir_precio_tipo_prenda();
ALTER TABLE tipos_prenda DROP COLUMN precio;
