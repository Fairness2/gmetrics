package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log глобальный логер приложения
// Log по умолчанию пустой логер
// Log по рекомендации документации для большинства приложений можно использовать обогащённый логер, поэтому сейчас используется он, если понадобится, заменить на стандартный логер
var Log *zap.SugaredLogger = zap.NewNop().Sugar()

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
