package main

import (
	"JumboDNS/database"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var RedisClient *redis.Client

func MarshalDNSMsg(msg *dns.Msg) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(msg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnmarshalDNSMsg(data []byte) (*dns.Msg, error) {
	var msg dns.Msg
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func SaveMsgToRedis(client *redis.Client, key string, msg *dns.Msg) error {
	ctx := context.Background()
	data, err := MarshalDNSMsg(msg)
	if err != nil {
		return err
	}
	err = client.Set(ctx, key, data, time.Second*10).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetMsgFromRedis(client *redis.Client, key string) (*dns.Msg, error) {
	ctx := context.Background()
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return UnmarshalDNSMsg(data)
}

func resolver(domain string, qtype uint16) []dns.RR {
	savedMsg, _ := GetMsgFromRedis(RedisClient, domain)

	if savedMsg != nil {
		log.Println("Got cache")
		return savedMsg.Answer
	} else {
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domain), qtype)
		m.RecursionDesired = true

		c := &dns.Client{Timeout: 5 * time.Second}

		response, _, err := c.Exchange(m, "8.8.8.8:53")
		if err != nil {
			log.Fatalf("[ERROR] : %v\n", err)
			return nil
		}

		if response == nil {
			log.Fatalf("[ERROR] : no response from server\n")
			return nil
		}

		for _, answer := range response.Answer {
			fmt.Printf("%s\n", answer.String())
		}

		err = SaveMsgToRedis(RedisClient, domain, response)
		if err != nil {
			fmt.Println("Error saving to Redis:", err)
			return nil
		}
		fmt.Println("Successfully saved to Redis")

		return response.Answer
	}
}

type dnsHandler struct{}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		answers := resolver(question.Name, question.Qtype)
		msg.Answer = append(msg.Answer, answers...)
	}

	err := w.WriteMsg(msg)
	if err != nil {
		return
	}
}

func StartDNSServer() {
	handler := new(dnsHandler)
	server := &dns.Server{
		Addr:      ":53",
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	fmt.Printf("JumboDNS started DNS server on %s", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err.Error())
	}
}

func main() {
	gob.Register(&dns.A{})
	gob.Register(&dns.CNAME{})
	gob.Register(&dns.SOA{})
	gob.Register(&dns.PTR{})
	gob.Register(&dns.MX{})
	gob.Register(&dns.TXT{})
	gob.Register(&dns.SRV{})
	gob.Register(&dns.NS{})
	gob.Register(&dns.AAAA{})
	gob.Register(&dns.OPT{})

	RedisClient = database.CacheClient()
	StartDNSServer()
}
