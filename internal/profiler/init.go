package profiler

// ProfilerConfig holds configuration for the profiler components
type ProfilerConfig struct {
	// Logging configuration
	LogDirectory        string
	EnableConsoleLogging bool
	MinLogLevel         LogLevel

	// Query profiling configuration
	SlowQueryThresholdMS float64
	MaxMetricsPerQuery  int

	// Slow query logging configuration
	PerformanceDegradationFactor float64
	WarningThresholdMS           float64
}

// DefaultConfig returns a default profiler configuration
func DefaultConfig() ProfilerConfig {
	return ProfilerConfig{
		LogDirectory:                 "./logs",
		EnableConsoleLogging:         false,
		MinLogLevel:                  INFO,
		SlowQueryThresholdMS:         500.0,
		MaxMetricsPerQuery:           100,
		PerformanceDegradationFactor: 1.2,
		WarningThresholdMS:           750.0,
	}
}

// ProfilerComponents holds all profiler instances
type ProfilerComponents struct {
	Logger         *Logger
	QueryProfiler  *QueryProfiler
	SlowQueryLog   *SlowQueryLogger
	QueryOptimizer *QueryOptimizer
}

// InitializeProfiler creates and initializes all profiler components
// This is the main entry point for setting up profiling infrastructure
func InitializeProfiler(config ProfilerConfig) (*ProfilerComponents, error) {
	// Initialize logger
	logger, err := NewLogger(config.LogDirectory, config.EnableConsoleLogging)
	if err != nil {
		return nil, err
	}

	logger.SetMinLevel(config.MinLogLevel)

	// Initialize query profiler
	queryProfiler := NewQueryProfiler(logger, config.SlowQueryThresholdMS)

	// Initialize slow query logger
	slowQueryLog, err := NewSlowQueryLogger(config.LogDirectory, config.SlowQueryThresholdMS)
	if err != nil {
		return nil, err
	}

	// Initialize query optimizer
	queryOptimizer := NewQueryOptimizer()

	components := &ProfilerComponents{
		Logger:         logger,
		QueryProfiler:  queryProfiler,
		SlowQueryLog:   slowQueryLog,
		QueryOptimizer: queryOptimizer,
	}

	logger.Info("Profiler components initialized", map[string]interface{}{
		"slow_query_threshold_ms": config.SlowQueryThresholdMS,
		"log_directory":           config.LogDirectory,
		"console_logging":         config.EnableConsoleLogging,
	})

	return components, nil
}

// Cleanup closes all profiler resources
func (p *ProfilerComponents) Cleanup() error {
	if p.Logger != nil {
		p.Logger.Info("Shutting down profiler components", nil)
		if err := p.Logger.Close(); err != nil {
			return err
		}
	}

	if p.SlowQueryLog != nil {
		if err := p.SlowQueryLog.Close(); err != nil {
			return err
		}
	}

	return nil
}
