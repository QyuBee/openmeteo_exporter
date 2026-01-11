package main

import (
	"fmt"
	"strings"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type SatelliteRadiationCollector struct {
	Client   *OpenMeteoClient
	Location *LocationConfig
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

	ch <- prometheus.MustNewConstMetric(
		satelliteGenerationTimeDesc,
		prometheus.GaugeValue,
		float64(satelliteResp.GenerationtimeMs),
		c.Location.Name,
	)

	for _, name := range c.Location.SatelliteRadiation.Variables {
		units := satelliteResp.CurrentUnits.Variables[name].(string)
		if units == "W/m²" {
			units = "w_per_m2"
		}
		units = strings.ToLower(units)

		description, _ := GetVariableDesc("satellite_radiation", name)
		desc := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "satellite_radiation", fmt.Sprintf("%s_%s", name, units)),
			description,
			[]string{"location"},
			nil,
		)

		if value := satelliteResp.Current.Variables[name]; value != nil {
			ch <- prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				float64(value.(float64)),
				c.Location.Name,
			)
		} else {
			level.Warn(logger).Log("msg", "No value for metric returned", "name", name)
		}
	}
}
