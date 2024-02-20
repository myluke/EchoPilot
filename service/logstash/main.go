package logstash

import (
	"errors"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/gommon/log"
	"github.com/mylukin/EchoPilot/helper"
	"github.com/telkomdev/go-stash"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var ErrorLogServerNoSet = errors.New("LOG_SERVER is not set")

// LogServer
var LogServer string

// client is a variable to store logstash client
var client *stash.Stash

// clientError is a variable to store logstash client error
var clientErr error

// clientLock is a variable to store logstash client lock
var clientLock sync.RWMutex

// clientCH is a variable to store logstash client channel
var clientCH chan Log

// Log
type Log struct {
	Path  string    `json:"path"`
	Value int64     `json:"value"`
	Time  time.Time `json:"addtime"`
}

func init() {
	var processNum int64 = 3
	logProcessNum := strings.Trim(helper.Config("LOG_PROCESS_NUM"), `"`)
	if logProcessNum != "" {
		processNum = helper.ToInt64(logProcessNum)
	}
	clientCH = make(chan Log, 100000)

	// 开启发送线程
	for i := 0; i < int(processNum); i++ {
		go func() {
			clientErr = Connect()
			for {
				select {
				case logData := <-clientCH:
					sendTo(logData)
				}
			}
		}()
	}
}

// Send is a function to send log to logstash
func Send(logData Log) bool {
	select {
	case clientCH <- logData:
		return true
	default:
		return false
	}
}

// Connect is a function to connect to logstash
func Connect() error {
	LogServer = strings.Trim(helper.Config("LOG_SERVER"), `"`)
	if LogServer == "" {
		return ErrorLogServerNoSet
	}

	var port uint64 = 8888
	var host string = "localhost"
	if pos := strings.Index(LogServer, ":"); pos > -1 {
		port = uint64(helper.ToInt64(LogServer[pos+1:]))
		host = LogServer[:pos]
	}
	client, clientErr = stash.Connect(host, port, stash.SetWriteTimeout(10*time.Second))
	if clientErr != nil {
		return clientErr
	}
	return nil
}

// sendTo is a function to send log to logstash
func sendTo(logData Log) error {
	if clientErr != nil {
		log.Errorf("server: %s, err: %s", LogServer, clientErr)
		return clientErr
	}

	if logData.Value == 0 {
		logData.Value = 1
	}
	logData.Time = time.Now()
	logDataJSON, err := json.Marshal(logData)
	if err != nil {
		log.Errorf("server: %s, err: %s", LogServer, err)
		return err
	}

	clientLock.Lock()
	defer clientLock.Unlock()
	retry := 0
RetryWrite:
	_, err = client.Write(logDataJSON)
	if err != nil {
		retry++
		log.Errorf("server: %s, err: %s", LogServer, err)
		clientErr = Connect()
		if clientErr != nil {
			log.Errorf("server: %s, err: %s", LogServer, clientErr)
			return clientErr
		}
		// 最多重试3次
		if retry < 3 {
			goto RetryWrite
		}
	}
	return nil
}

// Close is a function to close logstash client
func Close() error {
	return client.Close()
}
