package janitor

import (
	"bytes"
	"testing"
)

func TestEncryption(t *testing.T) {
	message := []byte("Hello, world!")
	encryptionKey := "1Pw1EPV7bx8sk0ugotIkRg=="

	encrypted, err := Encrypt(message, encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := Decrypt(encrypted, encryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(message, decrypted) {
		t.Fatal("expected and actual message not the same")
	}
}
