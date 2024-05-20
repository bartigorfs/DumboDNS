package core

import "github.com/redis/go-redis/v9"

var (
	LocalRecords *DNSRecords
	RCC          *redis.Client
)

type DNSRecord struct {
	Name string `json:"name"`
	Type string `json:"type"`
	IP   string `json:"ip"`
}

type CNAMERecord struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Target string `json:"target"`
}

type DNSRecords struct {
	ARecords    []DNSRecord   `json:"a_records"`
	AAAARecords []DNSRecord   `json:"aaaa_records"`
	CNAMERecord []CNAMERecord `json:"cname_records"`
}

type DnsHandler struct{}
