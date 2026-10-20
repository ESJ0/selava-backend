# SEL-109: promociones

La migración aditiva `025_promociones` crea `promociones` y
`pedido_promociones`. No modifica las migraciones 023/024, precios base,
servicios ni pedidos anteriores. Se aplica después de 024 con el mecanismo
habitual de migraciones del proyecto. El down rechaza eliminar aplicaciones
históricas; el rollback no pierde un desglose ya utilizado.

El administrador crea, edita y activa/desactiva paquetes. Recepción consulta y
puede seleccionarlos al crear pedidos. Operarios y tokens cliente no acceden a
estos endpoints administrativos. La vista privada de pedido del cliente dueño
sí devuelve sus promociones aplicadas, con el control de pertenencia existente.

## Contrato y cálculo

`POST /api/pedidos/cotizacion` recibe las mismas `prendas` que el alta y una lista
opcional `promociones: [{tipo_prenda_id, promocion_id}]`. Devuelve grupos por tipo,
opciones elegibles, selección y total. No persiste datos. No se aplica ninguna
promoción automáticamente. Una promoción por tipo evita acumular dos paquetes
sobre la misma medida. Si existen alternativas, el usuario elige explícitamente.

La elegibilidad comprueba tipo activo con tarifa, promoción activa, vigencia
RFC3339 inclusiva, modalidad compatible y medida suficiente. UNIDAD exige
enteros; LIBRA/METRO admiten tres decimales. No existe lógica por nombre.
Precio promocional positivo con dos decimales; fechas inicio <= fin.

El cálculo convierte exactamente las entradas decimales en centavos y milésimas
usando `big.Rat`; multiplica con `big.Int`. No utiliza operaciones float para
calcular dinero. La representación JSON conserva números por compatibilidad.
`aplicaciones = medida_milesimas / paquete_milesimas` mediante división entera;
cubiertas = aplicaciones × paquete; sobrante = disponible − cubiertas.
Promoción = aplicaciones × precio del paquete; normal = tarifa congelada ×
sobrante, redondeado al medio centavo hacia arriba, igual a PostgreSQL numeric.
Se rechazan desbordamientos con 422. Los servicios permanecen sin recargo.

Ejemplos:

| Modalidad | Tarifa normal | Paquete | Medida | Aplicaciones | Sobrante | Total |
| --- | --- | --- | --- | --- | --- | --- |
| UNIDAD | Q25/unidad | 2 por Q40 | 5 | 2 | 1 | Q105 |
| LIBRA | Q12/lb | 5 lb por Q50 | 12.5 lb | 2 | 2.5 lb | Q130 |
| METRO | Q40/m | 2 m por Q60 | 5.5 m | 2 | 1.5 m | Q180 |

## Transacción e históricos

Al crear, el backend vuelve a leer y bloquear tarifas/modalidades y promociones
en la misma transacción que inserta las prendas y sus servicios. Solo acepta
identificadores de selección y medidas; importes, aplicaciones y totales falsos
enviados por el navegador se ignoran. Una selección duplicada, ajena al tipo,
vencida, inactiva o incompatible devuelve 422 y revierte el pedido completo.

Las filas del mismo tipo se agrupan aunque tengan servicios o colores distintos.
Las prendas conservan los snapshots de SEL-107/108 y su subtotal normal. El
snapshot JSON tipado en `pedido_promociones` conserva nombre y descripción,
tipo y nombre, modalidad, paquete, tarifa promocional, tarifa normal, medida
disponible, aplicaciones, cubiertas, sobrante, subtotales e IDs de prendas.
`subtotal_base` y `subtotal_total` también se almacenan como NUMERIC para sumar
el ajuste exacto al total del pedido. El detalle muestra la lista persistida;
nunca consulta la promoción vigente para reconstruir el cobro.

El total se obtiene como suma normal histórica de las prendas menos el importe
base de los grupos promocionados más sus subtotales finales congelados. Agregar
o quitar servicios no pierde ni vuelve a aplicar la promoción. Las prendas
agregadas después se cobran normalmente; no consumen paquetes antiguos ni
modifican su snapshot. No hay endpoint para borrar/editar promociones aplicadas.
Cambiar/desactivar el catálogo no modifica pedidos, pagos ni reportes históricos.

## Frontends y ramas

Admin parte de develop: catálogo de promociones, selector de modalidad derivado
del tipo y cotización del servidor en el resumen. El total no se habilita para
guardar hasta recibir una cotización válida. Se pueden actualizar cotizaciones
y quitar una selección que haya dejado de ser elegible.

Cliente parte de main, como se solicitó. El correctivo
[SEL-108 #15](https://github.com/ESJ0/selava-cliente-frontend/pull/15) incorporó a
main únicamente los dos commits originales de SEL-108 #13 mediante cherry-pick,
conservando autenticación, pedidos, pagos y SEL-107. La rama publicada SEL-109
se actualiza mediante merge de ese main, sin reescribir sus commits existentes.
El detalle conserva medida, modalidad, tarifa y subtotal congelados; en prendas
promocionadas identifica el subtotal normal y muestra por separado la promoción
persistida. No reconstruye importes desde el catálogo ni calcula dinero localmente.

Los PRs de SEL-109 deben permanecer abiertos, sin merge ni auto-merge.
