// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package metricsproxy // nolint:dupl

import (
	"math/rand"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/tiflow/engine/pkg/promutil"

	"github.com/prometheus/client_golang/prometheus"
)

func (t *testMetricsProxySuite) TestHistogramVecProxy(c *C) {
	rand.Seed(time.Now().UnixNano())
	for _, oneCase := range testCases {
		histogram := NewHistogramVec(&promutil.PromFactory{},
			prometheus.HistogramOpts{
				Namespace:   "dm",
				Subsystem:   "metricsProxy",
				Name:        "Test_Histogram",
				Help:        "dm histogram metrics proxy test",
				ConstLabels: nil,
			}, oneCase.LabelsNames)
		for _, aArgs := range oneCase.AddArgs {
			if rand.Intn(199)%2 == 0 {
				histogram.WithLabelValues(aArgs...).Observe(float64(rand.Intn(199)))
			} else {
				labels := make(prometheus.Labels)
				for k, labelName := range oneCase.LabelsNames {
					labels[labelName] = aArgs[k]
				}
				histogram.With(labels)
			}
		}
		for _, dArgs := range oneCase.DeleteArgs {
			histogram.DeleteAllAboutLabels(dArgs)
		}

		cOutput := make(chan prometheus.Metric, len(oneCase.AddArgs)*3)

		histogram.Collect(cOutput)

		c.Assert(len(cOutput), Equals, oneCase.WantResLength)
	}
}
