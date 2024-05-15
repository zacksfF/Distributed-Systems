package lowwatermark

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

const (
	TXNLOG_MAGIC                                  = 0x766b6c67 // "ZKLG" in integer format
	VERSION                                       = 2
	LOG_FILE_PREFIX                               = "log"
	FSYNC_WARNING_THRESHOLD_MS_PROPERTY           = "fsync.warningthresholdms"
	ZOOKEEPER_FSYNC_WARNING_THRESHOLD_MS_PROPERTY = "zookeeper." + FSYNC_WARNING_THRESHOLD_MS_PROPERTY
	txnLogSizeLimitSetting                        = "zookeeper.txnLogSizeLimitInKb"
)

var (
	fsyncWarningThresholdMS int64
	txnLogSizeLimit         int64
	filePadding             = FilePadding{} // struct for file padding functionality (implementation needed)
)

type FileTxnLog struct {
	lastZxidSeen         int64
	logStream            *bufio.Writer // replaces BufferedOutputStream
	oa                   OutputArchive // type definition needed based on your implementation
	fos                  *os.File      // replaces FileOutputStream
	logDir               string
	forceSync            bool
	dbId                 int64
	streamsToFlush       chan *os.File // channel for file streams to flush
	logFileWrite         *os.File      // current log file being written to
	fileSize             int64
	unFlushedSize        int64
	filePosition         int64
	prevLogsRunningTotal int64
	serverStats          *ServerStats // type definition needed
	syncElapsedMS        int64
}

type ServerStats interface {
	// Get the number of connected clients
	GetNumClients() int

	// Get the number of active transactions
	GetNumTxn() int

	// Get the average latency of requests
	GetAvgLatency() float64
}

type OutputArchive interface {
	// Write a byte slice to the archive
	WriteBytes([]byte) error

	// Write an integer to the archive
	WriteInt(int) error

	// Write a long integer to the archive
	WriteLong(int64) error
}

type FilePadding struct { // Needs implementation based on your padding logic
	preallocSize int64
}

// type TxnLog struct{
// 	sessioned [8]byte
// 	cxId [4]byte
// 	ZxId [8]byte
// 	time time.Time
// 	Type [4]byte
// }

type TxnLog interface {
	// Append a transaction entry to the log
	AppendEntry(txn []byte) error
	// Close the transaction log and release resources
	Close() error
	//Synchronize the log data to disk
	Sync() error
}

// func (f *FileTnxLog) Implement(i interface{}) error {
// 	_, ok := i.(TxnLog)
// }

func init() {
	// Read fsync.warningthresholdms property
	thresholdStr := os.Getenv(ZOOKEEPER_FSYNC_WARNING_THRESHOLD_MS_PROPERTY)
	if thresholdStr == "" {
		thresholdStr = os.Getenv(FSYNC_WARNING_THRESHOLD_MS_PROPERTY)
	}
	threshold, err := strconv.ParseInt(thresholdStr, 10, 64)
	if err != nil {
		threshold = 1000 // default value
	}
	fsyncWarningThresholdMS = threshold

	// Read txnLogSizeLimit property
	sizeStr := os.Getenv(txnLogSizeLimitSetting)
	if sizeStr != "" {
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing %s: %v", txnLogSizeLimitSetting, err)
		} else {
			txnLogSizeLimit = size * 1024 // convert KB to bytes
			log.Printf("%s=%d", txnLogSizeLimitSetting, txnLogSizeLimit)
		}
	}
}

func (fp *FilePadding) SetPreallocSize(size int64) {
	fp.preallocSize = size
}

// Interface implementations (assuming TxnLog and io.Closer are defined elsewhere)
func (f *FileTxnLog) Implement(i interface{}) error {
	_, ok := i.(TxnLog)
	if !ok {
		return errors.New("FileTxnLog does not implement the provided TxnLog interface")
	}
	_, ok = i.(io.Closer)
	if !ok {
		return errors.New("FileTxnLog does not implement the provided io.Closer interface")
	}
	return nil
}

func (f *FileTxnLog) Close() error {
	var errs []error
	if f.logStream != nil {
		err := f.logStream.Flush()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if f.fos != nil {
		err := f.fos.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	for f := range f.streamsToFlush {
		err := f.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		var combinedErrMsg string
		for _, err := range errs {
			combinedErrMsg += err.Error() + "; "
		}
		return fmt.Errorf("Errors closing FileTxnLog: %s", combinedErrMsg[:len(combinedErrMsg)-2]) // remove trailing semicolon and space
	}
	return nil
}
