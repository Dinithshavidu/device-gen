package login

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

func encryptV2_3(k2 []byte, strValue string) (string, error) {
	encodedK2 := base64.StdEncoding.EncodeToString(k2)

	split := strings.Split(strValue, "-")
	if len(split) < 5 {
		return "", errors.New("invalid input string format")
	}

	encryptedPart, err := encrypt(encodedK2, split[4])
	if err != nil {
		return "", err
	}

	return split[0] + encryptedPart + split[4], nil
}

func encrypt(bArr string, strValue string) (string, error) {
	derivedKey, err := deriveKey(strValue + "qMm3@-C^YmdCjxAR")
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(derivedKey))
	if err != nil {
		return "", err
	}

	paddedInput := pad([]byte(bArr), aes.BlockSize)
	encrypted := make([]byte, len(paddedInput))
	ecbEncrypt(block, encrypted, paddedInput)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func decryptV2_3(encryptedPart []byte, strValue string) (string, error) {
	keyPart := strings.Split(strValue, "-")[0]
	return decrypt(encryptedPart, keyPart)
}

func decrypt(encryptedData []byte, strValue string) (string, error) {
	derivedKey, err := deriveKey(strValue + "qMm3@-C^YmdCjxAR")
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(derivedKey))
	if err != nil {
		return "", err
	}

	decodedData, err := base64.StdEncoding.DecodeString(string(encryptedData))
	if err != nil {
		return "", err
	}

	decryptedData, err := aesECBDecrypt(block, decodedData)
	if err != nil {
		return "", err
	}

	return string(decryptedData), nil
}

func deriveKey(input string) (string, error) {
	hash := sha1Hash(input)
	midIndex := len(hash) / 2
	if midIndex-8 < 0 || midIndex+8 > len(hash) {
		return "", errors.New("invalid derived key index range")
	}
	return hash[midIndex-8 : midIndex+8], nil
}

func sha1Hash(input string) string {
	hasher := sha1.New()
	hasher.Write([]byte(input))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func ecbEncrypt(block cipher.Block, dst, src []byte) {
	if len(src)%block.BlockSize() != 0 {
		panic("input not a multiple of block size")
	}
	if len(dst) < len(src) {
		panic("output buffer is too small")
	}

	for i := 0; i < len(src); i += block.BlockSize() {
		block.Encrypt(dst[i:i+block.BlockSize()], src[i:i+block.BlockSize()])
	}
}

func aesECBDecrypt(block cipher.Block, ciphertext []byte) ([]byte, error) {
	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	decrypted := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += block.BlockSize() {
		block.Decrypt(decrypted[i:i+block.BlockSize()], ciphertext[i:i+block.BlockSize()])
	}

	return pkcs5Unpad(decrypted), nil
}

func pad(input []byte, blockSize int) []byte {
	paddingSize := blockSize - (len(input) % blockSize)
	padding := bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)
	return append(input, padding...)
}

func pkcs5Unpad(data []byte) []byte {
	length := len(data)
	padding := int(data[length-1])
	return data[:(length - padding)]
}
