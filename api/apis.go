package api

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

func hashPassword(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	hashBytes := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hashBytes)
	return hashHex
}

func V2WssJsonApiCall(baseurl string, uuid string, jsonPayload []byte, extraHeaders map[string]string) ([]byte, error) {
	return V2ApiCall(baseurl+"v2/wss/json/", uuid, jsonPayload, extraHeaders)
}

func V2ApiCall(url string, uuid string, jsonPayload []byte, extraHeaders map[string]string) ([]byte, error) {
	method := "POST"

	encryptedPayload := encryptV2(jsonPayload, uuid)

	fmt.Println("Data:", string(jsonPayload))
	fmt.Println("Random uuid:", uuid)

	//fmt.Println("Encrypted req:", encryptedPayload)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(encryptedPayload)))
	if err != nil {
		return nil, fmt.Errorf("error creating request:%s", err)
	}

	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", `Eats365 mPOS/com.eats365.waiter(V1.8.4; Dalvik/2.1.0 (Linux; U; Android 10; SM-T510 Build/QP1A.190711.020)`)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request:%s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err)
	}

	fmt.Println("Response Status:", resp.Status)

	//fmt.Println("Raw body:", string(body))

	decryptedRes := bytes.TrimRight(decryptV2(body, uuid), "\a")

	fmt.Println("Decrypted Body:", string(decryptedRes))

	return decryptedRes, nil
}
