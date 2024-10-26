package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ILogger определяет набор методов ведения журнала с различными уровнями серьезности.
type ILogger interface {
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Error(args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warn(args ...interface{})
	Fatal(args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Debug(args ...interface{})
	Errorf(template string, args ...interface{})
}

// Log глобальный логер приложения
// Log по умолчанию пустой логер
// Log по рекомендации документации для большинства приложений можно использовать обогащённый логер, поэтому сейчас используется он, если понадобится, заменить на стандартный логер
var Log ILogger = zap.NewNop().Sugar()

// New creates a new logger with the specified log level.
func New(level string) (*zap.SugaredLogger, error) {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// создаём новую конфигурацию логера
	cnf := zap.NewProductionConfig()
	// устанавливаем уровень
	cnf.Level = lvl
	// устанавливаем отображение
	cnf.Encoding = "console"
	// Устанавливаем удобочитаемый формат времени
	cnf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	// создаём логер
	logger, err := cnf.Build()
	if err != nil {
		return nil, err
	}
	// Создаём обогащённый логер и возвращаем
	return logger.Sugar(), nil
}
