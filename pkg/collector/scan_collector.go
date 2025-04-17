package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shakapark/clamav-prometheus-exporter/pkg/clamav"
)

// ClamavCollector satisfies prometheus.Collector interface
type ClamscanCollector struct {
	clamScanReport        *clamav.ScanReport
	up                    *prometheus.Desc
	countLine             *prometheus.Desc
	lastScanStartTime     *prometheus.Desc
	lastScanEndTime       *prometheus.Desc
	lastScanDuration      *prometheus.Desc
	lastScanStatus        *prometheus.Desc
	lastScanInfectedFiles *prometheus.Desc
	lastScanErrors        *prometheus.Desc
}

// Describe satisfies prometheus.Collector.Describe
func (collector *ClamscanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.up
	ch <- collector.countLine
	ch <- collector.lastScanStartTime
	ch <- collector.lastScanEndTime
	ch <- collector.lastScanDuration
	ch <- collector.lastScanStatus
	ch <- collector.lastScanInfectedFiles
	ch <- collector.lastScanErrors
}

// Collect satisfies prometheus.Collector.Collect
func (collector *ClamscanCollector) Collect(ch chan<- prometheus.Metric) {
	if collector.clamScanReport.GetFilepath() == "" {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 0, "")
		return
	}

	if collector.clamScanReport.GetErrFile() != nil {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 0, collector.clamScanReport.GetFilepath())
		return
	}

	ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 1, collector.clamScanReport.GetFilepath())
	ch <- prometheus.MustNewConstMetric(collector.countLine, prometheus.GaugeValue, float64(collector.clamScanReport.GetLineCount()), "total")
	ch <- prometheus.MustNewConstMetric(collector.countLine, prometheus.GaugeValue, float64(collector.clamScanReport.GetParsedLineCount()), "parsed")
	ch <- prometheus.MustNewConstMetric(collector.countLine, prometheus.GaugeValue, float64(collector.clamScanReport.GetIgnoredLineCount()), "ignored")
	ch <- prometheus.MustNewConstMetric(collector.countLine, prometheus.GaugeValue, float64(collector.clamScanReport.GetUnknownLineCount()), "unknown")
	ch <- prometheus.MustNewConstMetric(collector.lastScanStartTime, prometheus.GaugeValue, float64(collector.clamScanReport.GetScanStartTime().Unix()))
	ch <- prometheus.MustNewConstMetric(collector.lastScanEndTime, prometheus.GaugeValue, float64(collector.clamScanReport.GetScanEndTime().Unix()))
	ch <- prometheus.MustNewConstMetric(collector.lastScanDuration, prometheus.GaugeValue, collector.clamScanReport.GetScanDuration().Seconds())
	ch <- prometheus.MustNewConstMetric(collector.lastScanStatus, prometheus.GaugeValue, float64(collector.clamScanReport.GetIntReportStatus()))
	ch <- prometheus.MustNewConstMetric(collector.lastScanInfectedFiles, prometheus.GaugeValue, float64(collector.clamScanReport.GetInfectedFiles()))
	ch <- prometheus.MustNewConstMetric(collector.lastScanErrors, prometheus.GaugeValue, float64(collector.clamScanReport.GetTotalErrors()))
}
