package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func InitLogger() *zap.SugaredLogger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder 

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),         
		zapcore.InfoLevel,                      
	)

	logger := zap.New(core)
	return logger.Sugar()
}