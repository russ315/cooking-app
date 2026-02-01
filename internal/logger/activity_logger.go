package logger

import (
	"fmt"
	"time"
)

// Event представляет событие для логирования
type Event struct {
	Action    string
	UserID    int
	Timestamp time.Time
}

// ActivityLogger логирует действия пользователей асинхронно
// Использует goroutine и channels (требование Assignment 4)
type ActivityLogger struct {
	events chan Event
}

// NewActivityLogger создает новый логгер
func NewActivityLogger() *ActivityLogger {
	logger := &ActivityLogger{
		events: make(chan Event, 100), // buffered channel
	}

	// Запускаем goroutine для обработки событий (Assignment 4 requirement)
	go logger.processEvents()

	return logger
}

// Log отправляет событие в channel (не блокирует)
func (l *ActivityLogger) Log(action string, userID int) {
	event := Event{
		Action:    action,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	// Отправляем в channel (асинхронно)
	select {
	case l.events <- event:
		// Событие отправлено
	default:
		// Channel переполнен, пропускаем
		fmt.Println("Warning: Event log full, dropping event")
	}
}

// processEvents обрабатывает события в отдельной goroutine
func (l *ActivityLogger) processEvents() {
	fmt.Println("🚀 Activity logger goroutine started (Assignment 4 concurrency)")

	for event := range l.events {
		// Симулируем асинхронную обработку
		fmt.Printf("[LOG] %s | User ID: %d | Action: %s\n",
			event.Timestamp.Format("15:04:05"),
			event.UserID,
			event.Action,
		)

		// Небольшая задержка для демонстрации async обработки
		time.Sleep(10 * time.Millisecond)
	}
}

func (l *ActivityLogger) Close() {
	close(l.events)
}
