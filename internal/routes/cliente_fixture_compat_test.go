package routes

import (
	"context"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/repository"
)

// Completa el contrato que develop ya exige al doble compartido de pruebas.
func (r *memoriaCuentaClienteRepository) GetClienteActivoByID(_ context.Context, id int) (*models.Cliente, error) {
	for _, cuenta := range r.cuentas {
		if cuenta.Cliente.ID == id && cuenta.Cliente.Activo && cuenta.Cuenta.Activo {
			return &cuenta.Cliente, nil
		}
	}
	return nil, repository.ErrCuentaClienteNoEncontrada
}
