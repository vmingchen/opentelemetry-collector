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

package dynamicconfigextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type dynamicConfigExtension struct {
	config Config
	logger *zap.Logger
	// more for actual grpc stuff
}

func newServer(config Config, logger *zap.Logger) (*dynamicConfigExtension, error) {
	de := &dynamicConfigExtension{
		config: config,
		logger: logger,
	}

	return de, nil
}

func (de *dynamicConfigExtension) Start(ctx context.Context, host component.Host) error {
	de.logger.Info("Starting dynamic config extension", zap.Any("config", de.config))
	// TODO: start server
	return nil
}

func (de *dynamicConfigExtension) Shutdown(ctx context.Context) error {
	// TODO: shutdown server
	return nil
}
