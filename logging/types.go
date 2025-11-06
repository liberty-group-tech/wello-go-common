package logging

// Logger 日志接口
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// NoOpLogger 空操作日志实现
type NoOpLogger struct{}

func (l *NoOpLogger) Errorf(format string, args ...interface{}) {}
func (l *NoOpLogger) Infof(format string, args ...interface{})  {}
func (l *NoOpLogger) Debugf(format string, args ...interface{}) {}
func (l *NoOpLogger) Error(format string, args ...interface{})  {}
func (l *NoOpLogger) Info(format string, args ...interface{})   {}
func (l *NoOpLogger) Debug(format string, args ...interface{})  {}

type LogBase struct {
	// required by keter
	Timestamp string `json:"@timestamp"`
	AppName   string `json:"appName,omitempty"`

	// common
	Level     string `json:"level"`
	Message   string `json:"message"`
	Env       string `json:"env,omitempty"`
	Error     error  `json:"error,omitempty"`
	Data      any    `json:"data,omitempty"`
	RequestId string `json:"requestId,omitempty"`
	TraceID   string `json:"traceId,omitempty"`
	SpanID    string `json:"spanId,omitempty"`
	Service   string `json:"service,omitempty"`
	Caller    string `json:"caller,omitempty"`
	Module    string `json:"module,omitempty"`
	Hostname  string `json:"host,omitempty"`
	PID       string `json:"pid,omitempty"`
}
