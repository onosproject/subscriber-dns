// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	paths   = []string{"testdata/config.json", "testdata/openidc.json"}
	testCfg = Config{
		Bind:     "0.0.0.0:53",
		Protocol: "udp",
		CustomRecords: []CustomRecord{
			{
				Host: "test-4g-pi1.device.test.aether.net",
				Addr: "10.250.0.254",
			},
		},
		Domain: "device.test.aether.net",
		Roc: Roc{
			Url: "https://roc.test.aether.org",
			OpenIDC: OpenIDC{
				TokenUrl: "https://keycloak.test.aether.org/auth/realms/master/protocol/openid-connect/token",
				Username: "testuser",
				Password: "testpassword",
			},
		},
		Site:     "Test Site",
		LogLevel: "debug",
	}
)

func TestConfig_LoadConfig(t *testing.T) {
	assert := require.New(t)

	cfg, err := LoadConfig(paths)

	assert.Nil(err)
	assert.Equal(cfg, &testCfg)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "all required values should validate",
			config: Config{
				Bind:     "0.0.0.0:53",
				Protocol: "udp",
				Domain:   "device.test.aether.net",
				Roc: Roc{
					Url: "roc.test.aether.net",
				},
			},
			wantErr: false,
		},
		{
			name: "missing bind should not validate",
			config: Config{
				Protocol: "udp",
				Domain:   "device.test.aether.net",
				Roc: Roc{
					Url: "roc.test.aether.net",
				},
			},
			wantErr: true,
		},
		{
			name: "missing protocol should not validate",
			config: Config{
				Bind:   "0.0.0.0:53",
				Domain: "device.test.aether.net",
				Roc: Roc{
					Url: "roc.test.aether.net",
				},
			},
			wantErr: true,
		},
		{
			name: "missing domain should not validate",
			config: Config{
				Bind:     "0.0.0.0:53",
				Protocol: "udp",
				Roc: Roc{
					Url: "roc.test.aether.net",
				},
			},
			wantErr: true,
		},
		{
			name: "missing ROC URL should not validate",
			config: Config{
				Bind:     "0.0.0.0:53",
				Protocol: "udp",
				Domain:   "device.test.aether.net",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.config.validate(); (err != nil) != test.wantErr {
				t.Errorf("validate error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
