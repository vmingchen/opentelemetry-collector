// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otlpreceiver

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/config/configtls"
)

func TestLoadConfig(t *testing.T) {
	factories, err := config.ExampleComponents()
	assert.NoError(t, err)

	factory := &Factory{}
	factories.Receivers[typeStr] = factory
	cfg, err := config.LoadConfigFile(t, path.Join(".", "testdata", "config.yaml"), factories)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 7)

	r0 := cfg.Receivers["otlp"]
	assert.Equal(t, r0, factory.CreateDefaultConfig())

	r1 := cfg.Receivers["otlp/customname"].(*Config)
	assert.Equal(t, r1,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:  typeStr,
				NameVal:  "otlp/customname",
				Endpoint: "localhost:9090",
			},
			Transport: "tcp",
		})

	r2 := cfg.Receivers["otlp/keepalive"].(*Config)
	assert.Equal(t, r2,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:  typeStr,
				NameVal:  "otlp/keepalive",
				Endpoint: "0.0.0.0:55680",
			},
			TLSCredentials: nil,
			Transport:      "tcp",
			Keepalive: &serverParametersAndEnforcementPolicy{
				ServerParameters: &keepaliveServerParameters{
					MaxConnectionIdle:     11 * time.Second,
					MaxConnectionAge:      12 * time.Second,
					MaxConnectionAgeGrace: 13 * time.Second,
					Time:                  30 * time.Second,
					Timeout:               5 * time.Second,
				},
				EnforcementPolicy: &keepaliveEnforcementPolicy{
					MinTime:             10 * time.Second,
					PermitWithoutStream: true,
				},
			},
		})

	r3 := cfg.Receivers["otlp/msg-size-conc-connect-max-idle"].(*Config)
	assert.Equal(t, r3,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:  typeStr,
				NameVal:  "otlp/msg-size-conc-connect-max-idle",
				Endpoint: "0.0.0.0:55680",
			},
			Transport:            "tcp",
			MaxRecvMsgSizeMiB:    32,
			MaxConcurrentStreams: 16,
			Keepalive: &serverParametersAndEnforcementPolicy{
				ServerParameters: &keepaliveServerParameters{
					MaxConnectionIdle: 10 * time.Second,
				},
			},
		})

	// NOTE: Once the config loader checks for the files existence, this test may fail and require
	// 	use of fake cert/key for test purposes.
	r4 := cfg.Receivers["otlp/tlscredentials"].(*Config)
	assert.Equal(t, r4,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:  typeStr,
				NameVal:  "otlp/tlscredentials",
				Endpoint: "0.0.0.0:55680",
			},
			TLSCredentials: &configtls.TLSSetting{
				CertFile: "test.crt",
				KeyFile:  "test.key",
			},
			Transport: "tcp",
		})

	r5 := cfg.Receivers["otlp/cors"].(*Config)
	assert.Equal(t, r5,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:  typeStr,
				NameVal:  "otlp/cors",
				Endpoint: "0.0.0.0:55680",
			},
			Transport:   "tcp",
			CorsOrigins: []string{"https://*.test.com", "https://test.com"},
		})

	r6 := cfg.Receivers["otlp/uds"].(*Config)
	assert.Equal(t, r6,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:  typeStr,
				NameVal:  "otlp/uds",
				Endpoint: "/tmp/otlp.sock",
			},
			Transport: "unix",
		})
}

func TestBuildOptions_TLSCredentials(t *testing.T) {
	cfg := Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			NameVal: "IncorrectTLS",
		},
		TLSCredentials: &configtls.TLSSetting{
			CertFile: "willfail",
		},
	}
	_, err := cfg.buildOptions()
	assert.EqualError(t, err, `error initializing OTLP receiver "IncorrectTLS" TLS Credentials: failed to load TLS config: for auth via TLS, either both certificate and key must be supplied, or neither`)

	cfg.TLSCredentials = &configtls.TLSSetting{}
	opt, err := cfg.buildOptions()
	assert.NoError(t, err)
	assert.NotNil(t, opt)
}
