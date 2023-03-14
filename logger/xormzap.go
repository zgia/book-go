package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	xormlog "xorm.io/xorm/log"
)

type XormZapLogger struct {
	level   xormlog.LogLevel
	showSQL bool
}

var _ xormlog.Logger = &XormZapLogger{}

func (c *XormZapLogger) Debug(v ...interface{}) {
	Log.Logger.Debug(fmt.Sprint(v...))
}

func (c *XormZapLogger) Debugf(format string, v ...interface{}) {
	Log.SugaredLogger.Debugf(format, v...)
}

func (c *XormZapLogger) Error(v ...interface{}) {
	Log.Logger.Error(fmt.Sprint(v...))
}

func (c *XormZapLogger) Errorf(format string, v ...interface{}) {
	Log.SugaredLogger.Errorf(format, v...)
}

func (c *XormZapLogger) Info(v ...interface{}) {
	Log.Logger.Info(fmt.Sprint(v...))
}

func (c *XormZapLogger) Infof(format string, v ...interface{}) {
	Log.SugaredLogger.Infof(format, v...)
}

func (c *XormZapLogger) Warn(v ...interface{}) {
	Log.Logger.Warn(fmt.Sprint(v...))
}

func (c *XormZapLogger) Warnf(format string, v ...interface{}) {
	Log.SugaredLogger.Warnf(format, v...)
}

func (c *XormZapLogger) Level() xormlog.LogLevel {
	return c.level
}

func (c *XormZapLogger) SetLevel(l xormlog.LogLevel) {
	c.level = l
}

func (c *XormZapLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		c.showSQL = true
		return
	}
	c.showSQL = show[0]
}

func (c *XormZapLogger) IsShowSQL() bool {
	return c.showSQL
}

func ZapLevelToXorm(lvl zapcore.Level) xormlog.LogLevel {

	levelMappings := map[zapcore.Level]xormlog.LogLevel{
		zap.DebugLevel: xormlog.LOG_DEBUG,
		zap.InfoLevel:  xormlog.LOG_INFO,
		zap.WarnLevel:  xormlog.LOG_WARNING,
		zap.ErrorLevel: xormlog.LOG_ERR,
	}

	if v, ok := levelMappings[lvl]; ok {
		return v
	} else {
		return xormlog.LOG_INFO
	}
}
