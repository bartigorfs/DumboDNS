package forward

import (
	"context"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
)

var ForwardersList Forwarders
var wg sync.WaitGroup

type ForwarderResponse struct {
	Name    string
	Message *dns.Msg
}

func LoadForwarders(rcc *redis.Client) {
	ctx := context.Background()
	val, err := rcc.SMembers(ctx, "FORWARDERS_DNS").Result()
	if err != nil {
		panic(err)
	}

	ForwardersList = Forwarders{
		List:  val,
		Count: len(val),
	}
	log.Printf("Loaded forwarders: %s , forwarders count: %d", ForwardersList.List, ForwardersList.Count)
}

func ackFwd(c *dns.Client, m *dns.Msg, eAddr string, wg *sync.WaitGroup, results chan<- ForwarderResponse) {
	defer wg.Done()
	response, _, err := c.Exchange(m, eAddr)
	if err != nil {
		log.Printf("[ERROR] : %v\n", err)
	}
	results <- ForwarderResponse{
		Name:    eAddr,
		Message: response,
	}
}

func SeekForwarders(c *dns.Client, m *dns.Msg) *dns.Msg {
	responses := make(chan ForwarderResponse)

	for i := 0; i < ForwardersList.Count; i++ {
		wg.Add(1)
		go ackFwd(c, m, ForwardersList.List[i], &wg, responses)
	}

	go func() {
		wg.Wait()
		close(responses)
	}()

	for result := range responses {
		if result.Message != nil {
			if result.Message.Answer != nil {
				return result.Message
			}
		}
	}
	return nil
}
