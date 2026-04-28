package api

import (
	"encoding/json"
	"fmt"
)

type obtainMerchantListRequest struct {
	AccessToken    string `json:"accessToken"`
	DeviceKeyChain string `json:"deviceKeyChain"`
	DisplayName    string `json:"displayName"`
	Service        string `json:"service"`
}

type ObtainMerchantListResponse struct {
	TelephoneCode string     `json:"telephoneCode"`
	Restaurant    Restaurant `json:"restaurant"`
	MerchantList  []Merchant `json:"merchantList"`
	MessageCode   int        `json:"messageCode"`
	Timestamp     int64      `json:"timestamp"`
	ServerInfo    ServerInfo `json:"serverInfo"`
}

type Restaurant struct {
	RestaurantCode string      `json:"restaurantCode"`
	DisplayName    DisplayName `json:"displayName"`
	IconURL        string      `json:"iconURL"`
}

type DisplayName struct {
	TC string `json:"tc"`
}

type Merchant struct {
	MerchantUID   int           `json:"merchantUID"`
	Username      string        `json:"username"`
	SecurityGroup SecurityGroup `json:"securityGroup"`
	NickName      string        `json:"nickName"`
}

type SecurityGroup struct {
	SecurityGroupUID          int  `json:"securityGroupUID"`
	IsWaiterAppAccessible     bool `json:"isWaiterAppAccessible"`
	IsCustomItemDisabled      bool `json:"isCustomItemDisabled"`
	IsAbleToViewAllItems      bool `json:"isAbleToViewAllItems"`
	IsAllPriceTiersAccessible bool `json:"isAllPriceTiersAccessible"`
}

type ServerInfo struct {
	TimeUsed  int    `json:"timeUsed"`
	Instance  string `json:"instance"`
	Timestamp int64  `json:"timestamp"`
}

func ObtainMerchantList(deviceKey string, uuid string, accessKey string, baseurl string) (*ObtainMerchantListResponse, error) {
	fmt.Println("\n-------------------------Obtain Merchant List--------------------------------")
	jsonPayload, err := json.Marshal(obtainMerchantListRequest{
		AccessToken:    accessKey,
		DeviceKeyChain: deviceKey,
		DisplayName:    "Qlub",
		Service:        "obtainMerchantList",
	})

	if err != nil {
		return nil, err
	}

	resByteArr, err := V2WssJsonApiCall(baseurl, uuid, jsonPayload, make(map[string]string))
	if err != nil {
		return nil, err
	}

	res := &ObtainMerchantListResponse{}
	err = json.Unmarshal(resByteArr, &res)
	if err != nil {
		return nil, err
	}

	if res.MessageCode != 0 {
		return nil, fmt.Errorf("response message code != 0:%d", res.MessageCode)
	}

	return res, nil
}
