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

package config

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/config/configprotocol"
)

func TestDecodeConfig(t *testing.T) {
	factories, err := ExampleComponents()
	assert.NoError(t, err)

	// Load the config
	config, err := LoadConfigFile(t, path.Join(".", "testdata", "valid-config.yaml"), factories)
	require.NoError(t, err, "Unable to load config")

	// Verify extensions.
	assert.Equal(t, 3, len(config.Extensions))
	assert.Equal(t, "some string", config.Extensions["exampleextension/1"].(*ExampleExtensionCfg).ExtraSetting)

	// Verify service.
	assert.Equal(t, 2, len(config.Service.Extensions))
	assert.Equal(t, "exampleextension/0", config.Service.Extensions[0])
	assert.Equal(t, "exampleextension/1", config.Service.Extensions[1])

	// Verify receivers
	assert.Equal(t, 2, len(config.Receivers), "Incorrect receivers count")

	assert.Equal(t,
		&ExampleReceiver{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal: "examplereceiver",
				NameVal: "examplereceiver",
			},
			ProtocolServerSettings: configprotocol.ProtocolServerSettings{
				Endpoint: "localhost:1000",
			},
			ExtraSetting: "some string",
		},
		config.Receivers["examplereceiver"],
		"Did not load receiver config correctly")

	assert.Equal(t,
		&ExampleReceiver{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal: "examplereceiver",
				NameVal: "examplereceiver/myreceiver",
			},
			ProtocolServerSettings: configprotocol.ProtocolServerSettings{
				Endpoint: "localhost:12345",
			},
			ExtraSetting: "some string",
		},
		config.Receivers["examplereceiver/myreceiver"],
		"Did not load receiver config correctly")

	// Verify exporters
	assert.Equal(t, 2, len(config.Exporters), "Incorrect exporters count")

	assert.Equal(t,
		&ExampleExporter{
			ExporterSettings: configmodels.ExporterSettings{
				NameVal: "exampleexporter",
				TypeVal: "exampleexporter",
			},
			ExtraSetting: "some export string",
		},
		config.Exporters["exampleexporter"],
		"Did not load exporter config correctly")

	assert.Equal(t,
		&ExampleExporter{
			ExporterSettings: configmodels.ExporterSettings{
				NameVal: "exampleexporter/myexporter",
				TypeVal: "exampleexporter",
			},
			ExtraSetting: "some export string 2",
		},
		config.Exporters["exampleexporter/myexporter"],
		"Did not load exporter config correctly")

	// Verify Processors
	assert.Equal(t, 1, len(config.Processors), "Incorrect processors count")

	assert.Equal(t,
		&ExampleProcessorCfg{
			ProcessorSettings: configmodels.ProcessorSettings{
				TypeVal: "exampleprocessor",
				NameVal: "exampleprocessor",
			},
			ExtraSetting: "some export string",
		},
		config.Processors["exampleprocessor"],
		"Did not load processor config correctly")

	// Verify Pipelines
	assert.Equal(t, 1, len(config.Service.Pipelines), "Incorrect pipelines count")

	assert.Equal(t,
		&configmodels.Pipeline{
			Name:       "traces",
			InputType:  configmodels.TracesDataType,
			Receivers:  []string{"examplereceiver"},
			Processors: []string{"exampleprocessor"},
			Exporters:  []string{"exampleexporter"},
		},
		config.Service.Pipelines["traces"],
		"Did not load pipeline config correctly")
}

func TestSimpleConfig(t *testing.T) {
	var testCases = []struct {
		name string // test case name (also file name containing config yaml)
	}{
		{name: "simple-config-with-no-env"},
		{name: "simple-config-with-partial-env"},
		{name: "simple-config-with-all-env"},
	}

	const extensionExtra = "some extension string"
	const extensionExtraMapValue = "some extension map value"
	const extensionExtraListElement = "some extension list value"
	assert.NoError(t, os.Setenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA", extensionExtra))
	assert.NoError(t, os.Setenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA_MAP_EXT_VALUE_1", extensionExtraMapValue+"_1"))
	assert.NoError(t, os.Setenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA_MAP_EXT_VALUE_2", extensionExtraMapValue+"_2"))
	assert.NoError(t, os.Setenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA_LIST_VALUE_1", extensionExtraListElement+"_1"))
	assert.NoError(t, os.Setenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA_LIST_VALUE_2", extensionExtraListElement+"_2"))

	const receiverExtra = "some receiver string"
	const receiverExtraMapValue = "some receiver map value"
	const receiverExtraListElement = "some receiver list value"
	assert.NoError(t, os.Setenv("RECEIVERS_EXAMPLERECEIVER_EXTRA", receiverExtra))
	assert.NoError(t, os.Setenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE_1", receiverExtraMapValue+"_1"))
	assert.NoError(t, os.Setenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE_2", receiverExtraMapValue+"_2"))
	assert.NoError(t, os.Setenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_LIST_VALUE_1", receiverExtraListElement+"_1"))
	assert.NoError(t, os.Setenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_LIST_VALUE_2", receiverExtraListElement+"_2"))

	const processorExtra = "some processor string"
	const processorExtraMapValue = "some processor map value"
	const processorExtraListElement = "some processor list value"
	assert.NoError(t, os.Setenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA", processorExtra))
	assert.NoError(t, os.Setenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA_MAP_PROC_VALUE_1", processorExtraMapValue+"_1"))
	assert.NoError(t, os.Setenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA_MAP_PROC_VALUE_2", processorExtraMapValue+"_2"))
	assert.NoError(t, os.Setenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA_LIST_VALUE_1", processorExtraListElement+"_1"))
	assert.NoError(t, os.Setenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA_LIST_VALUE_2", processorExtraListElement+"_2"))

	const exporterExtra = "some exporter string"
	const exporterExtraMapValue = "some exporter map value"
	const exporterExtraListElement = "some exporter list value"
	assert.NoError(t, os.Setenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_INT", "65"))
	assert.NoError(t, os.Setenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA", exporterExtra))
	assert.NoError(t, os.Setenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_MAP_EXP_VALUE_1", exporterExtraMapValue+"_1"))
	assert.NoError(t, os.Setenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_MAP_EXP_VALUE_2", exporterExtraMapValue+"_2"))
	assert.NoError(t, os.Setenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_LIST_VALUE_1", exporterExtraListElement+"_1"))
	assert.NoError(t, os.Setenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_LIST_VALUE_2", exporterExtraListElement+"_2"))

	defer func() {
		assert.NoError(t, os.Unsetenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA"))
		assert.NoError(t, os.Unsetenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA_MAP_EXT_VALUE"))
		assert.NoError(t, os.Unsetenv("EXTENSIONS_EXAMPLEEXTENSION_EXTRA_LIST_VALUE_1"))

		assert.NoError(t, os.Unsetenv("RECEIVERS_EXAMPLERECEIVER_EXTRA"))
		assert.NoError(t, os.Unsetenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE"))
		assert.NoError(t, os.Unsetenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_LIST_VALUE_1"))

		assert.NoError(t, os.Unsetenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA"))
		assert.NoError(t, os.Unsetenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA_MAP_PROC_VALUE"))
		assert.NoError(t, os.Unsetenv("PROCESSORS_EXAMPLEPROCESSOR_EXTRA_LIST_VALUE_1"))

		assert.NoError(t, os.Unsetenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_INT"))
		assert.NoError(t, os.Unsetenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA"))
		assert.NoError(t, os.Unsetenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_MAP_EXP_VALUE"))
		assert.NoError(t, os.Unsetenv("EXPORTERS_EXAMPLEEXPORTER_EXTRA_LIST_VALUE_1"))
	}()

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			factories, err := ExampleComponents()
			assert.NoError(t, err)

			// Load the config
			config, err := LoadConfigFile(t, path.Join(".", "testdata", test.name+".yaml"), factories)
			require.NoError(t, err, "Unable to load config")

			// Verify extensions.
			assert.Equalf(t, 1, len(config.Extensions), "TEST[%s]", test.name)
			assert.Equalf(t,
				&ExampleExtensionCfg{
					ExtensionSettings: configmodels.ExtensionSettings{
						TypeVal: "exampleextension",
						NameVal: "exampleextension",
					},
					ExtraSetting:     extensionExtra,
					ExtraMapSetting:  map[string]string{"ext-1": extensionExtraMapValue + "_1", "ext-2": extensionExtraMapValue + "_2"},
					ExtraListSetting: []string{extensionExtraListElement + "_1", extensionExtraListElement + "_2"},
				},
				config.Extensions["exampleextension"],
				"TEST[%s] Did not load extension config correctly", test.name)

			// Verify service.
			assert.Equalf(t, 1, len(config.Service.Extensions), "TEST[%s]", test.name)
			assert.Equalf(t, "exampleextension", config.Service.Extensions[0], "TEST[%s]", test.name)

			// Verify receivers
			assert.Equalf(t, 1, len(config.Receivers), "TEST[%s]", test.name)

			assert.Equalf(t,
				&ExampleReceiver{
					ReceiverSettings: configmodels.ReceiverSettings{
						TypeVal: "examplereceiver",
						NameVal: "examplereceiver",
					},
					ProtocolServerSettings: configprotocol.ProtocolServerSettings{
						Endpoint: "localhost:1234",
					},
					ExtraSetting:     receiverExtra,
					ExtraMapSetting:  map[string]string{"recv.1": receiverExtraMapValue + "_1", "recv.2": receiverExtraMapValue + "_2"},
					ExtraListSetting: []string{receiverExtraListElement + "_1", receiverExtraListElement + "_2"},
				},
				config.Receivers["examplereceiver"],
				"TEST[%s] Did not load receiver config correctly", test.name)

			// Verify exporters
			assert.Equalf(t, 1, len(config.Exporters), "TEST[%s]", test.name)

			assert.Equalf(t,
				&ExampleExporter{
					ExporterSettings: configmodels.ExporterSettings{
						NameVal: "exampleexporter",
						TypeVal: "exampleexporter",
					},
					ExtraInt:         65,
					ExtraSetting:     exporterExtra,
					ExtraMapSetting:  map[string]string{"exp_1": exporterExtraMapValue + "_1", "exp_2": exporterExtraMapValue + "_2"},
					ExtraListSetting: []string{exporterExtraListElement + "_1", exporterExtraListElement + "_2"},
				},
				config.Exporters["exampleexporter"],
				"TEST[%s] Did not load exporter config correctly", test.name)

			// Verify Processors
			assert.Equalf(t, 1, len(config.Processors), "TEST[%s]", test.name)

			assert.Equalf(t,
				&ExampleProcessorCfg{
					ProcessorSettings: configmodels.ProcessorSettings{
						TypeVal: "exampleprocessor",
						NameVal: "exampleprocessor",
					},
					ExtraSetting:     processorExtra,
					ExtraMapSetting:  map[string]string{"proc_1": processorExtraMapValue + "_1", "proc_2": processorExtraMapValue + "_2"},
					ExtraListSetting: []string{processorExtraListElement + "_1", processorExtraListElement + "_2"},
				},
				config.Processors["exampleprocessor"],
				"TEST[%s] Did not load processor config correctly", test.name)

			// Verify Pipelines
			assert.Equalf(t, 1, len(config.Service.Pipelines), "TEST[%s]", test.name)

			assert.Equalf(t,
				&configmodels.Pipeline{
					Name:       "traces",
					InputType:  configmodels.TracesDataType,
					Receivers:  []string{"examplereceiver"},
					Processors: []string{"exampleprocessor"},
					Exporters:  []string{"exampleexporter"},
				},
				config.Service.Pipelines["traces"],
				"TEST[%s] Did not load pipeline config correctly", test.name)
		})
	}
}

func TestEscapedEnvVars(t *testing.T) {
	const receiverExtraMapValue = "some receiver map value"
	assert.NoError(t, os.Setenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE_2", receiverExtraMapValue))
	defer func() {
		assert.NoError(t, os.Unsetenv("RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE_2"))
	}()

	factories, err := ExampleComponents()
	assert.NoError(t, err)

	// Load the config
	config, err := LoadConfigFile(t, path.Join(".", "testdata", "simple-config-with-escaped-env.yaml"), factories)
	require.NoError(t, err, "Unable to load config")

	// Verify extensions.
	assert.Equal(t, 1, len(config.Extensions))
	assert.Equal(t,
		&ExampleExtensionCfg{
			ExtensionSettings: configmodels.ExtensionSettings{
				TypeVal: "exampleextension",
				NameVal: "exampleextension",
			},
			ExtraSetting:     "${EXTENSIONS_EXAMPLEEXTENSION_EXTRA}",
			ExtraMapSetting:  map[string]string{"ext-1": "${EXTENSIONS_EXAMPLEEXTENSION_EXTRA_MAP_EXT_VALUE_1}", "ext-2": "${EXTENSIONS_EXAMPLEEXTENSION_EXTRA_MAP_EXT_VALUE_2}"},
			ExtraListSetting: []string{"${EXTENSIONS_EXAMPLEEXTENSION_EXTRA_LIST_VALUE_1}", "${EXTENSIONS_EXAMPLEEXTENSION_EXTRA_LIST_VALUE_2}"},
		},
		config.Extensions["exampleextension"],
		"Did not load extension config correctly")

	// Verify service.
	assert.Equal(t, 1, len(config.Service.Extensions))
	assert.Equal(t, "exampleextension", config.Service.Extensions[0])

	// Verify receivers
	assert.Equal(t, 1, len(config.Receivers))

	assert.Equal(t,
		&ExampleReceiver{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal: "examplereceiver",
				NameVal: "examplereceiver",
			},
			ProtocolServerSettings: configprotocol.ProtocolServerSettings{
				Endpoint: "localhost:1234",
			},
			ExtraSetting: "$RECEIVERS_EXAMPLERECEIVER_EXTRA",
			ExtraMapSetting: map[string]string{
				// $$ -> escaped $
				"recv.1": "$RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE_1",
				// $$$ -> escaped $ + substituted env var
				"recv.2": "$" + receiverExtraMapValue,
				// $$$$ -> two escaped $
				"recv.3": "$$RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE_3",
				// escaped $ in the middle
				"recv.4": "some${RECEIVERS_EXAMPLERECEIVER_EXTRA_MAP_RECV_VALUE_4}text",
				// $$$$ -> two escaped $
				"recv.5": "${ONE}${TWO}",
				// trailing escaped $
				"recv.6": "text$",
				// escaped $ alone
				"recv.7": "$",
			},
			ExtraListSetting: []string{"$RECEIVERS_EXAMPLERECEIVER_EXTRA_LIST_VALUE_1", "$RECEIVERS_EXAMPLERECEIVER_EXTRA_LIST_VALUE_2"},
		},
		config.Receivers["examplereceiver"],
		"Did not load receiver config correctly")

	// Verify exporters
	assert.Equal(t, 1, len(config.Exporters))

	assert.Equal(t,
		&ExampleExporter{
			ExporterSettings: configmodels.ExporterSettings{
				NameVal: "exampleexporter",
				TypeVal: "exampleexporter",
			},
			ExtraSetting:     "${EXPORTERS_EXAMPLEEXPORTER_EXTRA}",
			ExtraMapSetting:  map[string]string{"exp_1": "${EXPORTERS_EXAMPLEEXPORTER_EXTRA_MAP_EXP_VALUE_1}", "exp_2": "${EXPORTERS_EXAMPLEEXPORTER_EXTRA_MAP_EXP_VALUE_2}"},
			ExtraListSetting: []string{"${EXPORTERS_EXAMPLEEXPORTER_EXTRA_LIST_VALUE_1}", "${EXPORTERS_EXAMPLEEXPORTER_EXTRA_LIST_VALUE_2}"},
		},
		config.Exporters["exampleexporter"],
		"Did not load exporter config correctly")

	// Verify Processors
	assert.Equal(t, 1, len(config.Processors))

	assert.Equal(t,
		&ExampleProcessorCfg{
			ProcessorSettings: configmodels.ProcessorSettings{
				TypeVal: "exampleprocessor",
				NameVal: "exampleprocessor",
			},
			ExtraSetting:     "$PROCESSORS_EXAMPLEPROCESSOR_EXTRA",
			ExtraMapSetting:  map[string]string{"proc_1": "$PROCESSORS_EXAMPLEPROCESSOR_EXTRA_MAP_PROC_VALUE_1", "proc_2": "$PROCESSORS_EXAMPLEPROCESSOR_EXTRA_MAP_PROC_VALUE_2"},
			ExtraListSetting: []string{"$PROCESSORS_EXAMPLEPROCESSOR_EXTRA_LIST_VALUE_1", "$PROCESSORS_EXAMPLEPROCESSOR_EXTRA_LIST_VALUE_2"},
		},
		config.Processors["exampleprocessor"],
		"Did not load processor config correctly")

	// Verify Pipelines
	assert.Equal(t, 1, len(config.Service.Pipelines))

	assert.Equal(t,
		&configmodels.Pipeline{
			Name:       "traces",
			InputType:  configmodels.TracesDataType,
			Receivers:  []string{"examplereceiver"},
			Processors: []string{"exampleprocessor"},
			Exporters:  []string{"exampleexporter"},
		},
		config.Service.Pipelines["traces"],
		"Did not load pipeline config correctly")
}

func TestDecodeConfig_MultiProto(t *testing.T) {
	factories, err := ExampleComponents()
	assert.NoError(t, err)

	// Load the config
	config, err := LoadConfigFile(t, path.Join(".", "testdata", "multiproto-config.yaml"), factories)
	require.NoError(t, err, "Unable to load config")

	// Verify receivers
	assert.Equal(t, 2, len(config.Receivers), "Incorrect receivers count")

	assert.Equal(t,
		&MultiProtoReceiver{
			TypeVal: "multireceiver",
			NameVal: "multireceiver",
			Protocols: map[string]MultiProtoReceiverOneCfg{
				"http": {
					Endpoint:     "example.com:8888",
					ExtraSetting: "extra string 1",
				},
				"tcp": {
					Endpoint:     "omnition.com:9999",
					ExtraSetting: "extra string 2",
				},
			},
		},
		config.Receivers["multireceiver"],
		"Did not load receiver config correctly")

	assert.Equal(t,
		&MultiProtoReceiver{
			TypeVal: "multireceiver",
			NameVal: "multireceiver/myreceiver",
			Protocols: map[string]MultiProtoReceiverOneCfg{
				"http": {
					Endpoint:     "localhost:12345",
					ExtraSetting: "some string 1",
				},
				"tcp": {
					Endpoint:     "0.0.0.0:4567",
					ExtraSetting: "some string 2",
				},
			},
		},
		config.Receivers["multireceiver/myreceiver"],
		"Did not load receiver config correctly")
}

func TestDecodeConfig_Invalid(t *testing.T) {

	var testCases = []struct {
		name     string          // test case name (also file name containing config yaml)
		expected configErrorCode // expected error (if nil any error is acceptable)
	}{
		{name: "empty-config"},
		{name: "missing-all-sections"},
		{name: "missing-exporters", expected: errMissingExporters},
		{name: "missing-receivers", expected: errMissingReceivers},
		{name: "missing-processors"},
		{name: "invalid-extension-name", expected: errExtensionNotExists},
		{name: "invalid-receiver-name"},
		{name: "invalid-receiver-reference", expected: errPipelineReceiverNotExists},
		{name: "missing-extension-type", expected: errInvalidTypeAndNameKey},
		{name: "missing-receiver-type", expected: errInvalidTypeAndNameKey},
		{name: "missing-exporter-name-after-slash", expected: errInvalidTypeAndNameKey},
		{name: "missing-processor-type", expected: errInvalidTypeAndNameKey},
		{name: "missing-pipelines", expected: errMissingPipelines},
		{name: "pipeline-must-have-exporter", expected: errPipelineMustHaveExporter},
		{name: "pipeline-must-have-exporter2", expected: errPipelineMustHaveExporter},
		{name: "pipeline-must-have-receiver", expected: errPipelineMustHaveReceiver},
		{name: "pipeline-must-have-receiver2", expected: errPipelineMustHaveReceiver},
		{name: "pipeline-exporter-not-exists", expected: errPipelineExporterNotExists},
		{name: "pipeline-processor-not-exists", expected: errPipelineProcessorNotExists},
		{name: "unknown-extension-type", expected: errUnknownExtensionType},
		{name: "unknown-receiver-type", expected: errUnknownReceiverType},
		{name: "unknown-exporter-type", expected: errUnknownExporterType},
		{name: "unknown-processor-type", expected: errUnknownProcessorType},
		{name: "invalid-service-extensions-value", expected: errUnmarshalErrorOnService},
		{name: "invalid-sequence-value", expected: errUnmarshalErrorOnPipeline},
		{name: "invalid-pipeline-type", expected: errInvalidPipelineType},
		{name: "invalid-pipeline-type-and-name", expected: errInvalidTypeAndNameKey},
		{name: "duplicate-extension", expected: errDuplicateExtensionName},
		{name: "duplicate-receiver", expected: errDuplicateReceiverName},
		{name: "duplicate-exporter", expected: errDuplicateExporterName},
		{name: "duplicate-processor", expected: errDuplicateProcessorName},
		{name: "duplicate-pipeline", expected: errDuplicatePipelineName},
		{name: "invalid-top-level-section", expected: errUnmarshalErrorOnTopLevelSection},
		{name: "invalid-extension-section", expected: errUnmarshalErrorOnExtension},
		{name: "invalid-service-section", expected: errUnmarshalErrorOnService},
		{name: "invalid-receiver-section", expected: errUnmarshalErrorOnReceiver},
		{name: "invalid-processor-section", expected: errUnmarshalErrorOnProcessor},
		{name: "invalid-exporter-section", expected: errUnmarshalErrorOnExporter},
		{name: "invalid-pipeline-section", expected: errUnmarshalErrorOnPipeline},
	}

	factories, err := ExampleComponents()
	assert.NoError(t, err)

	for _, test := range testCases {
		_, err := LoadConfigFile(t, path.Join(".", "testdata", test.name+".yaml"), factories)
		if err == nil {
			t.Errorf("expected error but succeeded on invalid config case: %s", test.name)
		} else if test.expected != 0 {
			cfgErr, ok := err.(*configError)
			if !ok {
				t.Errorf("expected config error code %v but got a different error '%v' on invalid config case: %s",
					test.expected, err, test.name)
			} else {
				if cfgErr.code != test.expected {
					t.Errorf("expected config error code %v but got error code %v (msg: %q) on invalid config case: %s",
						test.expected, cfgErr.code, cfgErr.msg, test.name)
				}

				if cfgErr.Error() == "" {
					t.Errorf("returned config error %v with empty error message", cfgErr.code)
				}
			}
		}
	}
}

func TestLoadEmptyConfig(t *testing.T) {
	factories, err := ExampleComponents()
	assert.NoError(t, err)

	// Open the file for reading.
	file, err := os.Open(path.Join(".", "testdata", "empty-config.yaml"))
	require.NoError(t, err)

	defer func() {
		require.NoError(t, file.Close())
	}()

	// Read yaml config from file
	v := NewViper()
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(file))

	// Load the config from viper using the given factories.
	_, err = Load(v, factories)
	assert.NoError(t, err)
}
