package log

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger *zap.SugaredLogger

func InitLogger() {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	lumberJackLogger := &lumberjack.Logger{
		Filename:   conf.Conf.Log.InfoPath, // 日志文件的位置
		MaxSize:    100,                    // 在进行切割之前，日志文件的最大大小
		MaxBackups: 5,                      // 保留旧文件的最大个数
		Compress:   false,                  // 是否压缩/归档旧文件
	}
	consoleSyncer := zapcore.AddSync(os.Stdout)
	writeSync := zapcore.AddSync(lumberJackLogger)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, writeSync, zap.DebugLevel),
		zapcore.NewCore(encoder, consoleSyncer, zap.DebugLevel),
	)
	Logger = zap.New(core, zap.AddCaller()).Sugar()
	fmt.Println("logger initialized")
}
