package cloud

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := `{"access_key_id":"ak","access_key_secret":"sk"}`
	encrypted, err := encryptAESGCM(key, plaintext)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, encrypted)

	decrypted, err := decryptAESGCM(key, encrypted)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestEncryptRejectsBadKeySize(t *testing.T) {
	_, err := encryptAESGCM([]byte("tooshort"), "data")
	require.ErrorIs(t, err, ErrInvalidEncryptionKey)
}

func TestDecryptRejectsBadKeySize(t *testing.T) {
	_, err := decryptAESGCM([]byte("tooshort"), "data")
	require.ErrorIs(t, err, ErrInvalidEncryptionKey)
}
