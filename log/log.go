package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
	"time"
)

var (
	sugaredLogger *zap.SugaredLogger
	atomicLevel   zap.AtomicLevel
	once          sync.Once
)

// defaultEncoderConfig 默认的Encoder配置
var defaultEncoderConfig = zapcore.EncoderConfig{
	MessageKey:     "message",
	LevelKey:       "level",
	TimeKey:        "time",
	NameKey:        "logger",
	CallerKey:      "caller",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalLevelEncoder, //zapcore.CapitalColorLevelEncoder,
	EncodeTime:     CustomTimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.FullCallerEncoder,
}

func init() {
	Initialize(defaultEncoderConfig)
}

// Initialize 初始化日志系统，可传入自定义的EncoderConfig
func Initialize(encoderCfg zapcore.EncoderConfig) {
	once.Do(func() {
		// 定义默认级别为debug
		atomicLevel = zap.NewAtomicLevel()
		atomicLevel.SetLevel(zapcore.DebugLevel)

		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			os.Stdout,
			atomicLevel,
		)

		logger := zap.New(core)
		logger = logger.WithOptions(zap.AddCallerSkip(1))
		logger = logger.WithOptions(zap.AddCaller())
		sugaredLogger = logger.Sugar()
	})
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05.000"))
}

func Logger() *zap.Logger {
	return sugaredLogger.Desugar()
}

func SetLevel(level zapcore.Level) {
	atomicLevel.SetLevel(level)
}

func Info(template string, args ...interface{}) {
	sugaredLogger.Infof(template, args...)
}

func Debug(template string, args ...interface{}) {
	sugaredLogger.Debugf(template, args...)
}

func Error(template string, args ...interface{}) {
	sugaredLogger.Errorf(template, args...)
}

func Fatal(template string, args ...interface{}) {
	sugaredLogger.Fatalf(template, args...)
}

func Panic(template string, args ...interface{}) {
	sugaredLogger.Panicf(template, args...)
}

func Warn(template string, args ...interface{}) {
	sugaredLogger.Warnf(template, args...)
}
