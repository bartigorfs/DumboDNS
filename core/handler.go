package core

import (
	"github.com/miekg/dns"
	"log"
	"net"
)

func (h *DnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	for _, question := range r.Question {
		name := question.Name
		switch question.Qtype {
		case dns.TypeA:
			for _, record := range LocalRecords.ARecords {
				if record.Name == name && record.Type == "A" {
					m.Answer = append(m.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600},
						A:   net.ParseIP(record.IP),
					})
				}
			}
		case dns.TypeCNAME:

		case dns.TypeAAAA:
			for _, record := range LocalRecords.AAAARecords {
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
