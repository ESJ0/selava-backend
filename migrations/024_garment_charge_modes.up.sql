-- La modalidad pertenece al catalogo; el backend nunca la infiere del nombre.
ALTER TABLE tipos_prenda
    ADD COLUMN modalidad_cobro VARCHAR(10) NOT NULL DEFAULT 'UNIDAD';
ALTER TABLE tipos_prenda
    ADD CONSTRAINT chk_tipo_prenda_modalidad_cobro
    CHECK (modalidad_cobro IN ('UNIDAD', 'LIBRA', 'METRO'));

-- Configuracion inicial de datos. Si los tipos aun no existen, el seed de
-- desarrollo los crea; en produccion deben crearse con su tarifa real.
UPDATE tipos_prenda SET modalidad_cobro = 'LIBRA'
WHERE LOWER(BTRIM(nombre)) = 'canasta';
UPDATE tipos_prenda SET modalidad_cobro = 'METRO'
WHERE LOWER(BTRIM(nombre)) = 'alfombra';

-- NUMERIC sin escala evita redondeos silenciosos. La restriccion admite hasta
-- tres decimales para peso y longitud, y exige enteros para cobro por unidad.
ALTER TABLE prendas ALTER COLUMN cantidad TYPE NUMERIC USING cantidad::NUMERIC;
ALTER TABLE prendas ADD COLUMN modalidad_cobro_aplicada VARCHAR(10);
UPDATE prendas p
SET modalidad_cobro_aplicada = tp.modalidad_cobro
FROM tipos_prenda tp
WHERE tp.id = p.tipo_prenda_id;
ALTER TABLE prendas ALTER COLUMN modalidad_cobro_aplicada SET NOT NULL;
ALTER TABLE prendas ADD CONSTRAINT chk_prenda_modalidad_cobro_aplicada
    CHECK (modalidad_cobro_aplicada IN ('UNIDAD', 'LIBRA', 'METRO'));
ALTER TABLE prendas ADD CONSTRAINT chk_prenda_medida_precision
    CHECK (cantidad > 0 AND cantidad <= 99999999.999
           AND cantidad = ROUND(cantidad, 3));
ALTER TABLE prendas ADD CONSTRAINT chk_prenda_medida_modalidad
    CHECK (modalidad_cobro_aplicada <> 'UNIDAD' OR cantidad = TRUNC(cantidad));

COMMENT ON COLUMN tipos_prenda.modalidad_cobro IS 'Modalidad vigente: UNIDAD, LIBRA o METRO.';
COMMENT ON COLUMN prendas.cantidad IS 'Medida cobrada: unidades, libras o metros segun modalidad_cobro_aplicada.';
COMMENT ON COLUMN prendas.modalidad_cobro_aplicada IS 'Modalidad congelada al registrar la prenda.';
