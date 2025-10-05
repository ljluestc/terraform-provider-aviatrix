package envmanager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TerraformManager handles Terraform operations with isolated state
type TerraformManager struct {
	workDir      string
	stateDir     string
	stateBackend string
}

// NewTerraformManager creates a new Terraform manager
func NewTerraformManager(workDir, stateDir, stateBackend string) *TerraformManager {
	return &TerraformManager{
		workDir:      workDir,
		stateDir:     stateDir,
		stateBackend: stateBackend,
	}
}

// InitEnvironment initializes Terraform for an environment
func (tm *TerraformManager) InitEnvironment(ctx context.Context, env *Environment, moduleDir string) error {
	// Create state directory for this environment
	envStateDir := filepath.Join(tm.stateDir, env.ID)
	if err := os.MkdirAll(envStateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Set state file path
	env.StateFile = filepath.Join(envStateDir, "terraform.tfstate")

	// Initialize Terraform
	args := []string{"init"}
	if tm.stateBackend == "local" {
		args = append(args, "-backend=true")
	}

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_DATA_DIR=%s", filepath.Join(envStateDir, ".terraform")),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %w\nOutput: %s", err, string(output))
	}

	env.AddMetadata("terraform_initialized", "true")
	env.AddMetadata("module_dir", moduleDir)

	return nil
}

// Plan creates a Terraform plan for the environment
func (tm *TerraformManager) Plan(ctx context.Context, env *Environment, vars map[string]string) (string, error) {
	moduleDir, ok := env.GetMetadata("module_dir")
	if !ok {
		return "", fmt.Errorf("module directory not set for environment %s", env.ID)
	}

	envStateDir := filepath.Join(tm.stateDir, env.ID)
	planFile := filepath.Join(envStateDir, "tfplan")

	args := []string{"plan", "-out=" + planFile}
	for key, value := range vars {
		args = append(args, fmt.Sprintf("-var=%s=%s", key, value))
	}

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_DATA_DIR=%s", filepath.Join(envStateDir, ".terraform")),
		fmt.Sprintf("TF_STATE=%s", env.StateFile),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("terraform plan failed: %w\nOutput: %s", err, string(output))
	}

	env.AddMetadata("plan_file", planFile)
	return string(output), nil
}

// Apply applies a Terraform plan
func (tm *TerraformManager) Apply(ctx context.Context, env *Environment, vars map[string]string) error {
	moduleDir, ok := env.GetMetadata("module_dir")
	if !ok {
		return fmt.Errorf("module directory not set for environment %s", env.ID)
	}

	envStateDir := filepath.Join(tm.stateDir, env.ID)

	var args []string
	planFile, hasPlan := env.GetMetadata("plan_file")
	if hasPlan {
		args = []string{"apply", "-auto-approve", planFile}
	} else {
		args = []string{"apply", "-auto-approve"}
		for key, value := range vars {
			args = append(args, fmt.Sprintf("-var=%s=%s", key, value))
		}
	}

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_DATA_DIR=%s", filepath.Join(envStateDir, ".terraform")),
		fmt.Sprintf("TF_STATE=%s", env.StateFile),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w\nOutput: %s", err, string(output))
	}

	// Parse outputs and add to environment resources
	if err := tm.parseOutputs(ctx, env); err != nil {
		return fmt.Errorf("failed to parse outputs: %w", err)
	}

	env.SetStatus(StatusReady)
	return nil
}

// Destroy destroys all resources in the environment
func (tm *TerraformManager) Destroy(ctx context.Context, env *Environment) error {
	moduleDir, ok := env.GetMetadata("module_dir")
	if !ok {
		return fmt.Errorf("module directory not set for environment %s", env.ID)
	}

	envStateDir := filepath.Join(tm.stateDir, env.ID)

	cmd := exec.CommandContext(ctx, "terraform", "destroy", "-auto-approve")
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_DATA_DIR=%s", filepath.Join(envStateDir, ".terraform")),
		fmt.Sprintf("TF_STATE=%s", env.StateFile),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w\nOutput: %s", err, string(output))
	}

	// Clean up state directory
	if err := os.RemoveAll(envStateDir); err != nil {
		return fmt.Errorf("failed to remove state directory: %w", err)
	}

	return nil
}

// parseOutputs extracts Terraform outputs and adds them to environment resources
func (tm *TerraformManager) parseOutputs(ctx context.Context, env *Environment) error {
	moduleDir, ok := env.GetMetadata("module_dir")
	if !ok {
		return fmt.Errorf("module directory not set for environment %s", env.ID)
	}

	envStateDir := filepath.Join(tm.stateDir, env.ID)

	cmd := exec.CommandContext(ctx, "terraform", "output", "-json")
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_DATA_DIR=%s", filepath.Join(envStateDir, ".terraform")),
		fmt.Sprintf("TF_STATE=%s", env.StateFile),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// It's okay if there are no outputs
		if strings.Contains(string(output), "no outputs") {
			return nil
		}
		return fmt.Errorf("terraform output failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output and add to resources
	// For now, just store raw JSON
	env.AddMetadata("terraform_outputs", string(output))

	return nil
}

// GetOutput retrieves a specific Terraform output value
func (tm *TerraformManager) GetOutput(ctx context.Context, env *Environment, outputName string) (string, error) {
	moduleDir, ok := env.GetMetadata("module_dir")
	if !ok {
		return "", fmt.Errorf("module directory not set for environment %s", env.ID)
	}

	envStateDir := filepath.Join(tm.stateDir, env.ID)

	cmd := exec.CommandContext(ctx, "terraform", "output", "-raw", outputName)
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_DATA_DIR=%s", filepath.Join(envStateDir, ".terraform")),
		fmt.Sprintf("TF_STATE=%s", env.StateFile),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get output %s: %w\nOutput: %s", outputName, err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// Refresh refreshes the Terraform state
func (tm *TerraformManager) Refresh(ctx context.Context, env *Environment) error {
	moduleDir, ok := env.GetMetadata("module_dir")
	if !ok {
		return fmt.Errorf("module directory not set for environment %s", env.ID)
	}

	envStateDir := filepath.Join(tm.stateDir, env.ID)

	cmd := exec.CommandContext(ctx, "terraform", "refresh")
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_DATA_DIR=%s", filepath.Join(envStateDir, ".terraform")),
		fmt.Sprintf("TF_STATE=%s", env.StateFile),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform refresh failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// ValidateState checks if the Terraform state is valid
func (tm *TerraformManager) ValidateState(ctx context.Context, env *Environment) error {
	if _, err := os.Stat(env.StateFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("state file does not exist: %s", env.StateFile)
		}
		return fmt.Errorf("failed to check state file: %w", err)
	}

	return nil
}
