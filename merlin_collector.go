package merlin

import (
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &collector{}

type collector struct {
	UptimeSecondsTotal *prometheus.Desc

	c *Client
}

func NewCollector(c *Client) prometheus.Collector {
	return &collector{
		UptimeSecondsTotal: prometheus.NewDesc(
			"merlin_ofdm_downstream_power_decibelvolts",
			"Power of downstream signal",
			[]string{"frequency"},
			nil,
		),
		c: c,
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range []*prometheus.Desc{
		c.UptimeSecondsTotal,
	} {
		ch <- d
	}
}

// Collect implements prometheus.Collector.
func (c *collector) Collect(ch chan<- prometheus.Metric) {

	ofdm, err := c.c.GetOfdmData()
	if err != nil {
		log.Print(err)
	}
	for _, d := range ofdm.Downstream {
		ch <- prometheus.MustNewConstMetric(
			c.UptimeSecondsTotal,
			prometheus.GaugeValue,
			d.DownstreamPwr/1000,
			fmt.Sprintf("%.f", d.ChannelFrequency),
		)
	}
}
