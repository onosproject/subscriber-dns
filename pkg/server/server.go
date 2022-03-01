// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package server

import (
	"context"

	"github.com/miekg/dns"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/subscriber-dns/pkg/config"
)

var log = logging.GetLogger("subsriber-dns")

type Server struct {
	*dns.Server
	Mux           *dns.ServeMux
	Bind          string
	Protocol      string
	Domain        string
	CustomRecords []config.CustomRecord
	Roc           config.Roc
	SiteId        string
	Notify        func()
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	s.Mux = dns.NewServeMux()

	for _, cr := range s.CustomRecords {
		if sr, err := NewStaticRecord(cr.Host, cr.Addr); err == nil {
			s.Mux.HandleFunc(sr.rr.Header().Name, sr.HandleQuery)
		} else {
			log.Warn("failed to create custom record ", cr.Host)
		}
	}

	if s.Roc != (config.Roc{}) {
		if r, err := NewRocResolver(ctx, s.Roc, s.SiteId); err == nil {
			s.Mux.HandleFunc(s.Domain, r.HandleQuery)
		}
	}

	// TODO: Add arpa. handler for reverse lookup

	s.Mux.HandleFunc(".", s.refuseQuery)

	s.Server = &dns.Server{
		Addr:              s.Bind,
		Net:               s.Protocol,
		Handler:           s.Mux,
		NotifyStartedFunc: s.Notify,
	}
	if s.Protocol == "udp" {
		s.UDPSize = 65535
	}

	return s.Server.ListenAndServe()
}

func (s *Server) refuseQuery(resp dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) > 0 {
		for _, r := range req.Question {
			log.Debug("received query ", r.String())
		}
	}

	m := new(dns.Msg)
	m.SetReply(req)
	m.RecursionAvailable = false
	m.Compress = false

	m.SetRcode(req, dns.RcodeRefused)
	if len(m.Question) > 0 {
		for _, r := range m.Question {
			log.Debugf("sending answer ;%s \t%s", r.Name, dns.RcodeToString[m.MsgHdr.Rcode])
		}
	}

	if err := resp.WriteMsg(m); err != nil {
		log.Warn("failed to answer ", err)
	}
}
