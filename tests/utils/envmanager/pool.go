package envmanager

import (
	"fmt"
	"sync"
)

// EnvironmentPool manages a pool of reusable test environments
type EnvironmentPool struct {
	pools map[Provider]chan *Environment
	size  int
	mu    sync.RWMutex
}

// NewEnvironmentPool creates a new environment pool
func NewEnvironmentPool(size int) *EnvironmentPool {
	return &EnvironmentPool{
		pools: map[Provider]chan *Environment{
			ProviderAWS:   make(chan *Environment, size),
			ProviderAzure: make(chan *Environment, size),
			ProviderGCP:   make(chan *Environment, size),
			ProviderOCI:   make(chan *Environment, size),
		},
		size: size,
	}
}

// Acquire gets an environment from the pool
func (p *EnvironmentPool) Acquire(provider Provider) *Environment {
	p.mu.RLock()
	pool, ok := p.pools[provider]
	p.mu.RUnlock()

	if !ok {
		return nil
	}

	select {
	case env := <-pool:
		return env
	default:
		return nil
	}
}

// Release returns an environment to the pool
func (p *EnvironmentPool) Release(env *Environment) error {
	p.mu.RLock()
	pool, ok := p.pools[env.Provider]
	p.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown provider: %s", env.Provider)
	}

	select {
	case pool <- env:
		return nil
	default:
		// Pool is full, don't return the environment
		return fmt.Errorf("pool for provider %s is full", env.Provider)
	}
}

// Size returns the pool size for a provider
func (p *EnvironmentPool) Size(provider Provider) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pool, ok := p.pools[provider]
	if !ok {
		return 0
	}

	return len(pool)
}

// Available returns the number of available environments for a provider
func (p *EnvironmentPool) Available(provider Provider) int {
	return p.Size(provider)
}

// Stats returns pool statistics
func (p *EnvironmentPool) Stats() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]int)
	for provider, pool := range p.pools {
		stats[fmt.Sprintf("%s_available", provider)] = len(pool)
		stats[fmt.Sprintf("%s_capacity", provider)] = cap(pool)
	}

	totalAvailable := 0
	totalCapacity := 0
	for _, pool := range p.pools {
		totalAvailable += len(pool)
		totalCapacity += cap(pool)
	}

	stats["total_available"] = totalAvailable
	stats["total_capacity"] = totalCapacity

	return stats
}

// Clear removes all environments from all pools
func (p *EnvironmentPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for provider, pool := range p.pools {
		// Drain the channel
		for len(pool) > 0 {
			<-pool
		}
		// Recreate the channel
		p.pools[provider] = make(chan *Environment, p.size)
	}
}

// WarmUp pre-populates the pool with environments
func (p *EnvironmentPool) WarmUp(provider Provider, count int, factory func() *Environment) error {
	p.mu.RLock()
	pool, ok := p.pools[provider]
	p.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	for i := 0; i < count; i++ {
		env := factory()
		if env == nil {
			return fmt.Errorf("factory returned nil environment")
		}

		select {
		case pool <- env:
			// Successfully added to pool
		default:
			return fmt.Errorf("pool is full, cannot add more environments")
		}
	}

	return nil
}
