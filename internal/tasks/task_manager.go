package tasks

import (
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Task struct {
	id       string
	function interface{}
	args     []interface{}
}

type BackgroundTaskManager struct {
	logger *zap.Logger
	tasks  chan Task
	wg     sync.WaitGroup
}

func NewBackgroundTaskManager(logger *zap.Logger) *BackgroundTaskManager {
	btm := &BackgroundTaskManager{
		logger: logger,
		tasks:  make(chan Task, 100),
	}

	btm.wg.Add(1)
	go btm.processTasks()

	return btm
}

func (btm *BackgroundTaskManager) processTasks() {
	defer btm.wg.Done()

	for task := range btm.tasks {
		go func(t Task) {
			defer func() {
				if r := recover(); r != nil {
					btm.logger.Fatal("🔥 Critical panic in background task",
						zap.Any("panic_value", r),
						zap.String("task_id", t.id),
						zap.Stack("stacktrace"),
						zap.String("recovery_advice", `
                            1. Check task input parameters
                            2. Validate async function signatures
                            3. Add panic recovery middleware`),
					)
				}
			}()

			v := reflect.ValueOf(t.function)
			if v.Kind() != reflect.Func {
				btm.logger.Fatal("❌ Invalid task function type",
					zap.String("expected", "reflect.Func"),
					zap.String("received", v.Kind().String()),
					zap.String("task_id", t.id),
					zap.Any("task_details", map[string]interface{}{
						"args": t.args,
					}),
					zap.String("solution", "Use btm.AddTask() with valid function reference"),
				)
				return
			}

			args := make([]reflect.Value, len(t.args))
			for i, arg := range t.args {
				args[i] = reflect.ValueOf(arg)
			}

			btm.logger.Debug("Executing background task",
				zap.String("task_id", t.id),
				zap.Int("num_args", len(args)),
			)

			start := time.Now()
			results := v.Call(args)
			latency := time.Since(start)

			btm.logger.Info("Background task completed",
				zap.String("task_id", t.id),
				zap.Duration("latency", latency),
				zap.Int("results_count", len(results)),
			)
		}(task)
	}
}

func (btm *BackgroundTaskManager) AddTask(function interface{}, args ...interface{}) {
	btm.tasks <- Task{
		id:       uuid.New().String(),
		function: function,
		args:     args,
	}
}

func (btm *BackgroundTaskManager) Close() {
	close(btm.tasks)
	btm.wg.Wait()
}
