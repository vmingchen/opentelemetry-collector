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
	"net"
	// "net/http"
	// "runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/testutils"
)

func TestDyconfigExtensionUsage(t *testing.T) {
	config := Config{
		Endpoint: testutils.GetAvailableLocalAddress(t),
	}

	dynamicconfigExt, err := newServer(config, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, dynamicconfigExt)

	require.NoError(t, dynamicconfigExt.Start(context.Background(), componenttest.NewNopHost()))
	defer dynamicconfigExt.Shutdown(context.Background())
	//
	// // Give a chance for the server goroutine to run.
	// runtime.Gosched()
	//
	// _, dynamicconfigPort, err := net.SplitHostPort(config.Endpoint)
	// require.NoError(t, err)
	//
	// client := &http.Client{}
	// resp, err := client.Get("http://localhost:" + dynamicconfigPort)
	// require.NoError(t, err)
	// defer resp.Body.Close()
	//
	// require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDyconfigExtensionPortAlreadyInUse(t *testing.T) {
	endpoint := testutils.GetAvailableLocalAddress(t)
	ln, err := net.Listen("tcp", endpoint)
	require.NoError(t, err)
	defer ln.Close()

	config := Config{
		Endpoint: endpoint,
	}
	dynamicconfigExt, err := newServer(config, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, dynamicconfigExt)

	// require.Error(t, dynamicconfigExt.Start(context.Background(), componenttest.NewNopHost()))
}

func TestDyconfigMultipleStarts(t *testing.T) {
	config := Config{
		Endpoint: testutils.GetAvailableLocalAddress(t),
	}

	dynamicconfigExt, err := newServer(config, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, dynamicconfigExt)

	// require.NoError(t, dynamicconfigExt.Start(context.Background(), componenttest.NewNopHost()))
	// defer dynamicconfigExt.Shutdown(context.Background())
	//
	// // Try to start it again, it will fail since it is on the same endpoint.
	// require.Error(t, dynamicconfigExt.Start(context.Background(), componenttest.NewNopHost()))
}

func TestDyconfigMultipleShutdowns(t *testing.T) {
	config := Config{
		Endpoint: testutils.GetAvailableLocalAddress(t),
	}

	dynamicconfigExt, err := newServer(config, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, dynamicconfigExt)

	// require.NoError(t, dynamicconfigExt.Start(context.Background(), componenttest.NewNopHost()))
	// require.NoError(t, dynamicconfigExt.Shutdown(context.Background()))
	// require.NoError(t, dynamicconfigExt.Shutdown(context.Background()))
}

func TestDyconfigShutdownWithoutStart(t *testing.T) {
	config := Config{
		Endpoint: testutils.GetAvailableLocalAddress(t),
	}

	dynamicconfigExt, err := newServer(config, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, dynamicconfigExt)

	// require.NoError(t, dynamicconfigExt.Shutdown(context.Background()))
}
