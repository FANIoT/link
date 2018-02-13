package lora

import "time"

// RxMessage contains payloads received from your nodes
type RxMessage struct {
	ApplicationID   string
	ApplicationName string
	DeviceName      string
	DevEUI          string
	FPort           int
	FCnt            int
	RxInfo          []RxInfo
	TxInfo          TxInfo
	Data            []byte
}

// RxInfo contains gateway infomation that payloads
// received from it.
type RxInfo struct {
	Mac     string
	Name    string
	Time    time.Time
	RSSI    int     `json:"rssi"`
	LoRaSNR float64 `json:"LoRaSNR"`
}

// TxInfo contains transmission information
type TxInfo struct {
	Frequency int
	Adr       bool
	CodeRate  string
}
