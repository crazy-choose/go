package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var (
	sugaredLogger *zap.SugaredLogger
	atomicLevel   zap.AtomicLevel
	ioWrite       *lumberjack.Logger
)

func init() {
	Initialize(false)
}

// Initialize 默认控制台输出, 如果需要文件输出, 请调用 Initialize(true)
func Initialize(ioFile bool) {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:    "message",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "logger",
		CallerKey:     "caller",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		//EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// define default level as debug level
	atomicLevel = zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), os.Stdout, atomicLevel)
	if ioFile {
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), getLogWriter(), atomicLevel)
	}
	logger := zap.New(core)
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	logger = logger.WithOptions(zap.AddCaller())
	sugaredLogger = logger.Sugar()
}

func Logger() *zap.Logger {
	return sugaredLogger.Desugar()
}

func SetLevel(level zapcore.Level) {
	atomicLevel.SetLevel(level)
}

func Fatal(template string, args ...interface{}) {
	sugaredLogger.Fatalf(template, args...)
}

func Error(template string, args ...interface{}) {
	sugaredLogger.Errorf(template, args...)
}

func Panic(template string, args ...interface{}) {
	sugaredLogger.Panicf(template, args...)
}

func Warn(template string, args ...interface{}) {
	sugaredLogger.Warnf(template, args...)
}

func Info(template string, args ...interface{}) {
	sugaredLogger.Infof(template, args...)
}

func Debug(template string, args ...interface{}) {
	sugaredLogger.Debugf(template, args...)
}

func Rotate() {
	fmt.Println("logger.Rotate....")
	ioWrite.Rotate()
}

// 自定义日志输出时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05.000000"))
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   defLogName(),
		MaxSize:    1024,  //MB
		MaxBackups: 0,     //文件数量 -0 无限制
		MaxAge:     0,     //保留时间 -0 永久保存
		Compress:   false, //压缩
	}
	ioWrite = lumberJackLogger
	return zapcore.AddSync(ioWrite)
}

