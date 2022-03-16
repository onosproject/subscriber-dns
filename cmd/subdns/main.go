// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"flag"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/subscriber-dns/pkg/config"
	"github.com/onosproject/subscriber-dns/pkg/server"
)

var (
	log         = logging.GetLogger("subscriber-dns")
	configFiles arrayFlags
)

type arrayFlags []string

func (f *arrayFlags) String() string {
	return ""
}

func (f *arrayFlags) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func init() {
	flag.Var(&configFiles, "config", "config file location, can be used multiple times")
}

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(configFiles)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: fix log level configurable
	log.SetLevel(logging.InfoLevel)
	log.Infof("%+v", *cfg)

	server := &server.Server{
		Bind:     cfg.Bind,
		Protocol: cfg.Protocol,
		Domain:   cfg.Domain,
		Notify: func() {
			log.Infof("listening on %s %s", cfg.Protocol, cfg.Bind)
		},
		CustomRecords: cfg.CustomRecords,
		Roc:           cfg.Roc,
		SiteId:        cfg.Site,
	}

	if err := server.ListenAndServe(context.Background()); err != nil {
		log.Fatal("failed to start ", err)
	}
}
