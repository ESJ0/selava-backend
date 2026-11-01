package service

import (
	"context"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

type publicPromoRepo struct{ PromocionRepository }

func (publicPromoRepo) List(context.Context) ([]models.Promocion, error) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	p := func(id int, nombre string, activo bool, inicio, fin time.Time) models.Promocion {
		return models.Promocion{ID: id, TipoPrendaNombre: "Camisa", PromocionRequest: models.PromocionRequest{Nombre: nombre, TipoPrendaID: 1, MedidaRequerida: 2, PrecioPromocional: 40, FechaInicio: inicio, FechaFin: fin, Activo: activo}}
	}
	return []models.Promocion{
		p(5, "Futura", true, end, end.AddDate(0, 1, 0)),
		p(4, "Inactiva", false, start, end),
		p(3, "Dos por cuarenta", true, start, end),
		p(2, "Dos por cuarenta", true, start, end),
		p(1, "Vencida", true, start.AddDate(0, -2, 0), start.AddDate(0, -1, 0)),
		p(6, "Nueva", true, start, end),
	}, nil
}

func TestListPublicasSoloVigentesActivasSinRepetidasYMasNuevasPrimero(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	got, err := NewPromocionService(publicPromoRepo{}).ListPublicas(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != 6 || got[1].ID != 3 {
		t.Fatalf("promociones públicas inesperadas: %+v", got)
	}
}
