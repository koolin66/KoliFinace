package logger

import (
	"log/slog"
	"os"
)

// уровень логгирования (тип данных level с типом данных string)
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// обертка над slog, нужна чтобы в случае замены
// библиотеки не прокидывать ее везде
type Logger struct {
	sl *slog.Logger
}

// создает новй логгер с указанным уровнем
func New(level Level) *Logger {
	var l slog.Level
	switch level {
	case LevelDebug:
		l = slog.LevelDebug
	case LevelInfo:
		l = slog.LevelInfo
	case LevelWarn:
		l = slog.LevelWarn
	case LevelError:
		l = slog.LevelError
	}

	//создание лога который пишет в тепримнал(os.Stdout) в json формате
	sl := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: l,
	}))

	//тут возвращаем этот лог
	return &Logger{sl: sl}
}

//пишем методы для логгера

// отладочные сообщения
func (l *Logger) Debug(msg string, args ...any) {
	l.sl.Debug(msg, args...)
}

// информационные сообщения
func (l *Logger) Info(msg string, args ...any) {
	l.sl.Info(msg, args...)
}

// логирует предупреждение
func (l *Logger) Warn(msg string, args ...any) {
	l.sl.Warn(msg, args...)
}

// логирует ошибку
func (l *Logger) Error(msg string, args ...any) {
	l.sl.Error(msg, args...)
}

//использование
//l.log.Info("transaction added", "id", tx.ID, "amount", tx.Amount)
//l.log.Error("failed to save", "err", err)
