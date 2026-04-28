package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const lookupURL = "https://lookup.eats365pos.net/mpos/v1/login"

func CheckLoginRequest(deviceKey string, uuid string) (*CheckLoginResponse, error) {

	fmt.Println("\n-------------------------Check Login Request--------------------------------")
	method := "POST"

	// Example inputs
	jsonPayload := []byte(fmt.Sprintf(`{"deviceKeyChain":"%s"}`, deviceKey))

	encryptedPayload, err := encryptV2_3(jsonPayload, uuid)

	if err != nil {
		return nil, fmt.Errorf("error encrypting request:%s", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, lookupURL, bytes.NewBuffer([]byte(encryptedPayload)))
	if err != nil {
		return nil, fmt.Errorf("error creating request:%s", err)
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

		return nil, fmt.Errorf("error reading response:%s", err)
	}

	decryptedRes, err := decryptV2_3(body, uuid)

	//fmt.Println(fmt.Sprintf("resp.status: %s", resp.Status))
	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("response.status != 200: %s, decrypted res: %s", resp.Status, decryptedRes)
	} else {
		var res *CheckLoginResponse
		err := json.Unmarshal([]byte(decryptedRes), &res)
		if err != nil {
			return nil, fmt.Errorf("error while unmarshal response:%s", err)
		}

		if res.MessageCode != 0 {
			return nil, fmt.Errorf("res.MessageCode != 0: %d, body: %s", res.MessageCode, decryptedRes)
		}

		return res, nil
	}
}

type CheckLoginResponse struct {
	CountryCode string `json:"countryCode"`
	MPosURL     string `json:"mPosURL"`
	BaseURL     string `json:"baseURL"`
	AccessToken string `json:"accessToken"`
	MessageCode int    `json:"messageCode"`
}
