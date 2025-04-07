package clamav

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ScanReport corresponds to a ClamScan report file
type ScanReport struct {
	filePath         string
	countLineRead    int
	countLineParsed  int
	countLineIgnored int
	countLineUnknown int
	reportStatus     bool
	totalErrors      int
	infectedFiles    int
	scanDuration     time.Duration
	scanStartTime    time.Time
	scanEndTime      time.Time
	errFile          error
}

// NewScanReport create a new ScanReport
func NewScanReport(path string) *ScanReport {
	return &ScanReport{
		filePath:         path,
		countLineRead:    0,
		countLineParsed:  0,
		countLineIgnored: 0,
		countLineUnknown: 0,
		reportStatus:     false,
		totalErrors:      0,
		infectedFiles:    0,
		scanDuration:     0,
		scanStartTime:    time.Now(),
		scanEndTime:      time.Now(),
		errFile:          nil,
	}
}

// Get functions
func (sr *ScanReport) GetFilepath() string {
	return sr.filePath
}
func (sr *ScanReport) GetLineCount() int {
	return sr.countLineRead
}
func (sr *ScanReport) GetParsedLineCount() int {
	return sr.countLineParsed
}
func (sr *ScanReport) GetIgnoredLineCount() int {
	return sr.countLineIgnored
}
func (sr *ScanReport) GetUnknownLineCount() int {
	return sr.countLineUnknown
}
func (sr *ScanReport) GetReportStatus() bool {
	return sr.reportStatus
}
func (sr *ScanReport) GetIntReportStatus() int {
	if sr.reportStatus {
		return 1
	} else {
		return 0
	}
}
func (sr *ScanReport) GetTotalErrors() int {
	return sr.totalErrors
}
func (sr *ScanReport) GetInfectedFiles() int {
	return sr.infectedFiles
}
func (sr *ScanReport) GetScanDuration() time.Duration {
	return sr.scanDuration
}
func (sr *ScanReport) GetScanStartTime() time.Time {
	return sr.scanStartTime
}
func (sr *ScanReport) GetScanEndTime() time.Time {
	return sr.scanEndTime
}
func (sr *ScanReport) GetErrFile() error {
	return sr.errFile
}

// Set functions
func (sr *ScanReport) setReportStatus(b bool) {
	sr.reportStatus = b
}
func (sr *ScanReport) setTotalErrors(i int) {
	sr.totalErrors = i
}
func (sr *ScanReport) setInfectedFiles(i int) {
	sr.infectedFiles = i
}
func (sr *ScanReport) setScanDuration(d time.Duration) {
	sr.scanDuration = d
}
func (sr *ScanReport) setScanStartTime(t time.Time) {
	sr.scanStartTime = t
}
func (sr *ScanReport) setScanEndTime(t time.Time) {
	sr.scanEndTime = t
}

// Increase function for count variables
func (sr *ScanReport) increaseLineCount(i int) {
	sr.countLineRead = sr.countLineRead + i
}
func (sr *ScanReport) increaseParsedLineCount(i int) {
	sr.countLineParsed = sr.countLineParsed + i
}
func (sr *ScanReport) increaseIgnoredLineCount(i int) {
	sr.countLineIgnored = sr.countLineIgnored + i
}
func (sr *ScanReport) increaseUnknownLineCount(i int) {
	sr.countLineUnknown = sr.countLineUnknown + i
}

func (sr *ScanReport) Tail() {
	log.Debug("Begin to read file: " + sr.filePath)

	file, err := os.Open(sr.filePath)
	if err != nil {
		log.Error("Error reading file: ", err)
		sr.errFile = err
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// without this sleep you would hogg the CPU
				time.Sleep(500 * time.Millisecond)
				// truncated ?
				truncated, errTruncated := isTruncated(file)
				if errTruncated != nil {
					break
				}
				if truncated {
					// seek from start
					_, errSeekStart := file.Seek(0, io.SeekStart)
					if errSeekStart != nil {
						break
					}
				}
				continue
			}
			break
		}
		log.Debug("New line read: " + line)
		// Parse line
		sr.parseLine(line)

		sr.increaseLineCount(1)
		// cl := *sr.countLineRead + 1
		// sr.countLineRead = &cl
	}

}

func isTruncated(file *os.File) (bool, error) {
	// current read position in a file
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}
	// file stat to get the size
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	return currentPos > fileInfo.Size(), nil
}

func (sr *ScanReport) parseLine(l string) {
	// List of ignoredLines
	if l == "--------------------------------------\n" || l == "----------- SCAN SUMMARY -----------\n" || l == "\n" || strings.Contains(l, "ERROR: Could not connect to clamd") {
		sr.increaseIgnoredLineCount(1)
		return
	}

	// List of parsedLines
	if reportStatusHostFS, b := strings.CutPrefix(l, "/host-fs: "); b {
		log.Debug("HostFS report status: ", reportStatusHostFS)

		if reportStatusHostFS == "OK" {
			sr.setReportStatus(true)
		} else {
			sr.setReportStatus(false)
		}

		sr.increaseParsedLineCount(1)
		return
	}
	if reportInfectedFiles, b := strings.CutPrefix(l, "Infected files: "); b {
		log.Debug("Report infected files: ", reportInfectedFiles)
		infectedFiles, errConvInt := strconv.Atoi(reportInfectedFiles)
		if errConvInt != nil {
			log.Error("Error converting infected files to int: ", errConvInt)
		}
		sr.setInfectedFiles(infectedFiles)
		sr.increaseParsedLineCount(1)
		return
	}
	if reportTotalErrors, b := strings.CutPrefix(l, "Total errors: "); b {
		log.Debug("Report errors: ", reportTotalErrors)
		totalErrors, errConvInt := strconv.Atoi(reportTotalErrors)
		if errConvInt != nil {
			log.Error("Error converting total errors to int: ", errConvInt)
		}
		sr.setTotalErrors(totalErrors)
		sr.increaseParsedLineCount(1)
		return
	}
	if reportTime, b := strings.CutPrefix(l, "Time: "); b {
		log.Debug("Report time: ", reportTime)
		duration, errParseDuration := time.ParseDuration(strings.Split(reportTime, " ")[0] + "s")
		if errParseDuration != nil {
			log.Error("Error converting report time to duration: ", errParseDuration)
		}
		sr.setScanDuration(duration)
		sr.increaseParsedLineCount(1)
		return
	}
	if reportStartDate, b := strings.CutPrefix(l, "Start Date: "); b {
		log.Debug("Report start date: ", reportStartDate)
		startDate, errParseTime := time.Parse("2006:01:02 15:04:05", reportStartDate)
		if errParseTime != nil {
			log.Error("Error converting report time to duration: ", errParseTime)
		}
		sr.setScanStartTime(startDate)
		sr.increaseParsedLineCount(1)
		return
	}
	if reportEndDate, b := strings.CutPrefix(l, "End Date: "); b {
		log.Debug("Report end date: ", reportEndDate)
		endDate, errParseTime := time.Parse("2006:01:02 15:04:05", reportEndDate)
		if errParseTime != nil {
			log.Error("Error converting report time to duration: ", errParseTime)
		}
		sr.setScanEndTime(endDate)
		sr.increaseParsedLineCount(1)
		return
	}

	// Return unknown lines
	log.Error("Unknown line: " + l)
	sr.increaseUnknownLineCount(1)
	// return
}

// --------------------------------------
// /host-fs: OK

// ----------- SCAN SUMMARY -----------
// Infected files: 0
// Total errors: 2
// Time: 3609.617 sec (60 m 9 s)
// Start Date: 2025:03:27 16:14:48
// End Date:   2025:03:27 17:14:58
