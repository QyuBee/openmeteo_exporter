package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type SatelliteRadiationCollector struct {
	Client   *OpenMeteoClient
	Location *LocationConfig
	Today    string
	mu       sync.Mutex
	hourly   map[string]map[string]float64
}

func (c SatelliteRadiationCollector) Collect(ch chan<- prometheus.Metric) {
	satelliteResp, err := c.Client.GetSatelliteRadiation(c.Location)
	if err != nil {
		level.Warn(logger).Log(
			"msg", "Failed to collect satellite radiation information",
			"location", c.Location.Name,
			"err", err,
		)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.hourly == nil {
		c.hourly = make(map[string]map[string]float64)
	}

	for _, name := range c.Location.SatelliteRadiation.HourlyRadiationVariables {
		if _, ok := c.hourly[name]; !ok {
			c.hourly[name] = make(map[string]float64)
		}

		for i, t := range satelliteResp.Hourly.Time {
			// ignore null values
			if satelliteResp.Hourly.Variables[name][i] == nil {
				continue
			}
			c.hourly[name][t] = satelliteResp.Hourly.Variables[name][i].(float64)
		}

		// expose les metrics à Prometheus
		units := "w_per_m2"
		description, _ := GetVariableDesc("satellite_radiation", name)
		desc := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "satellite_radiation", fmt.Sprintf("%s_%s", name, units)),
			description,
			[]string{"location"},
			nil,
		)

		for t, v := range c.hourly[name] {
			m := prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				v,
				c.Location.Name,
			)

			ts, _ := time.Parse(
				"2006-01-02T15:04",
				t,
			)

			ch <- prometheus.NewMetricWithTimestamp(ts, m)
		}
	}
}
