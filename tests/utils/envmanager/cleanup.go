package envmanager

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CleanupManager handles automated resource cleanup
type CleanupManager struct {
	manager         *Manager
	terraformMgr    *TerraformManager
	dockerMgr       *DockerManager
	cleanupInterval time.Duration
	maxAge          time.Duration
	mu              sync.Mutex
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(manager *Manager, terraformMgr *TerraformManager, dockerMgr *DockerManager) *CleanupManager {
	return &CleanupManager{
		manager:         manager,
		terraformMgr:    terraformMgr,
		dockerMgr:       dockerMgr,
		cleanupInterval: 5 * time.Minute,
		maxAge:          2 * time.Hour,
	}
}

// StartAutomaticCleanup starts periodic automatic cleanup
func (cm *CleanupManager) StartAutomaticCleanup(ctx context.Context) {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()

	log.Println("[CleanupManager] Starting automatic cleanup service")

	for {
		select {
		case <-ctx.Done():
			log.Println("[CleanupManager] Stopping automatic cleanup service")
			return
		case <-ticker.C:
			if err := cm.performCleanup(ctx); err != nil {
				log.Printf("[CleanupManager] Cleanup cycle failed: %v", err)
			}
		}
	}
}

// performCleanup performs a cleanup cycle
func (cm *CleanupManager) performCleanup(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	log.Println("[CleanupManager] Starting cleanup cycle")

	// Clean up old environments
	if err := cm.cleanupOldEnvironments(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to cleanup old environments: %v", err)
	}

	// Clean up failed environments
	if err := cm.cleanupFailedEnvironments(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to cleanup failed environments: %v", err)
	}

	// Clean up orphaned containers
	if err := cm.cleanupOrphanedContainers(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to cleanup orphaned containers: %v", err)
	}

	// Clean up orphaned state files
	if err := cm.cleanupOrphanedStateFiles(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to cleanup orphaned state files: %v", err)
	}

	// Prune Docker resources
	if err := cm.dockerMgr.PruneContainers(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to prune containers: %v", err)
	}

	log.Println("[CleanupManager] Cleanup cycle completed")
	return nil
}

// cleanupOldEnvironments cleans up environments older than maxAge
func (cm *CleanupManager) cleanupOldEnvironments(ctx context.Context) error {
	envs := cm.manager.ListEnvironments()

	var wg sync.WaitGroup
	errors := make(chan error, len(envs))

	for _, env := range envs {
		if time.Since(env.CreatedAt) > cm.maxAge {
			log.Printf("[CleanupManager] Environment %s is older than max age, cleaning up", env.ID)

			wg.Add(1)
			go func(e *Environment) {
				defer wg.Done()

				if err := cm.cleanupEnvironment(ctx, e); err != nil {
					errors <- fmt.Errorf("failed to cleanup old environment %s: %w", e.ID, err)
				}
			}(env)
		}
	}

	wg.Wait()
	close(errors)

	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup old environments failed with %d errors: %v", len(errs), errs)
	}

	return nil
}

// cleanupFailedEnvironments cleans up environments in failed state
func (cm *CleanupManager) cleanupFailedEnvironments(ctx context.Context) error {
	failedEnvs := cm.manager.GetEnvironmentsByStatus(StatusFailed)

	var wg sync.WaitGroup
	errors := make(chan error, len(failedEnvs))

	for _, env := range failedEnvs {
		log.Printf("[CleanupManager] Cleaning up failed environment %s", env.ID)

		wg.Add(1)
		go func(e *Environment) {
			defer wg.Done()

			if err := cm.cleanupEnvironment(ctx, e); err != nil {
				errors <- fmt.Errorf("failed to cleanup failed environment %s: %w", e.ID, err)
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
		return fmt.Errorf("cleanup failed environments failed with %d errors: %v", len(errs), errs)
	}

	return nil
}

// cleanupEnvironment cleans up a single environment
func (cm *CleanupManager) cleanupEnvironment(ctx context.Context, env *Environment) error {
	// Destroy Terraform resources
	if env.StateFile != "" {
		if err := cm.terraformMgr.Destroy(ctx, env); err != nil {
			log.Printf("[CleanupManager] Failed to destroy Terraform resources for %s: %v", env.ID, err)
		}
	}

	// Remove Docker container
	if env.ContainerID != "" {
		if err := cm.dockerMgr.RemoveContainer(ctx, env.ContainerID); err != nil {
			log.Printf("[CleanupManager] Failed to remove container for %s: %v", env.ID, err)
		}
	}

	// Run environment cleanup handlers
	if err := env.Cleanup(ctx); err != nil {
		return fmt.Errorf("environment cleanup failed: %w", err)
	}

	// Remove from manager
	return cm.manager.DestroyEnvironment(ctx, env.ID)
}

// cleanupOrphanedContainers removes containers not managed by any environment
func (cm *CleanupManager) cleanupOrphanedContainers(ctx context.Context) error {
	// Get all containers in test network
	containers, err := cm.dockerMgr.ListContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Get all managed container IDs
	envs := cm.manager.ListEnvironments()
	managedContainers := make(map[string]bool)
	for _, env := range envs {
		if env.ContainerID != "" {
			managedContainers[env.ContainerID] = true
		}
	}

	// Remove orphaned containers
	for _, containerID := range containers {
		if !managedContainers[containerID] {
			log.Printf("[CleanupManager] Removing orphaned container %s", containerID)
			if err := cm.dockerMgr.RemoveContainer(ctx, containerID); err != nil {
				log.Printf("[CleanupManager] Failed to remove orphaned container %s: %v", containerID, err)
			}
		}
	}

	return nil
}

// cleanupOrphanedStateFiles removes state directories not managed by any environment
func (cm *CleanupManager) cleanupOrphanedStateFiles(ctx context.Context) error {
	stateDir := cm.terraformMgr.stateDir

	// Get all state directories
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read state directory: %w", err)
	}

	// Get all managed environment IDs
	envs := cm.manager.ListEnvironments()
	managedEnvs := make(map[string]bool)
	for _, env := range envs {
		managedEnvs[env.ID] = true
	}

	// Remove orphaned state directories
	for _, entry := range entries {
		if entry.IsDir() {
			envID := entry.Name()
			if !managedEnvs[envID] {
				log.Printf("[CleanupManager] Removing orphaned state directory %s", envID)
				statePath := filepath.Join(stateDir, envID)
				if err := os.RemoveAll(statePath); err != nil {
					log.Printf("[CleanupManager] Failed to remove orphaned state directory %s: %v", envID, err)
				}
			}
		}
	}

	return nil
}

// ForceCleanupAll forcefully cleans up all resources
func (cm *CleanupManager) ForceCleanupAll(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	log.Println("[CleanupManager] Starting force cleanup of all resources")

	// Clean up all environments
	if err := cm.manager.CleanupAll(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to cleanup all environments: %v", err)
	}

	// Clean up all containers in test network
	containers, err := cm.dockerMgr.ListContainers(ctx)
	if err == nil {
		for _, containerID := range containers {
			if err := cm.dockerMgr.RemoveContainer(ctx, containerID); err != nil {
				log.Printf("[CleanupManager] Failed to remove container %s: %v", containerID, err)
			}
		}
	}

	// Prune containers
	if err := cm.dockerMgr.PruneContainers(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to prune containers: %v", err)
	}

	// Remove Docker network
	if err := cm.dockerMgr.RemoveNetwork(ctx); err != nil {
		log.Printf("[CleanupManager] Failed to remove network: %v", err)
	}

	log.Println("[CleanupManager] Force cleanup completed")
	return nil
}

// SetMaxAge sets the maximum age for environments before cleanup
func (cm *CleanupManager) SetMaxAge(maxAge time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.maxAge = maxAge
}

// SetCleanupInterval sets the cleanup interval
func (cm *CleanupManager) SetCleanupInterval(interval time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cleanupInterval = interval
}
