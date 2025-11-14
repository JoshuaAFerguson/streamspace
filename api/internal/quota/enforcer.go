package quota

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/streamspace/streamspace/api/internal/db"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Limits represents resource limits for a user or group
type Limits struct {
	// Maximum number of concurrent sessions
	MaxSessions int `json:"max_sessions"`

	// Maximum CPU per session (in millicores)
	MaxCPUPerSession int64 `json:"max_cpu_per_session"`

	// Maximum memory per session (in MiB)
	MaxMemoryPerSession int64 `json:"max_memory_per_session"`

	// Maximum total CPU across all sessions (in millicores)
	MaxTotalCPU int64 `json:"max_total_cpu"`

	// Maximum total memory across all sessions (in MiB)
	MaxTotalMemory int64 `json:"max_total_memory"`

	// Maximum storage per user (in GiB)
	MaxStorage int64 `json:"max_storage"`

	// Maximum GPU count per session
	MaxGPUPerSession int `json:"max_gpu_per_session"`
}

// Usage represents current resource usage for a user
type Usage struct {
	// Current number of active sessions
	ActiveSessions int `json:"active_sessions"`

	// Total CPU usage across all sessions (in millicores)
	TotalCPU int64 `json:"total_cpu"`

	// Total memory usage across all sessions (in MiB)
	TotalMemory int64 `json:"total_memory"`

	// Total storage usage (in GiB)
	TotalStorage int64 `json:"total_storage"`

	// Total GPU count across all sessions
	TotalGPU int `json:"total_gpu"`
}

// Enforcer enforces resource quotas for users and groups
type Enforcer struct {
	userDB  *db.UserDB
	groupDB *db.GroupDB
}

// NewEnforcer creates a new quota enforcer
func NewEnforcer(userDB *db.UserDB, groupDB *db.GroupDB) *Enforcer {
	return &Enforcer{
		userDB:  userDB,
		groupDB: groupDB,
	}
}

// GetUserLimits retrieves the resource limits for a user
// It combines user-specific limits with group limits (taking the most restrictive)
func (e *Enforcer) GetUserLimits(ctx context.Context, username string) (*Limits, error) {
	// Get user from database
	user, err := e.userDB.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Start with default limits (for free tier users)
	limits := &Limits{
		MaxSessions:         5,
		MaxCPUPerSession:    2000,  // 2 CPU cores
		MaxMemoryPerSession: 4096,  // 4 GiB
		MaxTotalCPU:         4000,  // 4 CPU cores total
		MaxTotalMemory:      8192,  // 8 GiB total
		MaxStorage:          50,    // 50 GiB
		MaxGPUPerSession:    0,     // No GPU by default
	}

	// Override with user-specific limits if set
	if user.Quota != nil {
		if user.Quota.MaxSessions > 0 {
			limits.MaxSessions = user.Quota.MaxSessions
		}
		if user.Quota.MaxCPUPerSession > 0 {
			limits.MaxCPUPerSession = user.Quota.MaxCPUPerSession
		}
		if user.Quota.MaxMemoryPerSession > 0 {
			limits.MaxMemoryPerSession = user.Quota.MaxMemoryPerSession
		}
		if user.Quota.MaxTotalCPU > 0 {
			limits.MaxTotalCPU = user.Quota.MaxTotalCPU
		}
		if user.Quota.MaxTotalMemory > 0 {
			limits.MaxTotalMemory = user.Quota.MaxTotalMemory
		}
		if user.Quota.MaxStorage > 0 {
			limits.MaxStorage = user.Quota.MaxStorage
		}
		if user.Quota.MaxGPUPerSession >= 0 {
			limits.MaxGPUPerSession = user.Quota.MaxGPUPerSession
		}
	}

	// Check group limits and apply the most restrictive
	if len(user.Groups) > 0 {
		for _, groupName := range user.Groups {
			group, err := e.groupDB.GetByName(ctx, groupName)
			if err != nil {
				continue // Skip groups that don't exist
			}

			if group.Quota != nil {
				// Apply most restrictive limits
				if group.Quota.MaxSessions > 0 && group.Quota.MaxSessions < limits.MaxSessions {
					limits.MaxSessions = group.Quota.MaxSessions
				}
				if group.Quota.MaxCPUPerSession > 0 && group.Quota.MaxCPUPerSession < limits.MaxCPUPerSession {
					limits.MaxCPUPerSession = group.Quota.MaxCPUPerSession
				}
				if group.Quota.MaxMemoryPerSession > 0 && group.Quota.MaxMemoryPerSession < limits.MaxMemoryPerSession {
					limits.MaxMemoryPerSession = group.Quota.MaxMemoryPerSession
				}
				if group.Quota.MaxTotalCPU > 0 && group.Quota.MaxTotalCPU < limits.MaxTotalCPU {
					limits.MaxTotalCPU = group.Quota.MaxTotalCPU
				}
				if group.Quota.MaxTotalMemory > 0 && group.Quota.MaxTotalMemory < limits.MaxTotalMemory {
					limits.MaxTotalMemory = group.Quota.MaxTotalMemory
				}
				if group.Quota.MaxStorage > 0 && group.Quota.MaxStorage < limits.MaxStorage {
					limits.MaxStorage = group.Quota.MaxStorage
				}
				if group.Quota.MaxGPUPerSession >= 0 && group.Quota.MaxGPUPerSession < limits.MaxGPUPerSession {
					limits.MaxGPUPerSession = group.Quota.MaxGPUPerSession
				}
			}
		}
	}

	return limits, nil
}

// CheckSessionCreation validates if a user can create a new session with the requested resources
func (e *Enforcer) CheckSessionCreation(ctx context.Context, username string, requestedCPU, requestedMemory int64, requestedGPU int, currentUsage *Usage) error {
	limits, err := e.GetUserLimits(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to get user limits: %w", err)
	}

	// Check session count
	if currentUsage.ActiveSessions >= limits.MaxSessions {
		return fmt.Errorf("session quota exceeded: %d/%d sessions active", currentUsage.ActiveSessions, limits.MaxSessions)
	}

	// Check CPU per session
	if requestedCPU > limits.MaxCPUPerSession {
		return fmt.Errorf("CPU quota exceeded: requested %dm, limit is %dm per session", requestedCPU, limits.MaxCPUPerSession)
	}

	// Check memory per session
	if requestedMemory > limits.MaxMemoryPerSession {
		return fmt.Errorf("memory quota exceeded: requested %dMi, limit is %dMi per session", requestedMemory, limits.MaxMemoryPerSession)
	}

	// Check total CPU
	totalCPU := currentUsage.TotalCPU + requestedCPU
	if totalCPU > limits.MaxTotalCPU {
		return fmt.Errorf("total CPU quota exceeded: would use %dm, limit is %dm", totalCPU, limits.MaxTotalCPU)
	}

	// Check total memory
	totalMemory := currentUsage.TotalMemory + requestedMemory
	if totalMemory > limits.MaxTotalMemory {
		return fmt.Errorf("total memory quota exceeded: would use %dMi, limit is %dMi", totalMemory, limits.MaxTotalMemory)
	}

	// Check GPU per session
	if requestedGPU > limits.MaxGPUPerSession {
		return fmt.Errorf("GPU quota exceeded: requested %d, limit is %d per session", requestedGPU, limits.MaxGPUPerSession)
	}

	return nil
}

// CalculateUsage calculates current resource usage from a list of pods
func (e *Enforcer) CalculateUsage(pods []corev1.Pod) *Usage {
	usage := &Usage{}

	for _, pod := range pods {
		// Only count running pods
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		usage.ActiveSessions++

		// Sum up resource requests from all containers
		for _, container := range pod.Spec.Containers {
			// CPU
			if cpu := container.Resources.Requests[corev1.ResourceCPU]; !cpu.IsZero() {
				usage.TotalCPU += cpu.MilliValue()
			}

			// Memory (convert to MiB)
			if memory := container.Resources.Requests[corev1.ResourceMemory]; !memory.IsZero() {
				usage.TotalMemory += memory.Value() / (1024 * 1024)
			}

			// GPU (nvidia.com/gpu)
			if gpu := container.Resources.Requests["nvidia.com/gpu"]; !gpu.IsZero() {
				usage.TotalGPU += int(gpu.Value())
			}
		}
	}

	return usage
}

// ParseResourceQuantity parses a Kubernetes resource quantity string (e.g., "2000m", "4Gi")
func ParseResourceQuantity(quantity string, resourceType string) (int64, error) {
	q, err := resource.ParseQuantity(quantity)
	if err != nil {
		return 0, fmt.Errorf("invalid resource quantity: %w", err)
	}

	switch resourceType {
	case "cpu":
		// Return millicores
		return q.MilliValue(), nil
	case "memory":
		// Return MiB
		return q.Value() / (1024 * 1024), nil
	default:
		return q.Value(), nil
	}
}

// FormatResourceQuantity formats a resource value back to Kubernetes format
func FormatResourceQuantity(value int64, resourceType string) string {
	switch resourceType {
	case "cpu":
		// Convert millicores to string
		return fmt.Sprintf("%dm", value)
	case "memory":
		// Convert MiB to string
		return fmt.Sprintf("%dMi", value)
	default:
		return fmt.Sprintf("%d", value)
	}
}

// ValidateResourceRequest validates that a resource request is within acceptable bounds
func (e *Enforcer) ValidateResourceRequest(cpuStr, memoryStr string) (cpu, memory int64, err error) {
	// Parse CPU
	if cpuStr != "" {
		cpu, err = ParseResourceQuantity(cpuStr, "cpu")
		if err != nil {
			return 0, 0, fmt.Errorf("invalid CPU quantity: %w", err)
		}

		// Minimum 100m (0.1 CPU)
		if cpu < 100 {
			return 0, 0, fmt.Errorf("CPU request too low: minimum 100m")
		}

		// Maximum 64 CPUs (64000m)
		if cpu > 64000 {
			return 0, 0, fmt.Errorf("CPU request too high: maximum 64000m")
		}
	}

	// Parse memory
	if memoryStr != "" {
		memory, err = ParseResourceQuantity(memoryStr, "memory")
		if err != nil {
			return 0, 0, fmt.Errorf("invalid memory quantity: %w", err)
		}

		// Minimum 128Mi
		if memory < 128 {
			return 0, 0, fmt.Errorf("memory request too low: minimum 128Mi")
		}

		// Maximum 512Gi (524288Mi)
		if memory > 524288 {
			return 0, 0, fmt.Errorf("memory request too high: maximum 512Gi")
		}
	}

	return cpu, memory, nil
}

// GetDefaultResources returns default resource requests based on template category
func GetDefaultResources(category string) (cpu, memory string) {
	switch strings.ToLower(category) {
	case "browsers", "web browsers":
		return "1000m", "2048Mi" // 1 CPU, 2 GiB
	case "development", "ide":
		return "2000m", "4096Mi" // 2 CPUs, 4 GiB
	case "design", "graphics":
		return "2000m", "8192Mi" // 2 CPUs, 8 GiB
	case "gaming", "emulation":
		return "2000m", "4096Mi" // 2 CPUs, 4 GiB
	case "productivity", "office":
		return "1000m", "2048Mi" // 1 CPU, 2 GiB
	case "media", "video editing":
		return "4000m", "8192Mi" // 4 CPUs, 8 GiB
	case "ai", "machine learning":
		return "4000m", "16384Mi" // 4 CPUs, 16 GiB
	default:
		return "1000m", "2048Mi" // 1 CPU, 2 GiB (default)
	}
}

// QuotaExceededError represents a quota exceeded error
type QuotaExceededError struct {
	Message string
	Limit   interface{}
	Current interface{}
}

func (e *QuotaExceededError) Error() string {
	return e.Message
}

// IsQuotaExceeded checks if an error is a quota exceeded error
func IsQuotaExceeded(err error) bool {
	_, ok := err.(*QuotaExceededError)
	return ok
}

// ParseGPURequest parses a GPU request from a string
func ParseGPURequest(gpuStr string) (int, error) {
	if gpuStr == "" || gpuStr == "0" {
		return 0, nil
	}

	gpu, err := strconv.Atoi(gpuStr)
	if err != nil {
		return 0, fmt.Errorf("invalid GPU count: %w", err)
	}

	if gpu < 0 {
		return 0, fmt.Errorf("GPU count cannot be negative")
	}

	if gpu > 8 {
		return 0, fmt.Errorf("GPU count too high: maximum 8")
	}

	return gpu, nil
}
