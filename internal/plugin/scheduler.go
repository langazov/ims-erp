package plugin

import (
	"context"
	"fmt"
	"time"
)

// Scheduler manages scheduled tasks for plugins
type Scheduler struct {
	tasks   map[string]ScheduledTask
	tickers map[string]*time.Ticker
	enabled map[string]bool
	stopCh  chan struct{}
}

// NewScheduler creates a new task scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks:   make(map[string]ScheduledTask),
		tickers: make(map[string]*time.Ticker),
		enabled: make(map[string]bool),
		stopCh:  make(chan struct{}),
	}
}

// Register registers a scheduled task
func (s *Scheduler) Register(task ScheduledTask) error {
	id := generatePluginID(task.Name())

	s.tasks[id] = task
	s.enabled[id] = true

	// Parse schedule (simplified - expects duration string like "1h", "30m")
	schedule := task.Schedule()
	if schedule == "" {
		return fmt.Errorf("task %s has no schedule", task.Name())
	}

	// Parse duration
	duration, err := time.ParseDuration(schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule format for task %s: %w", task.Name(), err)
	}

	// Create ticker
	ticker := time.NewTicker(duration)
	s.tickers[id] = ticker

	// Start goroutine for this task
	go func(taskID string, t ScheduledTask) {
		for {
			select {
			case <-ticker.C:
				if !s.enabled[taskID] {
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if err := t.Run(ctx); err != nil {
					// Log error
				}
				cancel()

			case <-s.stopCh:
				return
			}
		}
	}(id, task)

	return nil
}

// Unregister removes a scheduled task
func (s *Scheduler) Unregister(taskID string) {
	if ticker, exists := s.tickers[taskID]; exists {
		ticker.Stop()
		delete(s.tickers, taskID)
	}
	delete(s.tasks, taskID)
	delete(s.enabled, taskID)
}

// Enable enables a scheduled task
func (s *Scheduler) Enable(taskID string) {
	s.enabled[taskID] = true
}

// Disable disables a scheduled task
func (s *Scheduler) Disable(taskID string) {
	s.enabled[taskID] = false
}

// Start starts the scheduler (no-op in this implementation)
func (s *Scheduler) Start() {
	// Tasks start automatically when registered
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
	for _, ticker := range s.tickers {
		ticker.Stop()
	}
}

// GetTasks returns all registered tasks
func (s *Scheduler) GetTasks() []ScheduledTask {
	tasks := make([]ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
