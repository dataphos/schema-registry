package janitor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func Encrypt(plaintext []byte, encryptionKey string) ([]byte, error) {
	// NewCipher returns a new cipher.Block in dependence to the key which has to be either 16, 24 or 32 characters long
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	// GCM instance generation in dependence to the given cipher block
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// initialization of the nonce size
	nonce := make([]byte, gcm.NonceSize())
	// io.ReadFull ensures that the nonce buffer is filled exactly with the specified number of random bytes
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// return nonce with encrypted message appended
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(encrypted []byte, encryptionKey string) ([]byte, error) {
	// NewCipher returns a new cipher.Block in dependence to the key which has to be either 16, 24 or 32 characters long
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	// gcm instance generation in dependence to the given cipher block
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// reading the nonce and encrypted
	nonceSize := gcm.NonceSize()
	var nonce []byte
	nonce, encrypted = encrypted[:nonceSize], encrypted[nonceSize:]

	// returns decrypted text except if the key is invalid or the nonce is too short for the cipher
	decrypted, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, err
}
