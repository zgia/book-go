package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/ini.v1"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	sync.Mutex
	*zap.Logger
	*zap.SugaredLogger

	RootDir    string
	Level      zapcore.Level
	TimeFormat string

	inited bool
}

var Log = &Logger{}

func NewConsole() {
	Log.Logger = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			os.Stdout,
			zapcore.DebugLevel,
		),
	)
	Log.SugaredLogger = Log.Logger.Sugar()
	defer Log.SugaredLogger.Sync()
}

func InitLogger(cfg *ini.File, logdir string, timeFormat string) {
	Log.Lock()
	defer Log.Unlock()

	if Log.inited {
		return
	}

	Log.TimeFormat = timeFormat
	Log.RootDir = logdir

	modes := strings.Split(cfg.Section("log").Key("MODE").MustString("console"), ",")
	logLevel := GetLevel(cfg.Section("log").Key("LEVEL").MustString("debug"))
	Log.Level = logLevel

	cores := []zapcore.Core{}

	for i := range modes {
		modes[i] = strings.ToLower(strings.TrimSpace(modes[i]))
		secName := "log." + modes[i]
		sec, err := cfg.GetSection(secName)
		if err != nil {
			panic(fmt.Sprintf("missing configuration section [%s] for %q logger", secName, modes[i]))
		}

		// Iterate over [log.*] sections to initialize individual logger
		level := GetLevel(sec.Key("LEVEL").MustString(logLevel.String()))

		switch modes[i] {
		case "file":
			lumberJackLogger := &lumberjack.Logger{
				Filename:   filepath.Join(Log.RootDir, sec.Key("LOG_NAME").MustString("book.log")),
				MaxSize:    sec.Key("MAX_SIZE").MustInt(100), // MB
				MaxBackups: sec.Key("MAX_BACKUPS").MustInt(10),
				MaxAge:     sec.Key("MAX_DAYS").MustInt(7), // days
				Compress:   false,
			}

			zec := zap.NewProductionEncoderConfig()
			zec.EncodeTime = encodeTime
			zec.EncodeCaller = zapcore.FullCallerEncoder
			fileEcoder := zapcore.NewJSONEncoder(zec)
			cores = append(cores, zapcore.NewCore(fileEcoder, zapcore.AddSync(lumberJackLogger), level))

		case "console":
			zec := zap.NewDevelopmentEncoderConfig()
			zec.EncodeTime = encodeTime
			zec.EncodeCaller = zapcore.FullCallerEncoder
			consoleEncoder := zapcore.NewConsoleEncoder(zec)
			cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))

		default:
			continue
		}
	}

	// zap.AddCallerSkip(1)，因为多了一层调用：func Debug(...) { Log.Logger.Debug(...) }
	Log.Logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	Log.SugaredLogger = Log.Logger.Sugar()
	defer Log.SugaredLogger.Sync()

	Log.inited = true
}

func GetLevel(lvl string) zapcore.Level {
	levelMappings := map[string]zapcore.Level{
		"debug":  zap.DebugLevel,
		"info":   zap.InfoLevel,
		"warn":   zap.WarnLevel,
		"error":  zap.ErrorLevel,
		"dpanic": zap.DPanicLevel,
		"panic":  zap.PanicLevel,
		"fatal":  zap.FatalLevel,
	}

	if v, ok := levelMappings[strings.ToLower(lvl)]; ok {
		return v
	} else {
		return zap.DebugLevel
	}
}

func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(Log.TimeFormat))
}

func Debug(msg string, fields ...zapcore.Field) {
	Log.Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	Log.Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	Log.Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	Log.Logger.Error(msg, fields...)
}

func DPanic(msg string, fields ...zapcore.Field) {
	Log.Logger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zapcore.Field) {
	Log.Logger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	Log.Logger.Fatal(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	Log.SugaredLogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	Log.SugaredLogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	Log.SugaredLogger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Log.SugaredLogger.Errorf(template, args...)
}

func DPanicf(args ...interface{}) {
	Log.SugaredLogger.DPanic(args...)
}

func Panicf(args ...interface{}) {
	Log.SugaredLogger.Panic(args...)
}

func Fatalf(template string, args ...interface{}) {
	Log.SugaredLogger.Fatalf(template, args...)
}
