package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ProjectSpec defines the git repository to clone for the agent workspace
type ProjectSpec struct {
	// RepoURL is the git repository URL (SSH or HTTPS)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	RepoURL string `json:"repoUrl"`

	// Branch is the git branch or ref to checkout
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Branch string `json:"branch"`

	// SecretRef is the name of a Kubernetes Secret in the agent namespace
	// containing `id_ed25519` and `known_hosts` for SSH authentication
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	SecretRef string `json:"secretRef"`
}

// RuntimeSpec defines the compute resources and container image for the agent
type RuntimeSpec struct {
	// Image is the OpenCode v2 base container image
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// CPU is the CPU request/limit (e.g., "2", "500m")
	// +kubebuilder:validation:Required
	CPU string `json:"cpu"`

	// Memory is the memory request/limit (e.g., "4Gi", "512Mi")
	// +kubebuilder:validation:Required
	Memory string `json:"memory"`

	// Storage is the size of the workspace PVC (e.g., "10Gi")
	// +kubebuilder:validation:Required
	Storage string `json:"storage"`
}

// ClusterResourceRef is a reference to a cluster-scoped resource by name
type ClusterResourceRef struct {
	// Name is the name of the cluster-scoped resource
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// OpenCodeAgentSpec defines the desired state of OpenCodeAgent
type OpenCodeAgentSpec struct {
	// Project defines the git repository to clone into the agent workspace
	// +required
	Project ProjectSpec `json:"project"`

	// Runtime defines the compute resources and container image
	// +required
	Runtime RuntimeSpec `json:"runtime"`

	// ClusterToolRefs is a list of ClusterAgentTool names to mount as OCI image volumes
	// +optional
	ClusterToolRefs []ClusterResourceRef `json:"clusterToolRefs,omitempty"`

	// ClusterSkillRefs is a list of ClusterAgentSkill names to inject as configuration files
	// +optional
	ClusterSkillRefs []ClusterResourceRef `json:"clusterSkillRefs,omitempty"`

	// ClusterMcpRefs is a list of ClusterAgentMCPServer names to run as sidecar containers
	// +optional
	ClusterMcpRefs []ClusterResourceRef `json:"clusterMcpRefs,omitempty"`

	// ClusterEnvRefs is a list of ClusterAgentEnv names to inject environment variables
	// +optional
	ClusterEnvRefs []ClusterResourceRef `json:"clusterEnvRefs,omitempty"`

	// ClusterEgressRefs is a list of ClusterAgentEgressRule names to allow network egress
	// +optional
	ClusterEgressRefs []ClusterResourceRef `json:"clusterEgressRefs,omitempty"`

	// TTL AlleIdle is the duration of inactivity after which the agent scales to zero (e.g., "2h")
	// +optional
	// +kubebuilder:validation:Pattern=`^([0-9]+(ns|us|µs|ms|s|m|h))+$`
	TTLIdleAfter string `json:"ttlAfterIdle,omitempty"`
}

// AgentPhase represents the lifecycle phase of an OpenCodeAgent
// +kubebuilder:validation:Enum=Pending;Ready;Failed;Suspended
type AgentPhase string

const (
	AgentPhasePending   AgentPhase = "Pending"
	AgentPhaseReady     AgentPhase = "Ready"
	AgentPhaseFailed    AgentPhase = "Failed"
	AgentPhaseSuspended AgentPhase = "Suspended"
)

// OpenCodeAgentStatus defines the observed state of OpenCodeAgent.
type OpenCodeAgentStatus struct {
	// Phase is the current lifecycle phase of the agent
	// +optional
	Phase AgentPhase `json:"phase,omitempty"`

	// URL is the HTTPS URL to access the OpenCode v2 API
	// +optional
	URL string `json:"url,omitempty"`

	// PairingTokenSecret is the name of the Secret containing the pairing token
	// +optional
	PairingTokenSecret string `json:"pairingTokenSecret,omitempty"`

	// PodName is the name of the generated Pod
	// +optional
	PodName string `json:"podName,omitempty"`

	// ServiceName is the name of the generated Service
	// +optional
	ServiceName string `json:"serviceName,omitempty"`

	// conditions represent the current state of the OpenCodeAgent resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// OpenCodeAgent is the Schema for the opencodeagents API
type OpenCodeAgent struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of OpenCodeAgent
	// +required
	Spec OpenCodeAgentSpec `json:"spec"`

	// status defines the observed state of OpenCodeAgent
	// +optional
	Status OpenCodeAgentStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// OpenCodeAgentList contains a list of OpenCodeAgent
type OpenCodeAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []OpenCodeAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &OpenCodeAgent{}, &OpenCodeAgentList{})
		return nil
	})
}
