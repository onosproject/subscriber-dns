// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/onosproject/onos-lib-go/pkg/errors"
)

type CustomRecord struct {
	Host string `json:"host"`
	Addr string `json:"address"`
}

type OpenIDC struct {
	TokenUrl string `json:"tokenUrl"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Roc struct {
	Url     string `json:"url"`
	OpenIDC OpenIDC
}

type Config struct {
	Bind          string         `json:"bind"`
	Protocol      string         `json:"protocol"` // "tcp", "tcp-tls", or "udp"
	Domain        string         `json:"domain"`   // domain to serve, will remove when site model provides site domain
	Site          string         `json:"site"`     // site name to serve, the name should be registered in ROC
	CustomRecords []CustomRecord `json:"customRecords"`
	Roc           Roc            `json:"roc"`
	LogLevel      string         `json:"logLevel"`
}

func (cfg *Config) validate() error {
	if cfg.Bind == "" {
		return errors.NewInvalid("missing bind address and port")
	}

	if cfg.Protocol == "" {
		// TODO: validate protocols
		return errors.NewInvalid("missing protocol")
	}

	if cfg.Domain == "" {
		return errors.NewInvalid("missing domain to serve")
	}

	if cfg.Roc == (Roc{}) || cfg.Roc.Url == "" {
		return errors.NewInvalid("missing ROC URL")
	}

	if _, err := url.Parse(cfg.Roc.Url); err != nil {
		return errors.NewInvalid("invalid ROC URL")
	}

	// TODO: validate oidc config if exists

	return nil
}

func defaultConfig() *Config {
	cfg := new(Config)

	cfg.Bind = "0.0.0.0:53"
	cfg.Protocol = "udp"
	cfg.LogLevel = "info"

	return cfg
}

// LoadConfig reads configuration from paths and validates it
func LoadConfig(paths []string) (*Config, error) {
	cfg := defaultConfig()

	for _, p := range paths {
		str, err := os.Open(p)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %s", p, err)
		}

		byteValue, err := ioutil.ReadAll(str)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %s", p, err)
		}

		err = json.Unmarshal(byteValue, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %s", p, err)
		}

		str.Close()
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
