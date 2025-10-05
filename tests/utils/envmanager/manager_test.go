package envmanager

import (
	"context"
	"testing"
	"time"
)

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	if env.ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", env.ID)
	}

	if env.Provider != ProviderAWS {
		t.Errorf("Expected provider AWS, got %s", env.Provider)
	}

	if env.Status != StatusPending {
		t.Errorf("Expected status pending, got %s", env.Status)
	}

	if env.Resources == nil {
		t.Error("Resources map should be initialized")
	}

	if env.Metadata == nil {
		t.Error("Metadata map should be initialized")
	}
}

func TestEnvironmentSetGetStatus(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	env.SetStatus(StatusReady)

	if status := env.GetStatus(); status != StatusReady {
		t.Errorf("Expected status ready, got %s", status)
	}
}

func TestEnvironmentAddGetResource(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	env.AddResource("vpc_id", "vpc-12345")

	if value, ok := env.GetResource("vpc_id"); !ok {
		t.Error("Resource should exist")
	} else if value != "vpc-12345" {
		t.Errorf("Expected 'vpc-12345', got '%s'", value)
	}

	if _, ok := env.GetResource("nonexistent"); ok {
		t.Error("Nonexistent resource should not exist")
	}
}

func TestEnvironmentAddGetMetadata(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	env.AddMetadata("region", "us-east-1")

	if value, ok := env.GetMetadata("region"); !ok {
		t.Error("Metadata should exist")
	} else if value != "us-east-1" {
		t.Errorf("Expected 'us-east-1', got '%s'", value)
	}
}

func TestEnvironmentRegisterCleanup(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	called := false
	cleanup := func(ctx context.Context) error {
		called = true
		return nil
	}

	env.RegisterCleanup(cleanup)

	if len(env.CleanupHandlers) != 1 {
		t.Errorf("Expected 1 cleanup handler, got %d", len(env.CleanupHandlers))
	}
}

func TestEnvironmentCleanup(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	call1 := false
	call2 := false

	env.RegisterCleanup(func(ctx context.Context) error {
		call1 = true
		return nil
	})

	env.RegisterCleanup(func(ctx context.Context) error {
		call2 = true
		return nil
	})

	ctx := context.Background()
	if err := env.Cleanup(ctx); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	if !call1 || !call2 {
		t.Error("All cleanup handlers should be called")
	}

	if env.GetStatus() != StatusDestroyed {
		t.Errorf("Expected status destroyed, got %s", env.GetStatus())
	}
}

func TestNewManager(t *testing.T) {
	config := DefaultConfig()
	manager := NewManager(config)

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if manager.environments == nil {
		t.Error("Environments map should be initialized")
	}

	if config.EnablePooling && manager.pool == nil {
		t.Error("Pool should be initialized when pooling is enabled")
	}
}

func TestManagerCreateEnvironment(t *testing.T) {
	config := DefaultConfig()
	config.EnablePooling = false
	manager := NewManager(config)

	ctx := context.Background()
	env, err := manager.CreateEnvironment(ctx, ProviderAWS)

	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	if env == nil {
		t.Fatal("Environment should not be nil")
	}

	if env.Provider != ProviderAWS {
		t.Errorf("Expected provider AWS, got %s", env.Provider)
	}

	// Verify environment is stored in manager
	retrievedEnv, err := manager.GetEnvironment(env.ID)
	if err != nil {
		t.Errorf("Failed to retrieve environment: %v", err)
	}

	if retrievedEnv.ID != env.ID {
		t.Errorf("Retrieved environment ID mismatch")
	}
}

func TestManagerMaxEnvironments(t *testing.T) {
	config := DefaultConfig()
	config.MaxEnvironments = 2
	config.EnablePooling = false
	manager := NewManager(config)

	ctx := context.Background()

	// Create first environment
	_, err := manager.CreateEnvironment(ctx, ProviderAWS)
	if err != nil {
		t.Fatalf("Failed to create first environment: %v", err)
	}

	// Create second environment
	_, err = manager.CreateEnvironment(ctx, ProviderAWS)
	if err != nil {
		t.Fatalf("Failed to create second environment: %v", err)
	}

	// Third environment should fail
	_, err = manager.CreateEnvironment(ctx, ProviderAWS)
	if err == nil {
		t.Error("Expected error when exceeding max environments")
	}
}

func TestManagerGetEnvironmentsByStatus(t *testing.T) {
	config := DefaultConfig()
	config.EnablePooling = false
	manager := NewManager(config)

	ctx := context.Background()

	env1, _ := manager.CreateEnvironment(ctx, ProviderAWS)
	env1.SetStatus(StatusReady)

	env2, _ := manager.CreateEnvironment(ctx, ProviderAzure)
	env2.SetStatus(StatusReady)

	env3, _ := manager.CreateEnvironment(ctx, ProviderGCP)
	env3.SetStatus(StatusInUse)

	readyEnvs := manager.GetEnvironmentsByStatus(StatusReady)
	if len(readyEnvs) != 2 {
		t.Errorf("Expected 2 ready environments, got %d", len(readyEnvs))
	}

	inUseEnvs := manager.GetEnvironmentsByStatus(StatusInUse)
	if len(inUseEnvs) != 1 {
		t.Errorf("Expected 1 in-use environment, got %d", len(inUseEnvs))
	}
}

func TestManagerGetStats(t *testing.T) {
	config := DefaultConfig()
	config.EnablePooling = false
	manager := NewManager(config)

	ctx := context.Background()

	env1, _ := manager.CreateEnvironment(ctx, ProviderAWS)
	env1.SetStatus(StatusReady)

	env2, _ := manager.CreateEnvironment(ctx, ProviderAzure)
	env2.SetStatus(StatusInUse)

	stats := manager.GetStats()

	if stats["total"] != 2 {
		t.Errorf("Expected total 2, got %d", stats["total"])
	}

	if stats["ready"] != 1 {
		t.Errorf("Expected ready 1, got %d", stats["ready"])
	}

	if stats["in_use"] != 1 {
		t.Errorf("Expected in_use 1, got %d", stats["in_use"])
	}
}

func TestNewEnvironmentPool(t *testing.T) {
	pool := NewEnvironmentPool(5)

	if pool == nil {
		t.Fatal("Pool should not be nil")
	}

	if pool.pools == nil {
		t.Error("Pools map should be initialized")
	}

	if pool.size != 5 {
		t.Errorf("Expected size 5, got %d", pool.size)
	}
}

func TestPoolAcquireRelease(t *testing.T) {
	pool := NewEnvironmentPool(5)

	env := NewEnvironment("test-1", ProviderAWS)
	env.SetStatus(StatusReady)

	// Release environment to pool
	if err := pool.Release(env); err != nil {
		t.Fatalf("Failed to release environment: %v", err)
	}

	// Acquire environment from pool
	acquired := pool.Acquire(ProviderAWS)
	if acquired == nil {
		t.Fatal("Should acquire environment from pool")
	}

	if acquired.ID != env.ID {
		t.Errorf("Acquired wrong environment")
	}

	// Pool should be empty now
	if acquired2 := pool.Acquire(ProviderAWS); acquired2 != nil {
		t.Error("Pool should be empty")
	}
}

func TestPoolStats(t *testing.T) {
	pool := NewEnvironmentPool(5)

	env1 := NewEnvironment("test-1", ProviderAWS)
	env2 := NewEnvironment("test-2", ProviderAWS)

	pool.Release(env1)
	pool.Release(env2)

	stats := pool.Stats()

	if stats["aws_available"] != 2 {
		t.Errorf("Expected aws_available 2, got %d", stats["aws_available"])
	}

	if stats["total_available"] != 2 {
		t.Errorf("Expected total_available 2, got %d", stats["total_available"])
	}
}

func TestEnvironmentIsHealthy(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	// Initially not healthy (pending status)
	if env.IsHealthy() {
		t.Error("Pending environment should not be healthy")
	}

	// Set to ready but no health check
	env.SetStatus(StatusReady)
	if env.IsHealthy() {
		t.Error("Environment without health check should not be healthy")
	}

	// Set health check status
	env.UpdateHealthCheck("healthy")
	if !env.IsHealthy() {
		t.Error("Environment with recent healthy check should be healthy")
	}

	// Old health check
	env.LastHealthCheck = time.Now().Add(-10 * time.Minute)
	if env.IsHealthy() {
		t.Error("Environment with old health check should not be healthy")
	}
}

func TestEnvironmentUpdateHealthCheck(t *testing.T) {
	env := NewEnvironment("test-1", ProviderAWS)

	env.UpdateHealthCheck("healthy")

	if env.HealthCheckStatus != "healthy" {
		t.Errorf("Expected health check status 'healthy', got '%s'", env.HealthCheckStatus)
	}

	if env.LastHealthCheck.IsZero() {
		t.Error("Last health check time should be set")
	}
}
