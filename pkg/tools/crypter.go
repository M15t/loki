package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// AES128Encrypt encrypts data using AES-128-CBC
func AES128Encrypt(origData, key, iv []byte) ([]byte, error) {
	// Validate key length
	if len(key) != 16 {
		return nil, errors.New("key length must be 16 bytes for AES-128")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Validate IV length
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	// Pad the original data to be a multiple of block size
	origData = pkcs5Padding(origData, block.BlockSize())

	// Encrypt the data using AES-128-CBC
	crypted := make([]byte, len(origData))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(crypted, origData)

	return crypted, nil
}

// AES128Decrypt decrypts data using AES-128-CBC
func AES128Decrypt(crypted, key, iv []byte) ([]byte, error) {
	// Validate key length
	if len(key) != 16 {
		return nil, errors.New("key length must be 16 bytes for AES-128")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Validate IV length
	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	// Decrypt the data using AES-128-CBC
	origData := make([]byte, len(crypted))
	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(origData, crypted)

	// Unpad the decrypted data
	origData, err = pkcs5UnPadding(origData)
	if err != nil {
		return nil, err
	}

	return origData, nil
}

// pkcs5Padding pads the input to be a multiple of the block size using PKCS#5
func pkcs5Padding(origData []byte, blockSize int) []byte {
	padding := blockSize - len(origData)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(origData, padText...)
}

// pkcs5UnPadding unpads the input data using PKCS#5
func pkcs5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("invalid padding size")
	}
	unPadding := int(origData[length-1])
	if unPadding > length || unPadding > aes.BlockSize {
		return nil, errors.New("invalid padding")
	}
	return origData[:(length - unPadding)], nil
}
