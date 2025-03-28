package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shakapark/clamav-prometheus-exporter/pkg/clamav"
)

// ClamavCollector satisfies prometheus.Collector interface
type ClamscanCollector struct {
	clamScanReport *clamav.ScanReport
	up             *prometheus.Desc
	countLine      *prometheus.Desc
}

// Describe satisfies prometheus.Collector.Describe
func (collector *ClamscanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.up
	ch <- collector.countLine
}

// Collect satisfies prometheus.Collector.Collect
func (collector *ClamscanCollector) Collect(ch chan<- prometheus.Metric) {
	if collector.clamScanReport.GetFilepath() == "" {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 0, "")
		return
	}

	if collector.clamScanReport.GetErrFile() != nil {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 0, "")
	}

	ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 1, collector.clamScanReport.GetFilepath())
	ch <- prometheus.MustNewConstMetric(collector.countLine, prometheus.GaugeValue, float64(collector.clamScanReport.GetLineCount()))

}

// ----------- SCAN SUMMARY -----------
// Infected files: 0
// Total errors: 2
// Time: 3609.617 sec (60 m 9 s)
// Start Date: 2025:03:27 16:14:48
// End Date:   2025:03:27 17:14:58
