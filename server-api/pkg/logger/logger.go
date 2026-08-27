package logger

import (
	"io"
	"math/rand"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	LogLevel       = zap.DebugLevel
	DefaultLogPath = "./logs/"
)

type LogConfig struct {
	LogPath      string // logPath 日志文件路径
	LogLevel     string // logLevel 日志级别 debug/info/warn/error
	MaxSize      int    // maxSize 单个文件大小,MB
	MaxBackups   int    // maxBackups 保存的文件个数
	MaxAge       int    // maxAge 保存的天数， 没有的话不删除
	Compress     bool   // compress 压缩
	JsonFormat   bool   // jsonFormat 是否输出为json格式
	ShowLine     bool   // shoowLine 显示代码行
	LogInConsole bool   // logInConsole 是否同时输出到控制台
}

var logger *zap.Logger

func InitLog(config LogConfig) {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		LineEnding:   zapcore.DefaultLineEnding,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	}
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	// 实现两个判断日志等级的interface (其实 zapcore.*Level 自身就是 interface)
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})
	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getWriter(DefaultLogPath + "info.log")
	warnWriter := getWriter(DefaultLogPath + "error.log")

	cores := []zapcore.Core{
		zapcore.NewCore(fileEncoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(warnWriter), warnLevel),
	}

	//TODO::在配置文件中增加选项,决定是否开启
	if config.LogInConsole {
		consoleEncoder := fileEncoder
		if !config.JsonFormat {
			consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		}
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(zapcore.Lock(os.Stdout)), zap.DebugLevel))
	}

	// 最后创建具体的Logger
	core := zapcore.NewTee(cores...)
	logger = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑

}

func getWriter(filename string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		filename+".%Y%m%d-%H", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*2),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func InfoRand(msg string, fields ...zap.Field) {
	//LogWriter.Info(nil, msg, fields)
	if rand.Int31n(100) == 1 {
		logger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// 慎用
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// 对象化接口
type Logger struct {
}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func (l *Logger) InfoRand(msg string, fields ...zap.Field) {
	InfoRand(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// 慎用
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
