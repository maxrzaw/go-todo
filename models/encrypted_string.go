package models

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"

	"github.com/mergermarket/go-pkcs7"
	"gorm.io/gorm/schema"
)

type EncryptedString string

var key []byte

// ctx: contains request-scoped values
// field: the field using the serializer, contains GORM settings, struct tags
// dst: current model value
// dbValue: current field's value in database
func (es *EncryptedString) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	if key == nil {
		panic("Encryption key not set")
	}

	switch value := dbValue.(type) {
	case []byte:
		*es = EncryptedString(DecryptAES(key, string(value)))
	case string:
		*es = EncryptedString(DecryptAES(key, value))
	default:
		return fmt.Errorf("unsupported data %#v", dbValue)
	}
	return nil
}

// ctx: contains request-scoped values
// field: the field using the serializer, contains GORM settings, struct tags
// dst: current model value
// fieldValue: current field's value of the dst
func (es EncryptedString) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if key == nil {
		panic("Encryption key not set")
	}
	return EncryptAES(key, string(es)), nil
}

func SetEncryptionKey(k []byte) {
	// throw an error if key is already set
	if key != nil {
		panic("Encryption key already set")
	}
	key = k
}

func EncryptAES(key []byte, plaintext string) string {
	// create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()

	var paddedPlaintext []byte
	if paddedPlaintext, err = pkcs7.Pad([]byte(plaintext), blockSize); err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, blockSize+len(paddedPlaintext))
	iv := ciphertext[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[blockSize:], paddedPlaintext)
	// encrypt

	fmt.Println("Padded: " + string(paddedPlaintext))
	fmt.Println("encrypted: " + string(ciphertext))
	// return hex string
	return hex.EncodeToString(ciphertext)
}

func DecryptAES(key []byte, ct string) string {
	paddedCiphertext, _ := hex.DecodeString(ct)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(paddedCiphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := paddedCiphertext[:aes.BlockSize]
	paddedCiphertext = paddedCiphertext[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(paddedCiphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(paddedCiphertext, paddedCiphertext)

	var plaintext []byte
	if plaintext, err = pkcs7.Unpad(paddedCiphertext, block.BlockSize()); err != nil {
		panic(err)
	}
	return string(plaintext[:])
}
