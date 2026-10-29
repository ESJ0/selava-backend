package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenInvalido = errors.New("token inválido o expirado")

const tokenTTL = 8 * time.Hour

const (
	ActorUsuario = "usuario"
	ActorCliente = "cliente"
)

// Claims pertenece exclusivamente a los usuarios internos del panel.
type Claims struct {
	UsuarioID int    `json:"usuario_id"`
	RolID     int    `json:"rol_id"`
	Rol       string `json:"rol,omitempty"`
	Actor     string `json:"actor"`
	jwt.RegisteredClaims
}

// ClienteClaims identifica a un cliente del portal sin otorgarle un rol del
// panel administrativo.
type ClienteClaims struct {
	UsuarioID int    `json:"usuario_id,omitempty"`
	ClienteID int    `json:"cliente_id"`
	RolID     int    `json:"rol_id"`
	Rol       string `json:"rol"`
	Actor     string `json:"actor"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, usuarioID, rolID int) (string, error) {
	return GenerateTokenConRol(secret, usuarioID, rolID, nombreRol(rolID))
}

func GenerateTokenConRol(secret string, usuarioID, rolID int, rol string) (string, error) {
	claims := Claims{
		UsuarioID: usuarioID,
		RolID:     rolID,
		Rol:       rol,
		Actor:     ActorUsuario,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrTokenInvalido
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid || claims.Actor != ActorUsuario || claims.UsuarioID <= 0 || claims.RolID <= 0 {
		return nil, ErrTokenInvalido
	}
	return claims, nil
}

func GenerateClienteToken(secret string, clienteID int) (string, error) {
	return GenerateClienteAccountToken(secret, 0, clienteID)
}

func GenerateClienteAccountToken(secret string, usuarioID, clienteID int) (string, error) {
	claims := ClienteClaims{
		UsuarioID: usuarioID,
		ClienteID: clienteID,
		RolID:     4,
		Rol:       "Cliente",
		Actor:     ActorCliente,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateClienteToken(secret, tokenString string) (*ClienteClaims, error) {
	claims := &ClienteClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrTokenInvalido
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid || claims.Actor != ActorCliente || claims.ClienteID <= 0 || claims.RolID != 4 {
		return nil, ErrTokenInvalido
	}
	return claims, nil
}

func nombreRol(rolID int) string {
	switch rolID {
	case 1:
		return "Administrador"
	case 2:
		return "Recepcionista"
	case 3:
		return "Operario"
	case 4:
		return "Cliente"
	default:
		return ""
	}
}
