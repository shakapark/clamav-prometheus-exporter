package clamav

import (
	"bufio"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// ScanReport corresponds to a ClamScan report file
type ScanReport struct {
	filePath      string
	countLineRead *int
	errFile       *error
}

// NewScanReport create a new ScanReport
func NewScanReport(path string) *ScanReport {
	clr := 0
	return &ScanReport{
		filePath:      path,
		countLineRead: &clr,
		errFile:       nil,
	}
}

func (sr *ScanReport) GetFilepath() string {
	return sr.filePath
}

func (sr *ScanReport) GetLineCount() *int {
	return sr.countLineRead
}

func (sr *ScanReport) GetErrFile() *error {
	return sr.errFile
}

func (sr *ScanReport) Tail() {
	log.Debug("Begin to read file: " + sr.filePath)

	file, err := os.Open(sr.filePath)
	if err != nil {
		log.Error("Error reading file: ", err)
		sr.errFile = &err
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
		cl := *sr.countLineRead + 1
		sr.countLineRead = &cl
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
