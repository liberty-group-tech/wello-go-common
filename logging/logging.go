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
