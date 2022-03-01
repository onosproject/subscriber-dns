// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package server

import (
	"fmt"

	"github.com/miekg/dns"
)

type StaticRecord struct {
	rr dns.RR
}

func NewStaticRecord(host, addr string) (*StaticRecord, error) {
	rr, err := dns.NewRR(fmt.Sprintf("%s A %s", host, addr))
	if err != nil {
		return nil, err
	}

	return &StaticRecord{
		rr: rr,
	}, nil
}

func (c *StaticRecord) HandleQuery(resp dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) > 0 {
		for _, r := range req.Question {
			log.Info("received query ", r.String())
		}
	}

	m := new(dns.Msg)
	m.SetReply(req)
	m.Authoritative = true
	m.Compress = false

	m.Answer = append(m.Answer, c.rr)
	if len(m.Answer) > 0 {
		for _, r := range m.Answer {
			if r != nil {
				log.Info("sending answer ;", r.String())
			}
		}
	}

	if err := resp.WriteMsg(m); err != nil {
		log.Warn("failed to answer: ", err)
	}
}
