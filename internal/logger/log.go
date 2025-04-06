package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
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

	requestToken string
	routeToken   int64
}

var Log = &Logger{}

func SetToken(requestToken string, routeToken int64) {
	Log.requestToken = requestToken
	Log.routeToken = routeToken
}

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

type AppendEncoder struct {
	zapcore.Encoder
	pool buffer.Pool
}

func (e *AppendEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {

	buf := e.pool.Get()

	buf.AppendString(Log.requestToken)
	buf.AppendByte(':')
	buf.AppendInt(Log.routeToken)
	buf.AppendString("\t")

	// calling the embedded encoder's EncodeEntry to keep the original encoding format
	consolebuf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// just write the output into your own buffer
	_, err = buf.Write(consolebuf.Bytes())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func InitLogger(cfg *ini.File, logdir string, timeFormat string) {
	Log.Lock()
	defer Log.Unlock()

	if Log.inited {
		return
	}

	Log.TimeFormat = timeFormat
	Log.RootDir = logdir

	SetToken(fmt.Sprint(time.Now().UnixMilli()), time.Now().UnixMilli())

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
				MaxAge:     sec.Key("MAX_AGE").MustInt(7), // days
				Compress:   false,
			}

			zec := zap.NewProductionEncoderConfig()
			zec.EncodeTime = encodeTime
			zec.EncodeCaller = zapcore.FullCallerEncoder
			fileEcoder := &AppendEncoder{
				Encoder: zapcore.NewJSONEncoder(zec),
				pool:    buffer.NewPool(),
			}

			cores = append(cores, zapcore.NewCore(fileEcoder, zapcore.AddSync(lumberJackLogger), level))

		case "console":
			zec := zap.NewDevelopmentEncoderConfig()
			zec.EncodeTime = encodeTime
			zec.EncodeLevel = zapcore.CapitalColorLevelEncoder
			zec.EncodeCaller = zapcore.FullCallerEncoder
			consoleEncoder := &AppendEncoder{
				Encoder: zapcore.NewConsoleEncoder(zec),
				pool:    buffer.NewPool(),
			}
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

func Debugw(template string, keysAndValues ...interface{}) {
	Log.SugaredLogger.Debugw(template, keysAndValues...)
}

func Infow(template string, keysAndValues ...interface{}) {
	Log.SugaredLogger.Infow(template, keysAndValues...)
}

func Warnw(template string, keysAndValues ...interface{}) {
	Log.SugaredLogger.Warnw(template, keysAndValues...)
}

func Errorw(template string, keysAndValues ...interface{}) {
	Log.SugaredLogger.Errorw(template, keysAndValues...)
}

func DPanicw(keysAndValues ...interface{}) {
	Log.SugaredLogger.DPanic(keysAndValues...)
}

func Panicw(keysAndValues ...interface{}) {
	Log.SugaredLogger.Panic(keysAndValues...)
}

func Fatalw(template string, keysAndValues ...interface{}) {
	Log.SugaredLogger.Fatalw(template, keysAndValues...)
}
