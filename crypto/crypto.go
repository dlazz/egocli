package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	DEFAULT_PASSWORD = "Xn2r5u8x/A%D-G+KaPdSgVk[p3s6v9*f"
	TAG              = "!seal"
)

//Secret empty struct
type Secret struct {
}

// Encrypt string
func (s *Secret) Encrypt(secret *string, output *string, password *string) error {

	key := []byte(*password)
	plaintext := []byte(*secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	*output = fmt.Sprintf("%s %s", TAG, base64.URLEncoding.EncodeToString(ciphertext))
	return nil
}

// Decrypt string
func (s *Secret) Decrypt(secret *string, output *string, password *string) error {
	if *password == "" {
		*password = "abcd"
	}

	key := []byte(*password)
	ciphertext, _ := base64.URLEncoding.DecodeString(*secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	*output = string(ciphertext)
	return nil
}
