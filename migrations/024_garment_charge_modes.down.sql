DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM prendas WHERE cantidad <> TRUNC(cantidad)
               OR modalidad_cobro_aplicada <> 'UNIDAD'
               OR cantidad > 2147483647) THEN
        RAISE EXCEPTION 'no se puede revertir: se perderian medidas o modalidades historicas';
    END IF;
END $$;

ALTER TABLE prendas DROP CONSTRAINT chk_prenda_medida_modalidad;
ALTER TABLE prendas DROP CONSTRAINT chk_prenda_medida_precision;
ALTER TABLE prendas DROP CONSTRAINT chk_prenda_modalidad_cobro_aplicada;
ALTER TABLE prendas DROP COLUMN modalidad_cobro_aplicada;
ALTER TABLE prendas ALTER COLUMN cantidad TYPE INT USING cantidad::INT;

ALTER TABLE tipos_prenda DROP CONSTRAINT chk_tipo_prenda_modalidad_cobro;
ALTER TABLE tipos_prenda DROP COLUMN modalidad_cobro;
