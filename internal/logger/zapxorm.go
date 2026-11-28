package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	xormlog "xorm.io/xorm/log"
)

// ZapXormLogger 把 xorm 的 SQL 日志接到 zap 上。
// 注意：xorm 的 SetLogLevel 只会调用这里的 SetLevel，是否真正输出由
// 下面各方法自己判断 —— 不判断的话级别形同虚设（zap 全局是 Debug，
// 每条 SQL 连同参数都会打出来，正文 2-gram 能写出几百兆日志）。
type ZapXormLogger struct {
	level   xormlog.LogLevel
	showSQL bool
}

var _ xormlog.Logger = &ZapXormLogger{}

func (c *ZapXormLogger) Debug(v ...interface{}) {
	if c.level > xormlog.LOG_DEBUG {
		return
	}
	Log.Logger.Debug(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Debugf(format string, v ...interface{}) {
	if c.level > xormlog.LOG_DEBUG {
		return
	}
	Log.SugaredLogger.Debugf(format, v...)
}

func (c *ZapXormLogger) Error(v ...interface{}) {
	if c.level > xormlog.LOG_ERR {
		return
	}
	Log.Logger.Error(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Errorf(format string, v ...interface{}) {
	if c.level > xormlog.LOG_ERR {
		return
	}
	Log.SugaredLogger.Errorf(format, v...)
}

func (c *ZapXormLogger) Info(v ...interface{}) {
	if c.level > xormlog.LOG_INFO {
		return
	}
	Log.Logger.Info(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Infof(format string, v ...interface{}) {
	if c.level > xormlog.LOG_INFO {
		return
	}
	Log.SugaredLogger.Infof(format, v...)
}

func (c *ZapXormLogger) Warn(v ...interface{}) {
	if c.level > xormlog.LOG_WARNING {
		return
	}
	Log.Logger.Warn(fmt.Sprint(v...))
}

func (c *ZapXormLogger) Warnf(format string, v ...interface{}) {
	if c.level > xormlog.LOG_WARNING {
		return
	}
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
