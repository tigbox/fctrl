package fctrl

var (
	Loger = newLogerManager()
)

type LogerManager struct {
	enable bool
	loger  ILoger
}

func newLogerManager() *LogerManager {
	return &LogerManager{
		enable: true,
		loger:  newDefaultLoger(),
	}
}

func (m *LogerManager) Info(args ...interface{}) {
	m.printLog(m.loger.Info, args)
}

func (m *LogerManager) Warn(args ...interface{}) {
	m.printLog(m.loger.Warn, args...)
}

func (m *LogerManager) Error(args ...interface{}) {
	m.printLog(m.loger.Error, args...)
}

func (m *LogerManager) Fatal(args ...interface{}) {
	m.printLog(m.loger.Fatal, args...)
}

func (m *LogerManager) Infof(format string, args ...interface{}) {
	m.printfLog(m.loger.Infof, format, args...)
}

func (m *LogerManager) Warnf(format string, args ...interface{}) {
	m.printfLog(m.loger.Warnf, format, args...)
}

func (m *LogerManager) Errorf(format string, args ...interface{}) {
	m.printfLog(m.loger.Errorf, format, args...)
}

func (m *LogerManager) Fatalf(format string, args ...interface{}) {
	m.printfLog(m.loger.Fatalf, format, args...)
}

func (m *LogerManager) printLog(logFunc func(...interface{}), args ...interface{}) {
	if !m.enable {
		return
	}
	if logFunc == nil {
		return
	}
	logFunc(args...)
}

func (m *LogerManager) printfLog(logFunc func(string, ...interface{}), format string, args ...interface{}) {
	if !m.enable {
		return
	}
	if logFunc == nil {
		return
	}
	logFunc(format, args...)
}

func (m *LogerManager) SetLoger(l ILoger) {
	m.loger = l
}

func (m *LogerManager) Enable(isEnable bool) {
	m.enable = isEnable
}

func SetLoger(l ILoger) {
	Loger.SetLoger(l)
}

func LogerEnable(isEnable bool) {
	Loger.Enable(isEnable)
}
