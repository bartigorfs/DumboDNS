package main

import (
	"JumboDNS/database"
	"JumboDNS/forward"
	localdns "JumboDNS/local-dns"
	"context"
	"encoding/gob"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var rcc *redis.Client

type dnsHandler struct{}

func resolver(domain string, qtype uint16) []dns.RR {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	var response *dns.Msg

	//savedMsg, _ := common_utils.GetMsgFromRedis(rcc, domain)
	//
	//if savedMsg != nil {
	//	log.Println("Got cache hit for address ", domain)
	//	response = savedMsg
	//	return response.Answer
	//} else {
	c := &dns.Client{Timeout: 5 * time.Second}

	response = forward.SeekForwarders(c, m)

	if response == nil {
		log.Printf("[ERROR] : no response from server\n")
		return nil
	}

	for _, answer := range response.Answer {
		log.Printf("%s\n", answer.String())
	}

	//err := common_utils.SaveMsgToRedis(rcc, domain, response)
	//if err != nil {
	//	log.Fatalf("Error saving to Redis: %s", err)
	//	return nil
	//}
	//log.Printf("Saved to cache from %s", domain)

	return response.Answer
	//}
	//return nil
}

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

	log.Printf("DumboDNS started DNS server on %s", server.Addr)

	err := server.ListenAndServe()
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
		forward.LoadForwarders(rcc)
		StartDNSServer()
	}()

	go func() {
		LDNSRecords, err := localdns.LoadDNSRecords("temp.json")
		if err != nil {
			log.Fatalf("Error while loading persistent DNS records: %v", err)
		}

		log.Printf("Local DNS records: %s", LDNSRecords)

		server := dns.Server{Addr: ":54", Net: "udp"}
		server.Handler = localdns.LocalHandler(LDNSRecords)

		err = server.ListenAndServe()
		if err != nil {
			log.Fatalf("Error when starting DNS server: %v", err)
		}
	}()

	<-ctxWc.Done()
	log.Println("Shutting down...")
}
