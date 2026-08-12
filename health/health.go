package health

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	Passing Status = "passing"
	Warning Status = "warning"
	Failing Status = "failing"
)

type Result struct {
	Name      string
	Status    Status
	Message   string
	StartedAt time.Time
	Duration  time.Duration
}

type Check interface {
	Name() string
	Run(context.Context) error
}

type CheckFunc struct {
	CheckName string
	Function  func(context.Context) error
}

func (check CheckFunc) Name() string { return check.CheckName }

func (check CheckFunc) Run(ctx context.Context) error {
	if check.Function == nil {
		return fmt.Errorf("health check has no function")
	}
	return check.Function(ctx)
}

type Monitor struct {
	mu      sync.RWMutex
	checks  map[string]Check
	results map[string]Result
	clock   func() time.Time
}

func New() *Monitor {
	return &Monitor{checks: map[string]Check{}, results: map[string]Result{}, clock: time.Now}
}

func (monitor *Monitor) Register(check Check) error {
	if check == nil || strings.TrimSpace(check.Name()) == "" {
		return fmt.Errorf("health check name is empty")
	}
	monitor.mu.Lock()
	defer monitor.mu.Unlock()
	if _, exists := monitor.checks[check.Name()]; exists {
		return fmt.Errorf("health check %q already exists", check.Name())
	}
	monitor.checks[check.Name()] = check
	return nil
}

func (monitor *Monitor) Run(ctx context.Context, name string) (Result, error) {
	monitor.mu.RLock()
	check, exists := monitor.checks[name]
	monitor.mu.RUnlock()
	if !exists {
		return Result{}, fmt.Errorf("health check %q not found", name)
	}
	started := monitor.clock()
	err := check.Run(ctx)
	finished := monitor.clock()
	result := Result{Name: name, Status: Passing, StartedAt: started, Duration: finished.Sub(started)}
	if err != nil {
		result.Status = Failing
		result.Message = err.Error()
	}
	monitor.mu.Lock()
	monitor.results[name] = result
	monitor.mu.Unlock()
	return result, nil
}

func (monitor *Monitor) RunAll(ctx context.Context) []Result {
	monitor.mu.RLock()
	names := make([]string, 0, len(monitor.checks))
	for name := range monitor.checks {
		names = append(names, name)
	}
	monitor.mu.RUnlock()
	sort.Strings(names)
	result := make([]Result, 0, len(names))
	for _, name := range names {
		value, _ := monitor.Run(ctx, name)
		result = append(result, value)
	}
	return result
}

func (monitor *Monitor) Last(name string) (Result, bool) {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()
	result, exists := monitor.results[name]
	return result, exists
}

type Summary struct {
	Passing int
	Warning int
	Failing int
	Total   int
}

func Summarize(results []Result) Summary {
	var summary Summary
	for _, result := range results {
		summary.Total++
		switch result.Status {
		case Passing:
			summary.Passing++
		case Warning:
			summary.Warning++
		default:
			summary.Failing++
		}
	}
	return summary
}

func (summary Summary) Healthy() bool { return summary.Failing == 0 }

func Filter(results []Result, status Status) []Result {
	filtered := make([]Result, 0)
	for _, result := range results {
		if result.Status == status {
			filtered = append(filtered, result)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Name < filtered[j].Name })
	return filtered
}

func FailureMessages(results []Result) []string {
	failures := Filter(results, Failing)
	messages := make([]string, 0, len(failures))
	for _, result := range failures {
		messages = append(messages, fmt.Sprintf("%s: %s", result.Name, result.Message))
	}
	return messages
}

func Count(results []Result, status Status) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}
