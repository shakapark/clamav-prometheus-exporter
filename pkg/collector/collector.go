/*
Copyright 2020 Christian Niehoff.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"bytes"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shakapark/clamav-prometheus-exporter/pkg/clamav"
	"github.com/shakapark/clamav-prometheus-exporter/pkg/commands"
	log "github.com/sirupsen/logrus"
)

// ClamavCollector satisfies prometheus.Collector interface
type ClamavCollector struct {
	client      clamav.Client
	up          *prometheus.Desc
	threadsLive *prometheus.Desc
	threadsIdle *prometheus.Desc
	threadsMax  *prometheus.Desc
	queue       *prometheus.Desc
	memHeap     *prometheus.Desc
	memMmap     *prometheus.Desc
	memUsed     *prometheus.Desc
	poolsUsed   *prometheus.Desc
	poolsTotal  *prometheus.Desc
	buildInfo   *prometheus.Desc
	databaseAge *prometheus.Desc
}

// New creates a ClamavCollector struct
func New(client clamav.Client, report clamav.ScanReport) (*ClamavCollector, *ClamscanCollector) {
	return &ClamavCollector{
			client:      client,
			up:          prometheus.NewDesc("clamav_up", "Shows UP Status", nil, nil),
			threadsLive: prometheus.NewDesc("clamav_threads_live", "Shows live threads", nil, nil),
			threadsIdle: prometheus.NewDesc("clamav_threads_idle", "Shows idle threads", nil, nil),
			threadsMax:  prometheus.NewDesc("clamav_threads_max", "Shows max threads", nil, nil),
			queue:       prometheus.NewDesc("clamav_queue_length", "Shows queued items", nil, nil),
			memHeap:     prometheus.NewDesc("clamav_mem_heap_bytes", "Shows heap memory usage in bytes", nil, nil),
			memMmap:     prometheus.NewDesc("clamav_mem_mmap_bytes", "Shows mmap memory usage in bytes", nil, nil),
			memUsed:     prometheus.NewDesc("clamav_mem_used_bytes", "Shows used memory in bytes", nil, nil),
			poolsUsed:   prometheus.NewDesc("clamav_pools_used_bytes", "Shows memory used by memory pool allocator for the signature database in bytes", nil, nil),
			poolsTotal:  prometheus.NewDesc("clamav_pools_total_bytes", "Shows total memory allocated by memory pool allocator for the signature database in bytes", nil, nil),
			buildInfo:   prometheus.NewDesc("clamav_build_info", "Shows ClamAV Build Info", []string{"clamav_version", "database_version"}, nil),
			databaseAge: prometheus.NewDesc("clamav_database_age", "Shows ClamAV signature database age in seconds", nil, nil),
		}, &ClamscanCollector{
			clamScanReport: report,
			up:             prometheus.NewDesc("clamscan_report_file", "Shows if report file is found", []string{"file_path"}, nil),
			countLine:      prometheus.NewDesc("clamscan_report_file_count_line", "Shows how many line has been read report file", nil, nil),
		}
}

// Describe satisfies prometheus.Collector.Describe
func (collector *ClamavCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.up
	ch <- collector.threadsLive
	ch <- collector.threadsIdle
	ch <- collector.threadsMax
	ch <- collector.queue
	ch <- collector.memHeap
	ch <- collector.memMmap
	ch <- collector.memUsed
	ch <- collector.poolsUsed
	ch <- collector.poolsTotal
	ch <- collector.buildInfo
	ch <- collector.databaseAge
}

// Collect satisfies prometheus.Collector.Collect
func (collector *ClamavCollector) Collect(ch chan<- prometheus.Metric) {
	pong := collector.client.Dial(commands.PING)
	if bytes.Equal(pong, []byte{'P', 'O', 'N', 'G', '\n'}) {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 1)
	} else {
		ch <- prometheus.MustNewConstMetric(collector.up, prometheus.GaugeValue, 0)
	}

	stats := collector.client.Dial(commands.STATS)
	idle, err := regexp.MatchString("IDLE", string(stats))
	if err != nil {
		log.Errorf(`error searching IDLE field in stats %v: %s`, idle, err)
		return
	}

	collector.CollectMemoryStats(ch, string(stats))
	collector.CollectThreads(ch, string(stats))
	collector.CollectQueue(ch, string(stats))
	collector.CollectBuildInfo(ch)
}

func float(s string) float64 {
	float, err := strconv.ParseFloat(s, 64)
	if err != nil {
		float = math.NaN()
	}
	return float
}

func (collector *ClamavCollector) CollectMemoryStats(ch chan<- prometheus.Metric, stats string) {
	regex := regexp.MustCompile(`(?:MEMSTATS:\sheap|mmap|used|free|releasable|pools|pools_used|pools_total)\s+([0-9.]+|N\/A)+`)
	matches := regex.FindAllStringSubmatch(stats, -1)

	log.Debug("Matches Memory Stats", matches)

	// MEMORY STATS
	if len(matches) > 0 {
		ch <- prometheus.MustNewConstMetric(collector.memHeap, prometheus.GaugeValue, float(matches[0][1])*1024)
		log.Debug("memHeap: ", float(matches[1][1])*1024)
		ch <- prometheus.MustNewConstMetric(collector.memMmap, prometheus.GaugeValue, float(matches[1][1])*1024)
		log.Debug("memMmap: ", float(matches[2][1])*1024)
		ch <- prometheus.MustNewConstMetric(collector.memUsed, prometheus.GaugeValue, float(matches[2][1])*1024)
		log.Debug("memUsed: ", float(matches[3][1])*1024)
		ch <- prometheus.MustNewConstMetric(collector.poolsUsed, prometheus.GaugeValue, float(matches[6][1])*1024)
		log.Debug("poolsUsed: ", float(matches[6][1])*1024)
		ch <- prometheus.MustNewConstMetric(collector.poolsTotal, prometheus.GaugeValue, float(matches[7][1])*1024)
		log.Debug("poolsTotal: ", float(matches[7][1])*1024)
	}
}

func (collector *ClamavCollector) CollectThreads(ch chan<- prometheus.Metric, stats string) {
	regex := regexp.MustCompile(`(?:THREADS:\slive|idle|max|idle-timeout)\s+([0-9.]+|N\/A)+`)
	matches := regex.FindAllStringSubmatch(stats, -1)

	log.Debug("Matches Threads", matches)

	// THREADS
	if len(matches) > 0 {
		ch <- prometheus.MustNewConstMetric(collector.threadsLive, prometheus.GaugeValue, float(matches[0][1]))
		log.Debug("threadsLive: ", float(matches[0][1]))
		ch <- prometheus.MustNewConstMetric(collector.threadsIdle, prometheus.GaugeValue, float(matches[1][1]))
		log.Debug("threadsIdle: ", float(matches[1][1]))
		ch <- prometheus.MustNewConstMetric(collector.threadsMax, prometheus.GaugeValue, float(matches[2][1]))
		log.Debug("threadsMax: ", float(matches[2][1]))
	}
}

func (collector *ClamavCollector) CollectQueue(ch chan<- prometheus.Metric, stats string) {
	regex := regexp.MustCompile(`(?:QUEUE:|FILDES|STATS)\s+([0-9.]+|N\/A)`)
	matches := regex.FindAllStringSubmatch(stats, -1)

	log.Debug("Matches Queue", matches)

	// QUEUE
	if len(matches) > 0 {
		ch <- prometheus.MustNewConstMetric(collector.queue, prometheus.GaugeValue, float(matches[0][1]))
		log.Debug("queue: ", float(matches[0][1]))
	}
}

func (collector *ClamavCollector) CollectBuildInfo(ch chan<- prometheus.Metric) {
	// The return of this should be something like: ClamAV 1.4.1/27523/Sun Jan 19 09:40:50 2025
	version := collector.client.Dial(commands.VERSION)
	regex := regexp.MustCompile(`ClamAV\s([0-9.]*)/(\d+)/(.+)`)

	// The match will be a list of four elements:
	// length=4 => [0]: ClamAV, [1]: 1.4.1, [2]: 27523, [3]: Sun Jan 19 09:40:50 2025
	matches := regex.FindStringSubmatch(string(version))

	log.Debug("Matches Version", matches)

	if len(matches) >= 3 {
		ch <- prometheus.MustNewConstMetric(collector.buildInfo, prometheus.GaugeValue, 1, matches[1], matches[2])

		strBuilddate := time.Now().UTC().String()

		// If the regular expression match returns a list with length 4,
		// it means that the VERSION command return has a date.
		// Otherwise, it ignores it and uses the current time.
		if len(matches) == 4 {
			strBuilddate = matches[3]
		}

		// Parse string as date type based on RFC850
		builddate, err := time.Parse("Mon Jan 2 15:04:05 2006", strBuilddate)

		if err != nil {
			log.Error("Error parsing ClamAV date: ", err)
		}

		ch <- prometheus.MustNewConstMetric(collector.databaseAge, prometheus.GaugeValue, time.Since(builddate).Seconds())
	}
}
