package core

import (
	"github.com/bytedance/sonic"
	"io/ioutil"
	"os"
)

func LoadDNSRecords(filename string) (*DNSRecords, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var dnsRecords DNSRecords
	err = sonic.Unmarshal(data, &dnsRecords)
	if err != nil {
		return nil, err
	}

	return &dnsRecords, nil
}
