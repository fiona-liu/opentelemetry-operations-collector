// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package agentmetricsprocessor

import "go.opentelemetry.io/collector/model/pdata"

func removeVersionAttribute(rms pdata.ResourceMetricsSlice) {
	for i := 0; i < rms.Len(); i++ {
		ilms := rms.At(i).InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			metrics := ilms.At(j).Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)

				var dps pdata.NumberDataPointSlice
				switch metric.DataType() {
				case pdata.MetricDataTypeGauge:
					dps = metric.Gauge().DataPoints()
				case pdata.MetricDataTypeSum:
					dps = metric.Sum().DataPoints()
				}

				for l := 0; l < dps.Len(); l++ {
					dp := dps.At(l)
					dp.Attributes().Delete("service_version")
				}
			}
		}
	}
}
