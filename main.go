package main

import (
	"JumboDNS/database"
	"context"
	"fmt"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var RedisClient *redis.Client
var ctx = context.Background()

func resolver(domain string, qtype uint16) []dns.RR {
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

	fmt.Println(response.Answer)

	return response.Answer
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
	RedisClient = database.CacheClient()
	err := RedisClient.Set(ctx, "key", "value", time.Second*300).Err()
	if err != nil {
		panic(err)
	}
	StartDNSServer()
}
