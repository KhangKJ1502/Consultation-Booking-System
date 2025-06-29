package initialize

import (
	"cbs_backend/global"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.TimeKey = "timestamp"

	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		panic("Can't create logs directory: " + err.Error())
	}

	logFile, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Can't open log file: " + err.Error())
	}

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(logFile), zapcore.InfoLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
	)

	global.Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
