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

package hostmetricsreceiver

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/config/configerror"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal"
)

var creationParams = component.ReceiverCreateParams{Logger: zap.NewNop()}

func TestCreateDefaultConfig(t *testing.T) {
	factory := &Factory{}
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestCreateReceiver(t *testing.T) {
	factory := &Factory{}
	cfg := factory.CreateDefaultConfig()

	tReceiver, err := factory.CreateTraceReceiver(context.Background(), creationParams, cfg, nil)

	assert.Equal(t, err, configerror.ErrDataTypeIsNotSupported)
	assert.Nil(t, tReceiver)

	mReceiver, err := factory.CreateMetricsReceiver(context.Background(), creationParams, cfg, nil)

	assert.NoError(t, err)
	assert.NotNil(t, mReceiver)
}

func TestCreateReceiver_ScraperKeyConfigError(t *testing.T) {
	const errorKey string = "error"

	factory := &Factory{}
	cfg := &Config{Scrapers: map[string]internal.Config{errorKey: &mockConfig{}}}

	_, err := factory.CreateMetricsReceiver(context.Background(), creationParams, cfg, nil)
	assert.EqualError(t, err, fmt.Sprintf("host metrics scraper factory not found for key: %q", errorKey))
}
