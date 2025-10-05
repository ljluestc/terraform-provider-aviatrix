package envmanager

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// HealthChecker performs health checks on environments
type HealthChecker struct {
	httpClient  *http.Client
	pingTimeout time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		pingTimeout: 5 * time.Second,
	}
}

// CheckEnvironment performs a comprehensive health check on an environment
func (hc *HealthChecker) CheckEnvironment(ctx context.Context, env *Environment) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		EnvironmentID: env.ID,
		Timestamp:     time.Now(),
		Checks:        make(map[string]CheckResult),
	}

	// Check 1: Environment status
	result.Checks["status"] = hc.checkStatus(env)

	// Check 2: Container health (if containerized)
	if env.ContainerID != "" {
		result.Checks["container"] = hc.checkContainer(ctx, env)
	}

	// Check 3: Network connectivity
	if ip, ok := env.GetResource("container_ip"); ok {
		result.Checks["network"] = hc.checkNetwork(ctx, ip)
	}

	// Check 4: Controller connectivity (if configured)
	if controllerIP, ok := env.GetResource("controller_ip"); ok {
		result.Checks["controller"] = hc.checkController(ctx, controllerIP)
	}

	// Check 5: State file integrity
	result.Checks["state"] = hc.checkStateFile(env)

	// Calculate overall health
	result.Healthy = hc.calculateOverallHealth(result.Checks)

	return result, nil
}

// checkStatus checks if the environment status is valid
func (hc *HealthChecker) checkStatus(env *Environment) CheckResult {
	status := env.GetStatus()

	if status == StatusReady || status == StatusInUse {
		return CheckResult{
			Passed:  true,
			Message: fmt.Sprintf("Environment status is %s", status),
		}
	}

	return CheckResult{
		Passed:  false,
		Message: fmt.Sprintf("Environment status is %s (expected ready or in_use)", status),
	}
}

// checkContainer checks if the container is running and healthy
func (hc *HealthChecker) checkContainer(ctx context.Context, env *Environment) CheckResult {
	dm := NewDockerManager("", "")

	healthy, err := dm.CheckContainerHealth(ctx, env.ContainerID)
	if err != nil {
		return CheckResult{
			Passed:  false,
			Message: fmt.Sprintf("Failed to check container health: %v", err),
		}
	}

	if !healthy {
		return CheckResult{
			Passed:  false,
			Message: "Container is not healthy",
		}
	}

	return CheckResult{
		Passed:  true,
		Message: "Container is running and healthy",
	}
}

// checkNetwork checks network connectivity
func (hc *HealthChecker) checkNetwork(ctx context.Context, ip string) CheckResult {
	// Try to connect to common ports
	ports := []string{"22", "80", "443"}

	for _, port := range ports {
		address := net.JoinHostPort(ip, port)
		conn, err := net.DialTimeout("tcp", address, hc.pingTimeout)
		if err == nil {
			conn.Close()
			return CheckResult{
				Passed:  true,
				Message: fmt.Sprintf("Network connectivity verified (connected to %s:%s)", ip, port),
			}
		}
	}

	return CheckResult{
		Passed:  false,
		Message: fmt.Sprintf("Could not establish network connection to %s", ip),
	}
}

// checkController checks Aviatrix controller connectivity
func (hc *HealthChecker) checkController(ctx context.Context, controllerIP string) CheckResult {
	// Try HTTPS connection to controller
	url := fmt.Sprintf("https://%s", controllerIP)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return CheckResult{
			Passed:  false,
			Message: fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return CheckResult{
			Passed:  false,
			Message: fmt.Sprintf("Controller not reachable at %s: %v", controllerIP, err),
		}
	}
	defer resp.Body.Close()

	return CheckResult{
		Passed:  true,
		Message: fmt.Sprintf("Controller is reachable at %s (status: %d)", controllerIP, resp.StatusCode),
	}
}

// checkStateFile checks if the Terraform state file exists and is valid
func (hc *HealthChecker) checkStateFile(env *Environment) CheckResult {
	if env.StateFile == "" {
		return CheckResult{
			Passed:  false,
			Message: "State file path not configured",
		}
	}

	tm := NewTerraformManager("", "", "local")
	if err := tm.ValidateState(context.Background(), env); err != nil {
		return CheckResult{
			Passed:  false,
			Message: fmt.Sprintf("State file validation failed: %v", err),
		}
	}

	return CheckResult{
		Passed:  true,
		Message: "State file is valid",
	}
}

// calculateOverallHealth determines overall health based on individual checks
func (hc *HealthChecker) calculateOverallHealth(checks map[string]CheckResult) bool {
	// Critical checks that must pass
	criticalChecks := []string{"status", "state"}

	for _, checkName := range criticalChecks {
		if result, ok := checks[checkName]; ok {
			if !result.Passed {
				return false
			}
		}
	}

	// If critical checks pass, environment is considered healthy
	// even if non-critical checks fail
	return true
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	EnvironmentID string                  `json:"environment_id"`
	Timestamp     time.Time               `json:"timestamp"`
	Healthy       bool                    `json:"healthy"`
	Checks        map[string]CheckResult  `json:"checks"`
}

// CheckResult represents the result of an individual check
type CheckResult struct {
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// Report generates a human-readable health check report
func (r *HealthCheckResult) Report() string {
	report := fmt.Sprintf("Health Check Report for Environment: %s\n", r.EnvironmentID)
	report += fmt.Sprintf("Timestamp: %s\n", r.Timestamp.Format(time.RFC3339))
	report += fmt.Sprintf("Overall Status: %s\n\n", map[bool]string{true: "HEALTHY", false: "UNHEALTHY"}[r.Healthy])

	report += "Individual Checks:\n"
	for name, result := range r.Checks {
		status := "✗ FAIL"
		if result.Passed {
			status = "✓ PASS"
		}
		report += fmt.Sprintf("  %s %s: %s\n", status, name, result.Message)
	}

	return report
}
