// Copyright 2020 Google LLC
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

package main

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter"
	upstreamgooglecloudexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/couchdbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/memcachedreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension/ballastextension"
	"go.opentelemetry.io/collector/extension/zpagesextension"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service/featuregate"
	"go.uber.org/multierr"

	"github.com/GoogleCloudPlatform/opentelemetry-operations-collector/internal/exporter/googlecloudexporter"
	"github.com/GoogleCloudPlatform/opentelemetry-operations-collector/processor/agentmetricsprocessor"
	"github.com/GoogleCloudPlatform/opentelemetry-operations-collector/processor/casttosumprocessor"
	"github.com/GoogleCloudPlatform/opentelemetry-operations-collector/processor/normalizesumsprocessor"
)

const pdataExporterFeatureGate = "exporter.googlecloud.OTLPDirect"

func init() {
	featuregate.Register(featuregate.Gate{
		ID:          pdataExporterFeatureGate,
		Description: "When enabled, the googlecloud exporter translates pdata directly to google cloud monitoring's types, rather than first translating to opencensus.",
		Enabled:     false,
	})
}

func components() (component.Factories, error) {
	errs := []error{}
	factories, err := Components()
	if err != nil {
		return component.Factories{}, err
	}

	extensions := []component.ExtensionFactory{}
	for _, ext := range factories.Extensions {
		extensions = append(extensions, ext)
	}
	factories.Extensions, err = component.MakeExtensionFactoryMap(extensions...)
	if err != nil {
		errs = append(errs, err)
	}

	receivers := []component.ReceiverFactory{
		hostmetricsreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		prometheusexecreceiver.NewFactory(),
		nginxreceiver.NewFactory(),
		jmxreceiver.NewFactory(),
		windowsperfcountersreceiver.NewFactory(),
		redisreceiver.NewFactory(),
		mysqlreceiver.NewFactory(),
		apachereceiver.NewFactory(),
		memcachedreceiver.NewFactory(),
		mongodbreceiver.NewFactory(),
		postgresqlreceiver.NewFactory(),
		elasticsearchreceiver.NewFactory(),
		couchdbreceiver.NewFactory(),
		zookeeperreceiver.NewFactory(),
		rabbitmqreceiver.NewFactory(),
	}
	for _, rcv := range factories.Receivers {
		receivers = append(receivers, rcv)
	}
	factories.Receivers, err = component.MakeReceiverFactoryMap(receivers...)
	if err != nil {
		errs = append(errs, err)
	}

	exporters := []component.ExporterFactory{
		fileexporter.NewFactory(),
	}
	if featuregate.IsEnabled(pdataExporterFeatureGate) {
		exporters = append(exporters, googlecloudexporter.NewFactory())
	} else {
		exporters = append(exporters, upstreamgooglecloudexporter.NewFactory())
	}
	for _, exp := range factories.Exporters {
		exporters = append(exporters, exp)
	}
	factories.Exporters, err = component.MakeExporterFactoryMap(exporters...)
	if err != nil {
		errs = append(errs, err)
	}

	processors := []component.ProcessorFactory{
		agentmetricsprocessor.NewFactory(),
		casttosumprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		normalizesumsprocessor.NewFactory(),
		metricstransformprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
	}
	for _, pr := range factories.Processors {
		processors = append(processors, pr)
	}
	factories.Processors, err = component.MakeProcessorFactoryMap(processors...)
	if err != nil {
		errs = append(errs, err)
	}

	return factories, multierr.Combine(errs...)
}

func Components() (
	component.Factories,
	error,
) {
	var errs error

	extensions, err := component.MakeExtensionFactoryMap(
		zpagesextension.NewFactory(),
		ballastextension.NewFactory(),
	)
	errs = multierr.Append(errs, err)

	receivers, err := component.MakeReceiverFactoryMap(
		otlpreceiver.NewFactory(),
	)
	errs = multierr.Append(errs, err)

	exporters, err := component.MakeExporterFactoryMap(
		loggingexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	)
	errs = multierr.Append(errs, err)

	processors, err := component.MakeProcessorFactoryMap(
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
	)
	errs = multierr.Append(errs, err)

	factories := component.Factories{
		Extensions: extensions,
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
	}

	return factories, errs
}
