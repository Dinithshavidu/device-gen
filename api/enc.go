package api

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"strings"
)

var predefinedIVs = [][]byte{
	{188, 24, 206, 185, 143, 135, 85, 116},
	{176, 53, 157, 75, 249, 189, 55, 115},
	{37, 101, 225, 240, 244, 83, 221, 16},
	{122, 137, 204, 232, 102, 142, 89, 24},
	{63, 187, 179, 189, 154, 222, 114, 20},
	{201, 110, 187, 230, 142, 205, 158, 100},
	{30, 51, 61, 224, 208, 140, 87, 197},
	{116, 80, 210, 21, 96, 186, 214, 221},
	{118, 72, 118, 214, 98, 63, 170, 171},
	{190, 135, 24, 146, 168, 202, 66, 169},
	{140, 149, 24, 127, 43, 152, 158, 172},
	{77, 160, 120, 45, 178, 94, 225, 96},
	{212, 116, 108, 40, 184, 219, 161, 78},
	{52, 212, 169, 100, 170, 105, 117, 39},
	{194, 191, 88, 48, 150, 27, 230, 103},
	{71, 142, 84, 70, 46, 254, 94, 37},
	{188, 34, 103, 118, 18, 169, 139, 143},
	{157, 120, 49, 160, 89, 126, 156, 112},
	{214, 150, 225, 169, 72, 73, 178, 44},
	{56, 77, 23, 146, 248, 254, 176, 48},
	{244, 22, 221, 134, 196, 204, 140, 68},
	{206, 247, 132, 217, 184, 158, 203, 250},
	{75, 221, 69, 252, 51, 117, 200, 126},
	{242, 225, 27, 197, 245, 193, 93, 245},
	{207, 101, 65, 23, 49, 232, 183, 208},
	{59, 198, 70, 36, 252, 27, 22, 56},
	{164, 41, 26, 109, 139, 165, 42, 177},
	{187, 175, 222, 149, 104, 29, 215, 23},
	{80, 163, 72, 214, 190, 208, 225, 71},
	{248, 216, 76, 21, 72, 121, 208, 126},
	{87, 205, 243, 65, 21, 92, 173, 137},
	{230, 26, 199, 20, 70, 20, 98, 254},
	{189, 98, 164, 239, 158, 186, 46, 237},
	{153, 225, 121, 223, 160, 78, 78, 96},
	{167, 237, 20, 252, 85, 125, 244, 74},
	{55, 199, 249, 86, 163, 139, 36, 42},
	{89, 57, 44, 24, 118, 216, 33, 167},
	{226, 38, 65, 167, 90, 29, 68, 50},
	{168, 126, 242, 82, 158, 45, 165, 105},
	{193, 73, 73, 48, 50, 230, 111, 80},
	{168, 73, 81, 55, 105, 47, 48, 55},
	{216, 90, 121, 81, 167, 74, 182, 89},
	{110, 180, 160, 86, 148, 49, 146, 120},
	{90, 245, 23, 166, 247, 249, 232, 52},
	{254, 136, 216, 237, 172, 99, 157, 51},
	{189, 66, 243, 16, 46, 40, 215, 39},
	{205, 129, 194, 76, 44, 75, 80, 29},
	{197, 110, 255, 156, 23, 24, 179, 234},
	{47, 137, 70, 81, 194, 73, 45, 94},
	{214, 209, 214, 185, 145, 253, 220, 226},
}

func decrypt(bArr []byte, str2 string) []byte {
	f2 := f(str2)
	key := []byte(f2)
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		log.Println("Error creating cipher:", err)
		return nil
	}

	iv := predefinedIVs[e(f2)]
	mode := cipher.NewCBCDecrypter(block, iv)

	decodedData, err := base64.StdEncoding.DecodeString(string(bArr))
	if err != nil {
		log.Println("Error decoding Base64:", err)
		return nil
	}

	mode.CryptBlocks(decodedData, decodedData)

	decodedData = pkcs5Unpadding(decodedData)

	return decodedData
}

func pkcs5Unpadding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	paddingLen := int(data[len(data)-1])
	if paddingLen > len(data) {
		return data
	}
	return data[:len(data)-paddingLen]
}

func encrypt(bArr []byte, str2 string) string {
	f2 := f(str2)
	key := []byte(f2)
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		log.Println(err)
		return ""
	}

	iv := predefinedIVs[e(f2)]
	mode := cipher.NewCBCEncrypter(block, iv)

	paddedData := pkcs5Padding(bArr, block.BlockSize())
	mode.CryptBlocks(paddedData, paddedData)

	return base64.StdEncoding.EncodeToString(paddedData)
}

func e(str string) int {
	num := 0
	for i := 0; i < len(str); i++ {
		charAt := str[i]
		if charAt >= '0' || charAt <= '9' {
			num += int(charAt) * (i % 3)
		}
	}
	return num % 50
}

func f(str string) string {
	hash := sha256.Sum256([]byte(str))
	hashStr := hex.EncodeToString(hash[:])
	if len(hashStr) >= 24 {
		if len(hashStr) > 24 {
			return hashStr[:24]
		}
		return hashStr
	}
	return f(hashStr + hashStr)
}

func decryptV2(bArr []byte, randomUUID string) []byte {
	return decrypt(bArr, r(randomUUID))
}

func encryptV2(bArr []byte, randomUUID string) string {
	base64EncodedData := base64.StdEncoding.EncodeToString(bArr)
	encrypted := encrypt([]byte(base64EncodedData), s(randomUUID))
	return r(randomUUID) + encrypted + s(randomUUID)
}

func s(str string) string {
	split := strings.Split(str, "-")
	return split[len(split)-1]
}

func r(str string) string {
	return strings.Split(str, "-")[0]
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
