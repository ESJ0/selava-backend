package auth

import "testing"

func TestTokensDeUsuarioYClienteNoSonIntercambiables(t *testing.T) {
	const secret = "test-secret"

	usuarioToken, err := GenerateToken(secret, 7, 2)
	if err != nil {
		t.Fatal(err)
	}
	usuarioClaims, err := ValidateToken(secret, usuarioToken)
	if err != nil || usuarioClaims.UsuarioID != 7 || usuarioClaims.Actor != ActorUsuario {
		t.Fatalf("token de usuario invalido: claims=%#v err=%v", usuarioClaims, err)
	}
	if _, err = ValidateClienteToken(secret, usuarioToken); err == nil {
		t.Fatal("un token de usuario fue aceptado como token de cliente")
	}

	clienteToken, err := GenerateClienteToken(secret, 42)
	if err != nil {
		t.Fatal(err)
	}
	clienteClaims, err := ValidateClienteToken(secret, clienteToken)
	if err != nil || clienteClaims.ClienteID != 42 || clienteClaims.Actor != ActorCliente {
		t.Fatalf("token de cliente invalido: claims=%#v err=%v", clienteClaims, err)
	}
	if _, err = ValidateToken(secret, clienteToken); err == nil {
		t.Fatal("un token de cliente fue aceptado como token administrativo")
	}
}
