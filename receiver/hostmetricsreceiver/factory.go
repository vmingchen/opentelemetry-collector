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
	"time"

	"github.com/spf13/viper"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configerror"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/diskscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/filesystemscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/loadscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/memoryscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/networkscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/processesscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/processscraper"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/swapscraper"
)

// This file implements Factory for HostMetrics receiver.

const (
	// The value of "type" key in configuration.
	typeStr     = "hostmetrics"
	scrapersKey = "scrapers"
)

// Factory is the Factory for receiver.
type Factory struct {
	scraperFactories         map[string]internal.ScraperFactory
	resourceScraperFactories map[string]internal.ResourceScraperFactory
}

// NewFactory creates a new factory for host metrics receiver.
func NewFactory() *Factory {
	return &Factory{
		scraperFactories: map[string]internal.ScraperFactory{
			cpuscraper.TypeStr:        &cpuscraper.Factory{},
			diskscraper.TypeStr:       &diskscraper.Factory{},
			loadscraper.TypeStr:       &loadscraper.Factory{},
			filesystemscraper.TypeStr: &filesystemscraper.Factory{},
			memoryscraper.TypeStr:     &memoryscraper.Factory{},
			networkscraper.TypeStr:    &networkscraper.Factory{},
			processesscraper.TypeStr:  &processesscraper.Factory{},
			swapscraper.TypeStr:       &swapscraper.Factory{},
		},
		resourceScraperFactories: map[string]internal.ResourceScraperFactory{
			processscraper.TypeStr: &processscraper.Factory{},
		},
	}
}

// Type returns the type of the Receiver config created by this Factory.
func (f *Factory) Type() configmodels.Type {
	return typeStr
}

// CustomUnmarshaler returns custom unmarshaler for this config.
func (f *Factory) CustomUnmarshaler() component.CustomUnmarshaler {
	return func(componentViperSection *viper.Viper, intoCfg interface{}) error {

		// load the non-dynamic config normally

		err := componentViperSection.Unmarshal(intoCfg)
		if err != nil {
			return err
		}

		cfg, ok := intoCfg.(*Config)
		if !ok {
			return fmt.Errorf("config type not hostmetrics.Config")
		}

		if cfg.CollectionInterval <= 0 {
			return fmt.Errorf("collection_interval must be a positive duration")
		}

		// dynamically load the individual collector configs based on the key name

		cfg.Scrapers = map[string]internal.Config{}

		scrapersViperSection := config.ViperSub(componentViperSection, scrapersKey)
		if scrapersViperSection == nil || len(scrapersViperSection.AllKeys()) == 0 {
			return fmt.Errorf("must specify at least one scraper when using hostmetrics receiver")
		}

		for key := range componentViperSection.GetStringMap(scrapersKey) {
			factory, ok := f.getScraperFactory(key)
			if !ok {
				return fmt.Errorf("invalid scraper key: %s", key)
			}

			collectorCfg := factory.CreateDefaultConfig()
			collectorViperSection := config.ViperSub(scrapersViperSection, key)
			err := collectorViperSection.UnmarshalExact(collectorCfg)
			if err != nil {
				return fmt.Errorf("error reading settings for scraper type %q: %v", key, err)
			}

			cfg.Scrapers[key] = collectorCfg
		}

		return nil
	}
}

func (f *Factory) getScraperFactory(key string) (internal.BaseFactory, bool) {
	if factory, ok := f.scraperFactories[key]; ok {
		return factory, true
	}

	if factory, ok := f.resourceScraperFactories[key]; ok {
		return factory, true
	}

	return nil, false
}

// CreateDefaultConfig creates the default configuration for receiver.
func (f *Factory) CreateDefaultConfig() configmodels.Receiver {
	return &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr,
		},
		CollectionInterval: time.Minute,
	}
}

// CreateTraceReceiver returns error as trace receiver is not applicable to host metrics receiver.
func (f *Factory) CreateTraceReceiver(
	ctx context.Context,
	params component.ReceiverCreateParams,
	cfg configmodels.Receiver,
	consumer consumer.TraceConsumer,
) (component.TraceReceiver, error) {
	// Host Metrics does not support traces
	return nil, configerror.ErrDataTypeIsNotSupported
}

// CreateMetricsReceiver creates a metrics receiver based on provided config.
func (f *Factory) CreateMetricsReceiver(
	ctx context.Context,
	params component.ReceiverCreateParams,
	cfg configmodels.Receiver,
	consumer consumer.MetricsConsumer,
) (component.MetricsReceiver, error) {
	config := cfg.(*Config)

	hmr, err := newHostMetricsReceiver(ctx, params.Logger, config, f.scraperFactories, f.resourceScraperFactories, consumer)
	if err != nil {
		return nil, err
	}

	return hmr, nil
}
