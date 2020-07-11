// Copyright 2020, OpenTelemetry Authors
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

package filterprocessor

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/config/configmodels"
)

func TestType(t *testing.T) {
	factory := Factory{}
	pType := factory.Type()

	assert.Equal(t, pType, configmodels.Type("filter"))
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := Factory{}
	cfg := factory.CreateDefaultConfig()
	assert.Equal(t, cfg, &Config{
		ProcessorSettings: configmodels.ProcessorSettings{
			NameVal: typeStr,
			TypeVal: typeStr,
		},
	})
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestCreateProcessors(t *testing.T) {
	tests := []struct {
		configName string
		succeed    bool
	}{
		{
			configName: "config_regexp.yaml",
			succeed:    true,
		}, {
			configName: "config_strict.yaml",
			succeed:    true,
		}, {
			configName: "config_invalid.yaml",
			succeed:    false,
		},
	}

	for _, test := range tests {
		factories, err := config.ExampleComponents()
		assert.Nil(t, err)

		factory := &Factory{}
		factories.Processors[typeStr] = factory
		cfg, err := config.LoadConfigFile(t, path.Join(".", "testdata", test.configName), factories)
		assert.Nil(t, err)

		for name, cfg := range cfg.Processors {
			t.Run(fmt.Sprintf("%s/%s", test.configName, name), func(t *testing.T) {
				factory := &Factory{}

				tp, tErr := factory.CreateTraceProcessor(
					context.Background(),
					component.ProcessorCreateParams{Logger: zap.NewNop()},
					nil,
					cfg)
				// Not implemented error
				assert.NotNil(t, tErr)
				assert.Nil(t, tp)

				mp, mErr := factory.CreateMetricsProcessor(
					context.Background(),
					component.ProcessorCreateParams{Logger: zap.NewNop()},
					nil,
					cfg)
				assert.Equal(t, test.succeed, mp != (*filterMetricProcessor)(nil))
				assert.Equal(t, test.succeed, mErr == nil)
			})
		}
	}
}
