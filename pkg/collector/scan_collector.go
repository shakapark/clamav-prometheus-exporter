package collector

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// ClamavCollector satisfies prometheus.Collector interface
type ClamscanCollector struct {
	clamScanFilePath string
	up               *prometheus.Desc
}

// Describe satisfies prometheus.Collector.Describe
func (collector *ClamscanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.up
}

// Collect satisfies prometheus.Collector.Collect
func (collector *ClamscanCollector) Collect(ch chan<- prometheus.Metric) {
	if collector.clamScanFilePath == "" {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 0, "")
		return
	}

	file, err := os.Open(collector.clamScanFilePath)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 0, collector.clamScanFilePath)
		log.Debug(err)
		return
	} else {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 1, collector.clamScanFilePath)
	}
	defer file.Close()
}
