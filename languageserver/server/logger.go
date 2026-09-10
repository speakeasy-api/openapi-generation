package server

type Logger interface {
	Error(message string, keysAndValues ...any)
	Errorf(format string, args ...any)
	Warning(message string, keysAndValues ...any)
	Warningf(format string, args ...any)
	Info(message string, keysAndValues ...any)
	Infof(format string, args ...any)
	Debug(message string, keysAndValues ...any)
	Debugf(format string, args ...any)
}
