-- SEL-109: catálogo independiente. Ninguna tarifa, prenda o pedido anterior se modifica.
CREATE TABLE promociones (
 id SERIAL PRIMARY KEY,
 nombre VARCHAR(100) NOT NULL CHECK (BTRIM(nombre) <> ''),
 descripcion VARCHAR(500) NOT NULL DEFAULT '',
 tipo_prenda_id INT NOT NULL REFERENCES tipos_prenda(id),
 modalidad_cobro VARCHAR(10) NOT NULL CHECK (modalidad_cobro IN ('UNIDAD','LIBRA','METRO')),
 medida_requerida NUMERIC NOT NULL CHECK (medida_requerida > 0 AND medida_requerida <= 99999999.999 AND medida_requerida = ROUND(medida_requerida,3)),
 precio_promocional NUMERIC NOT NULL CHECK (precio_promocional > 0 AND precio_promocional <= 99999999.99 AND precio_promocional = ROUND(precio_promocional,2)),
 fecha_inicio TIMESTAMPTZ NOT NULL,
 fecha_fin TIMESTAMPTZ NOT NULL,
 activo BOOLEAN NOT NULL DEFAULT TRUE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 CHECK (fecha_inicio <= fecha_fin),
 CHECK (modalidad_cobro <> 'UNIDAD' OR medida_requerida = TRUNC(medida_requerida))
);
CREATE INDEX promociones_tipo_vigencia_idx ON promociones(tipo_prenda_id, fecha_inicio, fecha_fin) WHERE activo;

-- Una selección por tipo de prenda en el pedido; todos los paquetes de ese tipo
-- se agrupan, incluso cuando la UI envía una fila por unidad y distintos servicios.
-- snapshot conserva el desglose tipado, nombre/modalidad/tarifas e IDs de prendas.
CREATE TABLE pedido_promociones (
 id SERIAL PRIMARY KEY,
 pedido_id INT NOT NULL REFERENCES pedidos(id),
 promocion_id INT NOT NULL REFERENCES promociones(id),
 tipo_prenda_id INT NOT NULL REFERENCES tipos_prenda(id),
 subtotal_base NUMERIC(10,2) NOT NULL CHECK (subtotal_base >= 0),
 subtotal_total NUMERIC(10,2) NOT NULL CHECK (subtotal_total > 0),
 snapshot JSONB NOT NULL CHECK (jsonb_typeof(snapshot) = 'object'),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 UNIQUE(pedido_id, tipo_prenda_id)
);
CREATE INDEX pedido_promociones_pedido_idx ON pedido_promociones(pedido_id);
