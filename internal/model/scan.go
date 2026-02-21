package model

import (
	"strconv"
	"time"
)

type MasscanOutput struct {
	IP        string `json:"ip"`
	Timestamp string `json:"timestamp"`
	Ports     []struct {
		Port    int    `json:"port"`
		Proto   string `json:"proto"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		TTL     int    `json:"ttl"`
		Service struct {
			Name   string `json:"name"`
			Banner string `json:"banner"`
		} `json:"service"`
	} `json:"ports"`
}

type ScanResult struct {
	IP        string
	Port      int
	Proto     string
	Banner    string
	FirstSeen time.Time
	LastSeen  time.Time
}

func (r ScanResult) Key() string {
	return r.IP + ":" + strconv.Itoa(r.Port) + "/" + r.Proto
}
