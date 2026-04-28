package api

import (
	"encoding/json"
	"fmt"
)

type registerDeviceRequest struct {
	Service        string `json:"service"`
	Version        string `json:"version"`
	OS             string `json:"os"`
	DeviceKeyChain string `json:"deviceKeyChain"`
	OSVersion      string `json:"osVersion"`
	Model          string `json:"model"`
	Manufacturer   string `json:"manufacturer"`
}

type registerDeviceResponse struct {
	MessageCode int `json:"messageCode"`
}

func RegisterDevice(deviceKey, uuid string) error {
	url := "https://glu.eats365.net/v2/waiter/wss/json"

	jsonPayload, err := json.Marshal(registerDeviceRequest{
		Service:        "registerDevice",
		Version:        "1.8.4",
		OS:             "Android",
		DeviceKeyChain: deviceKey,
		OSVersion:      "11",
		Model:          "Pixel 6",
		Manufacturer:   "Google",
	})

	if err != nil {
		return err
	}

	resByteArr, err := V2ApiCall(url, uuid, jsonPayload, make(map[string]string))
	if err != nil {
		return err
	}

	res := &registerDeviceResponse{}
	err = json.Unmarshal(resByteArr, &res)
	if err != nil {
		return err
	}

	if res.MessageCode != 0 {
		return fmt.Errorf("response message code != 0:%d", res.MessageCode)
	}

	return nil
}
