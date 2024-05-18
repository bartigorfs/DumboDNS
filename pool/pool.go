package pool

import "github.com/miekg/dns"

type Pool struct {
	records map[string]struct{}
}

func (p *Pool) Add(record string) []dns.RR {
	var rec dns.RR
	return rec
}
