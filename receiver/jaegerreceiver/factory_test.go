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

package jaegerreceiver

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/config/configerror"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configprotocol"
	"go.opentelemetry.io/collector/config/configtls"
)

func TestTypeStr(t *testing.T) {
	factory := Factory{}

	assert.Equal(t, "jaeger", string(factory.Type()))
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestCreateReceiver(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()
	// have to enable at least one protocol for the jaeger receiver to be created
	cfg.(*Config).Protocols.GRPC = &configgrpc.GRPCServerSettings{
		NetAddr: confignet.NetAddr{
			Endpoint:  defaultGRPCBindEndpoint,
			Transport: "tcp",
		},
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	tReceiver, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.NoError(t, err, "receiver creation failed")
	assert.NotNil(t, tReceiver, "receiver creation failed")

	mReceiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, nil)
	assert.Equal(t, err, configerror.ErrDataTypeIsNotSupported)
	assert.Nil(t, mReceiver)
}

// default ports retrieved from https://www.jaegertracing.io/docs/1.16/deployment/
func TestCreateDefaultGRPCEndpoint(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.GRPC = &configgrpc.GRPCServerSettings{
		NetAddr: confignet.NetAddr{
			Endpoint:  defaultGRPCBindEndpoint,
			Transport: "tcp",
		},
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	r, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.NoError(t, err, "unexpected error creating receiver")
	assert.Equal(t, 14250, r.(*jReceiver).config.CollectorGRPCPort, "grpc port should be default")
}

func TestCreateTLSGPRCEndpoint(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.GRPC = &configgrpc.GRPCServerSettings{
		NetAddr: confignet.NetAddr{
			Endpoint:  defaultGRPCBindEndpoint,
			Transport: "tcp",
		},
		TLSSetting: &configtls.TLSServerSetting{
			TLSSetting: configtls.TLSSetting{
				CertFile: "./testdata/certificate.pem",
				KeyFile:  "./testdata/key.pem",
			},
		},
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}

	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.NoError(t, err, "tls-enabled receiver creation failed")
}

func TestCreateInvalidHTTPEndpoint(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.ThriftHTTP = &confighttp.HTTPServerSettings{
		Endpoint: defaultHTTPBindEndpoint,
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	r, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.NoError(t, err, "unexpected error creating receiver")
	assert.Equal(t, 14268, r.(*jReceiver).config.CollectorHTTPPort, "http port should be default")
}

func TestCreateInvalidThriftBinaryEndpoint(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.ThriftBinary = &configprotocol.ProtocolServerSettings{
		Endpoint: defaultThriftBinaryBindEndpoint,
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	r, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.NoError(t, err, "unexpected error creating receiver")
	assert.Equal(t, 6832, r.(*jReceiver).config.AgentBinaryThriftPort, "thrift port should be default")
}

func TestCreateInvalidThriftCompactEndpoint(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.ThriftCompact = &configprotocol.ProtocolServerSettings{
		Endpoint: defaultThriftCompactBindEndpoint,
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	r, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.NoError(t, err, "unexpected error creating receiver")
	assert.Equal(t, 6831, r.(*jReceiver).config.AgentCompactThriftPort, "thrift port should be default")
}

func TestDefaultAgentRemoteSamplingEndpointAndPort(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()
	rCfg := cfg.(*Config)

	rCfg.Protocols.ThriftCompact = &configprotocol.ProtocolServerSettings{
		Endpoint: defaultThriftCompactBindEndpoint,
	}
	rCfg.RemoteSampling = &RemoteSamplingConfig{}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	r, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.NoError(t, err, "create trace receiver should not error")
	assert.Equal(t, defaultGRPCBindEndpoint, r.(*jReceiver).config.RemoteSamplingClientSettings.Endpoint)
	assert.Equal(t, defaultAgentRemoteSamplingHTTPPort, r.(*jReceiver).config.AgentHTTPPort, "agent http port should be default")
}

func TestAgentRemoteSamplingEndpoint(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()
	rCfg := cfg.(*Config)

	endpoint := "localhost:1234"
	rCfg.Protocols.ThriftCompact = &configprotocol.ProtocolServerSettings{
		Endpoint: defaultThriftCompactBindEndpoint,
	}
	rCfg.RemoteSampling = &RemoteSamplingConfig{
		GRPCClientSettings: configgrpc.GRPCClientSettings{
			Endpoint: endpoint,
		},
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	r, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.NoError(t, err, "create trace receiver should not error")
	assert.Equal(t, endpoint, r.(*jReceiver).config.RemoteSamplingClientSettings.Endpoint)
	assert.Equal(t, defaultAgentRemoteSamplingHTTPPort, r.(*jReceiver).config.AgentHTTPPort, "agent http port should be default")
}

func TestCreateNoPort(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.ThriftHTTP = &confighttp.HTTPServerSettings{
		Endpoint: "localhost:",
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.Error(t, err, "receiver creation with no port number must fail")
}

func TestCreateLargePort(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.ThriftHTTP = &confighttp.HTTPServerSettings{
		Endpoint: "localhost:65536",
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.Error(t, err, "receiver creation with too large port number must fail")
}

func TestCreateInvalidHost(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.GRPC = &configgrpc.GRPCServerSettings{
		NetAddr: confignet.NetAddr{
			Endpoint:  "1234",
			Transport: "tcp",
		},
	}

	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.Error(t, err, "receiver creation with bad hostname must fail")
}

func TestCreateNoProtocols(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols = Protocols{}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.Error(t, err, "receiver creation with no protocols must fail")
}

func TestThriftBinaryBadPort(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.ThriftBinary = &configprotocol.ProtocolServerSettings{
		Endpoint: "localhost:65536",
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.Error(t, err, "receiver creation with a bad thrift binary port must fail")
}

func TestThriftCompactBadPort(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()

	cfg.(*Config).Protocols.ThriftCompact = &configprotocol.ProtocolServerSettings{
		Endpoint: "localhost:65536",
	}

	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)
	assert.Error(t, err, "receiver creation with a bad thrift compact port must fail")
}

func TestRemoteSamplingConfigPropagation(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()
	rCfg := cfg.(*Config)

	hostPort := 5778
	endpoint := "localhost:1234"
	strategyFile := "strategies.json"
	rCfg.Protocols.GRPC = &configgrpc.GRPCServerSettings{
		NetAddr: confignet.NetAddr{
			Endpoint:  defaultGRPCBindEndpoint,
			Transport: "tcp",
		},
	}
	rCfg.RemoteSampling = &RemoteSamplingConfig{
		GRPCClientSettings: configgrpc.GRPCClientSettings{
			Endpoint: endpoint,
		},
		HostEndpoint: fmt.Sprintf("localhost:%d", hostPort),
		StrategyFile: strategyFile,
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	r, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.NoError(t, err, "create trace receiver should not error")
	assert.Equal(t, endpoint, r.(*jReceiver).config.RemoteSamplingClientSettings.Endpoint)
	assert.Equal(t, hostPort, r.(*jReceiver).config.AgentHTTPPort, "agent http port should be configured value")
	assert.Equal(t, strategyFile, r.(*jReceiver).config.RemoteSamplingStrategyFile)
}

func TestRemoteSamplingFileRequiresGRPC(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()
	rCfg := cfg.(*Config)

	// Remove all default protocols
	rCfg.Protocols = Protocols{}
	rCfg.Protocols.ThriftCompact = &configprotocol.ProtocolServerSettings{
		Endpoint: defaultThriftCompactBindEndpoint,
	}
	rCfg.RemoteSampling = &RemoteSamplingConfig{
		StrategyFile: "strategies.json",
	}
	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	_, err := factory.CreateTraceReceiver(context.Background(), params, cfg, nil)

	assert.Error(t, err, "create trace receiver should error")
}

func TestCustomUnmarshalErrors(t *testing.T) {
	factory := Factory{}

	f := factory.CustomUnmarshaler()
	assert.NotNil(t, f, "custom unmarshal function should not be nil")

	err := f(config.NewViper(), nil)
	assert.Error(t, err, "should not have been able to marshal to a nil config")

	err = f(config.NewViper(), &RemoteSamplingConfig{})
	assert.Error(t, err, "should not have been able to marshal to a non-jaegerreceiver config")
}
