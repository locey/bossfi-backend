package logger

import (
	"os"

	"bossfi-blockchain-backend/pkg/config"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 包装zap.Logger，提供更友好的接口
type Logger struct {
	*zap.Logger
}

var GlobalLogger *Logger
var Sugar *zap.SugaredLogger

// New 创建新的Logger实例
func New(cfg *config.Config) (*Logger, error) {
	zapLogger, err := createZapLogger(&cfg.Logger)
	if err != nil {
		return nil, err
	}

	logger := &Logger{Logger: zapLogger}
	GlobalLogger = logger
	Sugar = zapLogger.Sugar()

	return logger, nil
}

func createZapLogger(cfg *config.LoggerConfig) (*zap.Logger, error) {
	// 确保日志目录存在
	if err := os.MkdirAll("./logs", 0755); err != nil {
		return nil, err
	}

	// 设置日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 设置日志轮转
	writer := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台编码器配置（更易读）
	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	var cores []zapcore.Core

	// 文件输出核心（JSON格式）
	cores = append(cores, zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		level,
	))

	// 控制台输出核心（开发环境）
	serverConfig := config.Get()
	if serverConfig != nil && serverConfig.Server.Mode == "debug" {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// 错误日志单独文件
	errorWriter := &lumberjack.Logger{
		Filename:   "./logs/error.log",
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}

	cores = append(cores, zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(errorWriter),
		zapcore.ErrorLevel, // 只记录错误级别以上的日志
	))

	// 创建核心
	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// 记录初始化日志
	logger.Sugar().Info("Logger initialized successfully",
		"level", cfg.Level,
		"filename", cfg.Filename,
	)

	return logger, nil
}

// InitLogger 保持向后兼容
func InitLogger(cfg *config.LoggerConfig) error {
	logger, err := createZapLogger(cfg)
	if err != nil {
		return err
	}

	GlobalLogger = &Logger{Logger: logger}
	Sugar = logger.Sugar()
	return nil
}

func GetLogger() *zap.Logger {
	return GlobalLogger.Logger
}

func GetSugar() *zap.SugaredLogger {
	return Sugar
}

func Sync() error {
	if GlobalLogger != nil {
		return GlobalLogger.Logger.Sync()
	}
	return nil
}

// LogInfo 便捷的信息日志函数
func LogInfo(message string, fields ...zap.Field) {
	if GlobalLogger != nil {
		GlobalLogger.Logger.Info(message, fields...)
	}
}

// LogError 便捷的错误日志函数
func LogError(message string, fields ...zap.Field) {
	if GlobalLogger != nil {
		GlobalLogger.Logger.Error(message, fields...)
	}
}

// LogWarn 便捷的警告日志函数
func LogWarn(message string, fields ...zap.Field) {
	if GlobalLogger != nil {
		GlobalLogger.Logger.Warn(message, fields...)
	}
}

// LogDebug 便捷的调试日志函数
func LogDebug(message string, fields ...zap.Field) {
	if GlobalLogger != nil {
		GlobalLogger.Logger.Debug(message, fields...)
	}
}
