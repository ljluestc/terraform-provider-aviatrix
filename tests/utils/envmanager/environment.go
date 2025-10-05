package envmanager

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Provider represents a cloud provider
type Provider string

const (
	ProviderAWS   Provider = "aws"
	ProviderAzure Provider = "azure"
	ProviderGCP   Provider = "gcp"
	ProviderOCI   Provider = "oci"
)

// EnvironmentStatus represents the current state of an environment
type EnvironmentStatus string

const (
	StatusPending     EnvironmentStatus = "pending"
	StatusProvisioning EnvironmentStatus = "provisioning"
	StatusReady       EnvironmentStatus = "ready"
	StatusInUse       EnvironmentStatus = "in_use"
	StatusCleaning    EnvironmentStatus = "cleaning"
	StatusFailed      EnvironmentStatus = "failed"
	StatusDestroyed   EnvironmentStatus = "destroyed"
)

// Environment represents a test environment instance
type Environment struct {
	ID                string            `json:"id"`
	Provider          Provider          `json:"provider"`
	Status            EnvironmentStatus `json:"status"`
	StateFile         string            `json:"state_file"`
	StateBackend      string            `json:"state_backend"`
	ContainerID       string            `json:"container_id,omitempty"`
	NetworkID         string            `json:"network_id,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	LastHealthCheck   time.Time         `json:"last_health_check,omitempty"`
	HealthCheckStatus string            `json:"health_check_status,omitempty"`
	Resources         map[string]string `json:"resources"`
	Metadata          map[string]string `json:"metadata"`
	CleanupHandlers   []CleanupFunc     `json:"-"`
	mu                sync.RWMutex
}

// CleanupFunc is a function that cleans up resources
type CleanupFunc func(ctx context.Context) error

// NewEnvironment creates a new environment instance
func NewEnvironment(id string, provider Provider) *Environment {
	return &Environment{
		ID:              id,
		Provider:        provider,
		Status:          StatusPending,
		Resources:       make(map[string]string),
		Metadata:        make(map[string]string),
		CleanupHandlers: make([]CleanupFunc, 0),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// SetStatus updates the environment status
func (e *Environment) SetStatus(status EnvironmentStatus) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Status = status
	e.UpdatedAt = time.Now()
}

// GetStatus returns the current environment status
func (e *Environment) GetStatus() EnvironmentStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Status
}

// AddResource adds a resource to the environment
func (e *Environment) AddResource(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Resources[key] = value
	e.UpdatedAt = time.Now()
}

// GetResource retrieves a resource from the environment
func (e *Environment) GetResource(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	val, ok := e.Resources[key]
	return val, ok
}

// AddMetadata adds metadata to the environment
func (e *Environment) AddMetadata(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Metadata[key] = value
	e.UpdatedAt = time.Now()
}

// GetMetadata retrieves metadata from the environment
func (e *Environment) GetMetadata(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	val, ok := e.Metadata[key]
	return val, ok
}

// RegisterCleanup adds a cleanup handler to the environment
func (e *Environment) RegisterCleanup(handler CleanupFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.CleanupHandlers = append(e.CleanupHandlers, handler)
}

// Cleanup runs all cleanup handlers for the environment
func (e *Environment) Cleanup(ctx context.Context) error {
	e.SetStatus(StatusCleaning)

	e.mu.RLock()
	handlers := make([]CleanupFunc, len(e.CleanupHandlers))
	copy(handlers, e.CleanupHandlers)
	e.mu.RUnlock()

	var errors []error
	for i := len(handlers) - 1; i >= 0; i-- {
		if err := handlers[i](ctx); err != nil {
			errors = append(errors, fmt.Errorf("cleanup handler %d failed: %w", i, err))
		}
	}

	if len(errors) > 0 {
		e.SetStatus(StatusFailed)
		return fmt.Errorf("cleanup failed with %d errors: %v", len(errors), errors)
	}

	e.SetStatus(StatusDestroyed)
	return nil
}

// IsHealthy checks if the environment is healthy
func (e *Environment) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.Status != StatusReady && e.Status != StatusInUse {
		return false
	}

	// Check if health check is recent (within last 5 minutes)
	if !e.LastHealthCheck.IsZero() {
		if time.Since(e.LastHealthCheck) > 5*time.Minute {
			return false
		}
	}

	return e.HealthCheckStatus == "healthy"
}

// UpdateHealthCheck updates the health check status
func (e *Environment) UpdateHealthCheck(status string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.LastHealthCheck = time.Now()
	e.HealthCheckStatus = status
	e.UpdatedAt = time.Now()
}
