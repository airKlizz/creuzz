package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterAgentMCPServerSpec defines the desired state of ClusterAgentMCPServer
type ClusterAgentMCPServerSpec struct {
	// Image is the OCI image reference for the MCP server sidecar container
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// Port is the port on which the MCP server listens (localhost within the pod)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// Args are the command-line arguments to pass to the MCP server container
	// +optional
	Args []string `json:"args,omitempty"`

	// EnvFromSecret is the name of a Kubernetes Secret in the target namespace
	// containing environment variables to inject into the MCP server container
	// +optional
	EnvFromSecret string `json:"envFromSecret,omitempty"`
}

// ClusterAgentMCPServerStatus defines the observed state of ClusterAgentMCPServer.
type ClusterAgentMCPServerStatus struct {
	// conditions represent the current state of the ClusterAgentMCPServer resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Port",type=integer,JSONPath=`.spec.port`

// ClusterAgentMCPServer is the Schema for the clusteragentmcpservers API
type ClusterAgentMCPServer struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterAgentMCPServer
	// +required
	Spec ClusterAgentMCPServerSpec `json:"spec"`

	// status defines the observed state of ClusterAgentMCPServer
	// +optional
	Status ClusterAgentMCPServerStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAgentMCPServerList contains a list of ClusterAgentMCPServer
type ClusterAgentMCPServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAgentMCPServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAgentMCPServer{}, &ClusterAgentMCPServerList{})
		return nil
	})
}
