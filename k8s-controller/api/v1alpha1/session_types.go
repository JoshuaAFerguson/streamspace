// Package v1alpha1 contains API Schema definitions for the stream v1alpha1 API group.
//
// This package defines the custom resource definitions (CRDs) for StreamSpace:
//   - Session: Represents a user's containerized workspace session
//   - Template: Defines application templates that can be launched as sessions
//
// These types are automatically registered with the Kubernetes API server when the
// controller starts, enabling kubectl operations like:
//   kubectl get sessions
//   kubectl describe session user1-firefox
//   kubectl delete session user1-firefox
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SessionSpec defines the desired state of a Session.
//
// The spec contains all user-configurable parameters for a session.
// When the spec changes, the controller reconciles the actual state to match.
//
// Example:
//
//	spec:
//	  user: alice
//	  template: firefox-browser
//	  state: running
//	  resources:
//	    requests:
//	      memory: "2Gi"
//	      cpu: "1000m"
//	    limits:
//	      memory: "4Gi"
//	      cpu: "2000m"
//	  persistentHome: true
//	  idleTimeout: "30m"
//	  maxSessionDuration: "8h"
//	  tags: ["development", "web-browsing"]
type SessionSpec struct {
	// User specifies the username who owns this session.
	// The controller uses this to:
	//   - Create/mount user-specific PersistentVolumeClaims
	//   - Apply user resource quotas
	//   - Track session ownership
	//
	// Required: Yes
	// Example: "alice", "bob@example.com"
	// +kubebuilder:validation:Required
	User string `json:"user"`

	// Template specifies the name of the Template resource to use for this session.
	// The Template defines:
	//   - Container image to run
	//   - Default resource requirements
	//   - Port configurations
	//   - Environment variables
	//
	// The Template must exist in the same namespace before creating the Session.
	//
	// Required: Yes
	// Example: "firefox-browser", "vscode-dev"
	// +kubebuilder:validation:Required
	Template string `json:"template"`

	// State defines the desired lifecycle state of the session.
	//
	// Valid values:
	//   - "running": Session pod is running and accepting connections
	//   - "hibernated": Session pod is scaled to zero replicas (sleeping)
	//   - "terminated": Session is deleted and all resources are cleaned up
	//
	// State transitions:
	//   running → hibernated: Controller scales Deployment to 0 replicas
	//   hibernated → running: Controller scales Deployment to 1 replica
	//   * → terminated: Controller deletes all session resources
	//
	// Default: "running"
	// +kubebuilder:validation:Enum=running;hibernated;terminated
	// +kubebuilder:default=running
	State string `json:"state"`

	// Resources specifies CPU and memory limits for the session pod.
	//
	// If not specified, defaults from the Template are used.
	// Requests and limits can be overridden independently.
	//
	// Example:
	//   resources:
	//     requests:
	//       memory: "2Gi"
	//       cpu: "1000m"
	//     limits:
	//       memory: "4Gi"
	//       cpu: "2000m"
	//
	// Optional: Yes
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// PersistentHome determines whether to mount a persistent volume for user data.
	//
	// When enabled:
	//   - A PVC named "home-{user}" is created if it doesn't exist
	//   - The PVC is mounted at /config in the container
	//   - User data persists across session lifecycles
	//
	// When disabled:
	//   - No PVC is created
	//   - Data is ephemeral and lost when session terminates
	//
	// Default: true
	// Optional: Yes
	// +kubebuilder:default=true
	// +optional
	PersistentHome bool `json:"persistentHome,omitempty"`

	// IdleTimeout specifies the duration of inactivity before auto-hibernation.
	//
	// Format: Duration string (e.g., "30m", "1h", "2h30m")
	//
	// The HibernationReconciler checks lastActivity timestamps and transitions
	// sessions to "hibernated" state when the idle timeout is exceeded.
	//
	// Set to empty string to disable auto-hibernation.
	//
	// Example: "30m", "1h", "2h30m"
	// Optional: Yes
	// +optional
	IdleTimeout string `json:"idleTimeout,omitempty"`

	// MaxSessionDuration specifies the maximum lifetime of a session.
	//
	// Format: Duration string (e.g., "8h", "24h")
	//
	// After this duration, the session is automatically terminated regardless
	// of activity. This is useful for preventing resource leaks from forgotten sessions.
	//
	// Set to empty string for unlimited duration.
	//
	// Example: "8h", "24h"
	// Optional: Yes
	// +optional
	MaxSessionDuration string `json:"maxSessionDuration,omitempty"`

	// Tags are user-defined labels for organizing and filtering sessions.
	//
	// Tags can be used to:
	//   - Group sessions by project or team
	//   - Filter sessions in the UI
	//   - Apply batch operations
	//
	// Example: ["development", "web-browsing", "project-alpha"]
	// Optional: Yes
	// +optional
	Tags []string `json:"tags,omitempty"`
}

// SessionStatus defines the observed state of a Session.
//
// The status is managed entirely by the controller and should not be modified by users.
// It provides real-time information about the session's current state, resources, and health.
//
// Example:
//
//	status:
//	  phase: Running
//	  podName: ss-alice-firefox-abc123
//	  url: https://alice-firefox.streamspace.local
//	  lastActivity: "2025-01-15T14:30:00Z"
//	  resourceUsage:
//	    memory: "1.2Gi"
//	    cpu: "450m"
//	  conditions:
//	    - type: Ready
//	      status: "True"
//	      lastTransitionTime: "2025-01-15T14:25:00Z"
type SessionStatus struct {
	// Phase indicates the current lifecycle phase of the session.
	//
	// Possible values:
	//   - "Pending": Resources are being created
	//   - "Running": Pod is running and ready
	//   - "Hibernated": Session is scaled to zero (sleeping)
	//   - "Failed": Session encountered an error
	//   - "Terminated": Session is being deleted
	//
	// The phase is derived from the underlying Kubernetes resources (Pod, Deployment).
	//
	// Optional: Yes (computed by controller)
	// +optional
	Phase string `json:"phase,omitempty"`

	// PodName is the name of the Kubernetes Pod running this session.
	//
	// This can be used to:
	//   - View pod logs: kubectl logs -n streamspace {podName}
	//   - Exec into pod: kubectl exec -n streamspace {podName} -- /bin/bash
	//   - Debug pod issues: kubectl describe pod -n streamspace {podName}
	//
	// Empty when session is hibernated or terminated.
	//
	// Optional: Yes (computed by controller)
	// +optional
	PodName string `json:"podName,omitempty"`

	// URL is the HTTP(S) endpoint to access this session in a web browser.
	//
	// Format: https://{session-name}.{ingress-domain}
	// Example: https://alice-firefox.streamspace.local
	//
	// The URL is constructed from:
	//   - Session name (metadata.name)
	//   - Ingress domain (from controller configuration)
	//
	// Empty when session is hibernated or terminated.
	//
	// Optional: Yes (computed by controller)
	// +optional
	URL string `json:"url,omitempty"`

	// LastActivity is the timestamp of the last user interaction with this session.
	//
	// This timestamp is updated by:
	//   - API backend on WebSocket connections
	//   - Activity tracker on keyboard/mouse events
	//   - Heartbeat requests from the UI
	//
	// Used by HibernationReconciler to determine when to hibernate idle sessions.
	//
	// Optional: Yes (updated by external components)
	// +optional
	LastActivity *metav1.Time `json:"lastActivity,omitempty"`

	// ResourceUsage tracks the current CPU and memory consumption of the session pod.
	//
	// Values are fetched from Kubernetes metrics API and updated periodically.
	// Used for:
	//   - Quota enforcement
	//   - Dashboard displays
	//   - Usage analytics
	//   - Auto-scaling decisions
	//
	// Optional: Yes (computed by controller)
	// +optional
	ResourceUsage *ResourceUsage `json:"resourceUsage,omitempty"`

	// Conditions represent the latest available observations of the session's state.
	//
	// Standard condition types:
	//   - "Ready": Pod is running and accepting connections
	//   - "PVCBound": Persistent volume is bound and mounted
	//   - "TemplateResolved": Template was found and applied
	//   - "QuotaExceeded": User has exceeded resource quotas
	//
	// Conditions follow the Kubernetes standard:
	//   - type: Condition name
	//   - status: True, False, or Unknown
	//   - reason: Machine-readable reason code
	//   - message: Human-readable explanation
	//   - lastTransitionTime: When this condition last changed
	//
	// Optional: Yes (managed by controller)
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ResourceUsage tracks current resource consumption for a session.
//
// Values are fetched from the Kubernetes metrics API (metrics-server required).
// Format follows Kubernetes resource quantity conventions.
//
// Example:
//
//	resourceUsage:
//	  memory: "1.2Gi"   # 1.2 gibibytes
//	  cpu: "450m"       # 450 millicores (0.45 CPU cores)
type ResourceUsage struct {
	// Memory is the current memory usage in Kubernetes quantity format.
	// Examples: "512Mi", "1.5Gi", "2048M"
	Memory string `json:"memory,omitempty"`

	// CPU is the current CPU usage in Kubernetes quantity format.
	// Examples: "100m" (0.1 cores), "1" (1 core), "2500m" (2.5 cores)
	CPU string `json:"cpu,omitempty"`
}

// Session is the Schema for the sessions API.
//
// A Session represents a single user's containerized workspace session.
// It creates and manages:
//   - A Kubernetes Deployment (for pod lifecycle)
//   - A Service (for networking)
//   - A PersistentVolumeClaim (for persistent storage, optional)
//   - An Ingress (for external access)
//
// Sessions support auto-hibernation to save resources when idle.
//
// Example usage:
//
//	kubectl apply -f - <<EOF
//	apiVersion: stream.space/v1alpha1
//	kind: Session
//	metadata:
//	  name: alice-firefox
//	  namespace: streamspace
//	spec:
//	  user: alice
//	  template: firefox-browser
//	  state: running
//	  resources:
//	    requests:
//	      memory: "2Gi"
//	      cpu: "1000m"
//	  persistentHome: true
//	  idleTimeout: "30m"
//	EOF
//
// Kubebuilder annotations:
//   - +kubebuilder:object:root=true - Marks this as a root Kubernetes object
//   - +kubebuilder:subresource:status - Enables /status subresource (separates spec and status updates)
//   - +kubebuilder:resource:shortName=ss - Allows "kubectl get ss" as shorthand
//   - +kubebuilder:printcolumn - Defines columns shown in "kubectl get" output
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ss
// +kubebuilder:printcolumn:name="User",type=string,JSONPath=`.spec.user`
// +kubebuilder:printcolumn:name="Template",type=string,JSONPath=`.spec.template`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.spec.state`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Session struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SessionSpec   `json:"spec,omitempty"`
	Status SessionStatus `json:"status,omitempty"`
}

// SessionList contains a list of Session resources.
//
// This is the type returned by "kubectl get sessions" and used by the Kubernetes
// API when listing multiple Session resources.
//
// Example response:
//
//	apiVersion: stream.space/v1alpha1
//	kind: SessionList
//	metadata:
//	  resourceVersion: "123456"
//	items:
//	  - metadata:
//	      name: alice-firefox
//	    spec:
//	      user: alice
//	      template: firefox-browser
//	  - metadata:
//	      name: bob-vscode
//	    spec:
//	      user: bob
//	      template: vscode-dev
//
// +kubebuilder:object:root=true
type SessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Session `json:"items"`
}

// init registers the Session and SessionList types with the SchemeBuilder.
// This is called automatically when the package is imported and enables
// the controller-runtime to recognize these types.
func init() {
	SchemeBuilder.Register(&Session{}, &SessionList{})
}
