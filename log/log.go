package log

import (
	"conf"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var flyLogger *logrus.Logger
var once sync.Once

type Fields map[string]interface{}

type logFileWriter struct {
	file *os.File
	size int64
}

func (p *logFileWriter) Write(data []byte) (n int, err error) {
	if p == nil {
		return 0, errors.New("logFileWriter is nil")
	}
	if p.file == nil {
		return 0, errors.New("file not opened")
	}
	n, e := p.file.Write(data)
	p.size += int64(n)
	//文件最大 64 K byte
	if p.size > 1024*64 {
		p.file.Close()
		p.file, _ = os.OpenFile(conf.LogPath+strconv.FormatInt(time.Now().Unix(), 10), os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
		p.size = 0
	}
	return n, e
}

func GetInstance() *logrus.Logger {
	return flyLogger
}

func init() {
	flyLogger = new(logrus.Logger)
	file, err := os.OpenFile(conf.LogPath+strconv.FormatInt(time.Now().Unix(), 10), os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
	if err == nil {
		info, err := file.Stat()
		if err == nil {
			fileWriter := logFileWriter{file, info.Size()}
			flyLogger.SetOutput(&fileWriter)
		} else {
			flyLogger.Out = os.Stdout
		}
	} else {
		flyLogger.Out = os.Stdout
	}
	flyLogger.Formatter = &logrus.TextFormatter{}
	flyLogger.SetLevel(logrus.TraceLevel)
}
