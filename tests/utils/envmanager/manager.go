package envmanager

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager manages test environments
type Manager struct {
	environments map[string]*Environment
	pool         *EnvironmentPool
	mu           sync.RWMutex
	config       *Config
}

// Config holds configuration for the environment manager
type Config struct {
	MaxEnvironments     int
	HealthCheckInterval time.Duration
	CleanupTimeout      time.Duration
	StateBackend        string
	StateBasePath       string
	DockerNetwork       string
	EnablePooling       bool
	PoolSize            int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxEnvironments:     10,
		HealthCheckInterval: 1 * time.Minute,
		CleanupTimeout:      10 * time.Minute,
		StateBackend:        "local",
		StateBasePath:       "./test-state",
		DockerNetwork:       "test-network",
		EnablePooling:       true,
		PoolSize:            5,
	}
}

// NewManager creates a new environment manager
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	m := &Manager{
		environments: make(map[string]*Environment),
		config:       config,
	}

	if config.EnablePooling {
		m.pool = NewEnvironmentPool(config.PoolSize)
	}

	return m
}

// CreateEnvironment creates a new test environment
func (m *Manager) CreateEnvironment(ctx context.Context, provider Provider) (*Environment, error) {
	m.mu.Lock()
	if len(m.environments) >= m.config.MaxEnvironments {
		m.mu.Unlock()
		return nil, fmt.Errorf("maximum number of environments (%d) reached", m.config.MaxEnvironments)
	}
	m.mu.Unlock()

	// Generate unique ID
	id := fmt.Sprintf("%s-%d", provider, time.Now().Unix())

	env := NewEnvironment(id, provider)
	env.StateBackend = m.config.StateBackend
	env.StateFile = fmt.Sprintf("%s/%s/terraform.tfstate", m.config.StateBasePath, id)

	m.mu.Lock()
	m.environments[id] = env
	m.mu.Unlock()

	return env, nil
}

// GetEnvironment retrieves an environment by ID
func (m *Manager) GetEnvironment(id string) (*Environment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	env, ok := m.environments[id]
	if !ok {
		return nil, fmt.Errorf("environment %s not found", id)
	}

	return env, nil
}

// AcquireEnvironment gets an environment from the pool or creates a new one
func (m *Manager) AcquireEnvironment(ctx context.Context, provider Provider) (*Environment, error) {
	if m.pool != nil && m.config.EnablePooling {
		// Try to get from pool first
		if env := m.pool.Acquire(provider); env != nil {
			env.SetStatus(StatusInUse)
			return env, nil
		}
	}

	// Create new environment if pool is empty or pooling is disabled
	env, err := m.CreateEnvironment(ctx, provider)
	if err != nil {
		return nil, err
	}

	env.SetStatus(StatusProvisioning)
	return env, nil
}

// ReleaseEnvironment returns an environment to the pool or destroys it
func (m *Manager) ReleaseEnvironment(ctx context.Context, env *Environment) error {
	if m.pool != nil && m.config.EnablePooling {
		// Reset environment and return to pool
		env.SetStatus(StatusReady)
		return m.pool.Release(env)
	}

	// Destroy environment if pooling is disabled
	return m.DestroyEnvironment(ctx, env.ID)
}

// DestroyEnvironment destroys an environment
func (m *Manager) DestroyEnvironment(ctx context.Context, id string) error {
	m.mu.Lock()
	env, ok := m.environments[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("environment %s not found", id)
	}
	delete(m.environments, id)
	m.mu.Unlock()

	// Run cleanup with timeout
	cleanupCtx, cancel := context.WithTimeout(ctx, m.config.CleanupTimeout)
	defer cancel()

	return env.Cleanup(cleanupCtx)
}

// ListEnvironments returns all environments
func (m *Manager) ListEnvironments() []*Environment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	envs := make([]*Environment, 0, len(m.environments))
	for _, env := range m.environments {
		envs = append(envs, env)
	}

	return envs
}

// GetEnvironmentsByStatus returns environments with a specific status
func (m *Manager) GetEnvironmentsByStatus(status EnvironmentStatus) []*Environment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	envs := make([]*Environment, 0)
	for _, env := range m.environments {
		if env.GetStatus() == status {
			envs = append(envs, env)
		}
	}

	return envs
}

// StartHealthChecks starts periodic health checks for all environments
func (m *Manager) StartHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks checks health of all environments
func (m *Manager) performHealthChecks(ctx context.Context) {
	envs := m.ListEnvironments()

	for _, env := range envs {
		go func(e *Environment) {
			// Perform health check
			healthy := m.checkEnvironmentHealth(ctx, e)

			if healthy {
				e.UpdateHealthCheck("healthy")
			} else {
				e.UpdateHealthCheck("unhealthy")

				// If environment is unhealthy and in use, mark as failed
				if e.GetStatus() == StatusInUse {
					e.SetStatus(StatusFailed)
				}
			}
		}(env)
	}
}

// checkEnvironmentHealth performs actual health check
func (m *Manager) checkEnvironmentHealth(ctx context.Context, env *Environment) bool {
	// Check if environment is in a valid state
	status := env.GetStatus()
	if status == StatusFailed || status == StatusDestroyed {
		return false
	}

	// Check if container is running (if containerized)
	if env.ContainerID != "" {
		// Docker health check would go here
		// For now, assume healthy if status is ready or in_use
		return status == StatusReady || status == StatusInUse
	}

	return true
}

// CleanupAll destroys all environments
func (m *Manager) CleanupAll(ctx context.Context) error {
	envs := m.ListEnvironments()

	var wg sync.WaitGroup
	errors := make(chan error, len(envs))

	for _, env := range envs {
		wg.Add(1)
		go func(e *Environment) {
			defer wg.Done()
			if err := m.DestroyEnvironment(ctx, e.ID); err != nil {
				errors <- fmt.Errorf("failed to destroy environment %s: %w", e.ID, err)
			}
		}(env)
	}

	wg.Wait()
	close(errors)

	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup failed with %d errors: %v", len(errs), errs)
	}

	return nil
}

// GetStats returns statistics about managed environments
func (m *Manager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"total":        len(m.environments),
		"pending":      0,
		"provisioning": 0,
		"ready":        0,
		"in_use":       0,
		"cleaning":     0,
		"failed":       0,
		"destroyed":    0,
	}

	for _, env := range m.environments {
		status := env.GetStatus()
		switch status {
		case StatusPending:
			stats["pending"]++
		case StatusProvisioning:
			stats["provisioning"]++
		case StatusReady:
			stats["ready"]++
		case StatusInUse:
			stats["in_use"]++
		case StatusCleaning:
			stats["cleaning"]++
		case StatusFailed:
			stats["failed"]++
		case StatusDestroyed:
			stats["destroyed"]++
		}
	}

	if m.pool != nil {
		poolStats := m.pool.Stats()
		for k, v := range poolStats {
			stats["pool_"+k] = v
		}
	}

	return stats
}
