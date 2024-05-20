package core

import "github.com/redis/go-redis/v9"

type DNSRecord struct {
	Name string `json:"name"`
	Type string `json:"type"`
	IP   string `json:"ip"`
}

type DNSRecords struct {
	ARecords    []DNSRecord `json:"a_records"`
	AAAARecords []DNSRecord `json:"aaaa_records"`
}

type DnsHandler struct{}

var LocalRecords *DNSRecords

var RCC *redis.Client
