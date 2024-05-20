package main

import (
	core "JumboDNS/core"
	"JumboDNS/database"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var rcc *redis.Client
var LDNSRecords *core.DNSRecords

type dnsHandler struct{}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	for _, question := range r.Question {
		name := question.Name
		switch question.Qtype {
		case dns.TypeA:
			for _, record := range LDNSRecords.ARecords {
				if record.Name == name && record.Type == "A" {
					m.Answer = append(m.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   net.ParseIP(record.IP),
					})
				}
			}
		case dns.TypeAAAA:
			for _, record := range LDNSRecords.AAAARecords {
				if record.Name == name && record.Type == "AAAA" {
					m.Answer = append(m.Answer, &dns.AAAA{
						Hdr:  dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 3600},
						AAAA: net.ParseIP(record.IP),
					})
				}
			}
		}
	}
	log.Println(m)
	err := w.WriteMsg(m)
	if err != nil {
		log.Printf("[ERROR] : %v\n", err)
	}
}

func StartDNSServer() {
	var err error
	LDNSRecords, err = core.LoadDNSRecords("temp.json")
	if err != nil {
		log.Fatalf("Error while loading persistent DNS records: %v", err)
	}

	log.Printf("Local DNS records: %s", LDNSRecords)

	handler := new(dnsHandler)
	server := &dns.Server{
		Addr:      ":53",
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	log.Printf("DumboDNS started DNS server on %s", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n", err.Error())
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

	ctxWc, runCancel := context.WithCancel(context.Background())

	signalWCH := make(chan os.Signal, 1)
	signal.Notify(signalWCH, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalWCH
		log.Printf("Received signal: %v", sig)
		runCancel()
	}()

	go func() {
		rcc = database.RedisCacheClient()
		StartDNSServer()

		go func() {
			ctx := context.Background()
			err := rcc.Publish(ctx, "mychannel1", "test").Err()
			if err != nil {
				log.Fatalf("Error publishing message: %s", err)
			} else {
				log.Printf("Published message to %s: %s", "mychannel1", "test")
			}
			database.LocalDNSSub = rcc.Subscribe(ctx, "DynamicRecords")
			defer func(LocalDNSSub *redis.PubSub) {
				err := LocalDNSSub.Close()
				if err != nil {

				}
			}(database.LocalDNSSub)

			for {
				msg, err := database.LocalDNSSub.ReceiveMessage(ctx)
				if err != nil {
					log.Fatalf("Error receiving message: %s", err)
				}

				fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
			}
		}()
	}()

	<-ctxWc.Done()
	log.Println("Shutting down...")
}
