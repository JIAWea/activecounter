package log

import (
	"io"
	"os"
	"strings"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger logger struct
type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

// newZapLogger new zap logger
func newZapLogger(cfg *Config) (Logger, error) {
	encoder := getJSONEncoder()

	var cores []zapcore.Core
	writers := strings.Split(cfg.Writers, ",")
	for _, w := range writers {
		if w == "stdout" {
			core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
			cores = append(cores, core)
		}

		if w == "file" {
			// backCount := uint(cfg.LogBackupCount)
			backCount := uint(24)
			// 注意：如果多个文件，最后一个会是全的，前两个可能会丢日志
			infoFilename := cfg.LoggerFile
			infoWrite := getLogWriterWithTime(infoFilename, backCount)
			warnFilename := cfg.LoggerWarnFile
			warnWrite := getLogWriterWithTime(warnFilename, backCount)
			errorFilename := cfg.LoggerErrorFile
			errorWrite := getLogWriterWithTime(errorFilename, backCount)

			infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl <= zapcore.InfoLevel
			})
			warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl == zapcore.WarnLevel
			})
			errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			})

			core := zapcore.NewCore(encoder, zapcore.AddSync(infoWrite), infoLevel)
			cores = append(cores, core)
			core = zapcore.NewCore(encoder, zapcore.AddSync(warnWrite), warnLevel)
			cores = append(cores, core)
			core = zapcore.NewCore(encoder, zapcore.AddSync(errorWrite), errorLevel)
			cores = append(cores, core)
		}
	}

	combinedCore := zapcore.NewTee(cores...)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段
	filed := zap.Fields(zap.String("ip", "127.0.0.1"))
	// 构造日志
	logger := zap.New(
		combinedCore,
		zap.AddCallerSkip(2), // 跳过文件调用层数
		caller,
		development,
		filed,
	).Sugar()

	return &zapLogger{
		sugaredLogger: logger,
	}, nil
}

// getJSONEncoder
func getJSONEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		NameKey:        "app",
		CallerKey:      "file",
		StacktraceKey:  "stacktrace",
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriterWithTime 按时间(小时)进行切割
func getLogWriterWithTime(filename string, backCount uint) io.Writer {
	logFullPath := strings.Replace(filename, ".log", "", -1)
	hook, err := rotatelogs.New(
		logFullPath + ".%Y-%m.log", // 时间格式使用shell的date时间格式
		// logFullPath + ".%Y-%m-%d-%H.log", // 时间格式使用shell的date时间格式
		// rotatelogs.WithLinkName(logFullPath),      // 生成软链，指向最新日志文件
		// rotatelogs.WithRotationCount(backCount),   // 文件最大保存份数
		// rotatelogs.WithRotationTime(time.Hour*24), // 日志切割时间间隔
	)

	if err != nil {
		panic(err)
	}
	return hook
}

// Debug logger
func (l *zapLogger) Debug(args ...interface{}) {
	l.sugaredLogger.Debug(args...)
}

// Info logger
func (l *zapLogger) Info(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

// Warn logger
func (l *zapLogger) Warn(args ...interface{}) {
	l.sugaredLogger.Warn(args...)
}

// Error logger
func (l *zapLogger) Error(args ...interface{}) {
	l.sugaredLogger.Error(args...)
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.sugaredLogger.Fatal(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) WithFields(fields Fields) Logger {
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	newLogger := l.sugaredLogger.With(f...)
	return &zapLogger{newLogger}
}
