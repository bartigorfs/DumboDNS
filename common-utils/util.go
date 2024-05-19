package common_utils

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
	"time"
)

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

func SaveMsgToRedis(rcc *redis.Client, key string, msg *dns.Msg) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(msg)
	if err != nil {
		return err
	}

	return rcc.Set(context.Background(), key, buf.Bytes(), 60*time.Second).Err()
}

func GetMsgFromRedis(rcc *redis.Client, key string) (*dns.Msg, error) {
	val, err := rcc.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, err
	}

	var msg dns.Msg
	buf := bytes.NewBuffer(val)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
