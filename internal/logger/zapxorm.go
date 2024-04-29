package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	xormlog "xorm.io/xorm/log"
)

type ZapXormLogger struct {
	level   xormlog.LogLevel
	showSQL bool
}

var _ xormlog.Logger = &ZapXormLogger{}

func (c *ZapXormLogger) Debug(v ...interface{}) {
	Log.Logger.Debug(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Debugf(format string, v ...interface{}) {
	Log.SugaredLogger.Debugf(format, v...)
}

func (c *ZapXormLogger) Error(v ...interface{}) {
	Log.Logger.Error(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Errorf(format string, v ...interface{}) {
	Log.SugaredLogger.Errorf(format, v...)
}

func (c *ZapXormLogger) Info(v ...interface{}) {
	Log.Logger.Info(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Infof(format string, v ...interface{}) {
	Log.SugaredLogger.Infof(format, v...)
}

func (c *ZapXormLogger) Warn(v ...interface{}) {
	Log.Logger.Warn(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Warnf(format string, v ...interface{}) {
	Log.SugaredLogger.Warnf(format, v...)
}

func (c *ZapXormLogger) Level() xormlog.LogLevel {
	return c.level
}

func (c *ZapXormLogger) SetLevel(l xormlog.LogLevel) {
	c.level = l
}

func (c *ZapXormLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		c.showSQL = true
		return
	}
	c.showSQL = show[0]
}

func (c *ZapXormLogger) IsShowSQL() bool {
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
