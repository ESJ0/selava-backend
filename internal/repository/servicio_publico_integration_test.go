package repository

import "testing"

func TestServicioPublicoPostgresActivePricesAndOrder(t *testing.T) {
	f := newSprint3Fixture(t)
	_, err := f.db.Exec(f.ctx, `UPDATE servicios SET activo=FALSE WHERE id=$1`, f.services[1])
	mustSprint3(t, err)
	_, err = f.db.Exec(f.ctx, `UPDATE servicios SET descripcion='Desde PostgreSQL', precio_base=73.45, tiempo_estimado_horas=48 WHERE id=$1`, f.services[0])
	mustSprint3(t, err)
	repo := NewServicioRepository(f.db)
	catalog, err := repo.ListActive(f.ctx)
	mustSprint3(t, err)
	if len(catalog) != 2 || catalog[0].ID != f.services[0] || catalog[1].ID != f.services[2] {
		t.Fatalf("filtro/orden inesperado: %+v", catalog)
	}
	if catalog[0].PrecioBase != 73.45 || catalog[1].PrecioBase != 2 || catalog[0].Descripcion == nil || *catalog[0].Descripcion != "Desde PostgreSQL" || catalog[0].TiempoEstimadoHoras == nil || *catalog[0].TiempoEstimadoHoras != 48 {
		t.Fatalf("datos distintos a PostgreSQL: %+v", catalog)
	}
	if catalog[1].Descripcion != nil || catalog[1].TiempoEstimadoHoras != nil {
		t.Fatal("campos opcionales deben conservar NULL")
	}
	_, err = f.db.Exec(f.ctx, `UPDATE servicios SET activo=FALSE`)
	mustSprint3(t, err)
	catalog, err = repo.ListActive(f.ctx)
	mustSprint3(t, err)
	if catalog == nil || len(catalog) != 0 {
		t.Fatalf("esperado arreglo vacio, obtenido: %#v", catalog)
	}
}
