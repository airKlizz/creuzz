package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterAgentToolSpec defines the desired state of ClusterAgentTool
type ClusterAgentToolSpec struct {
	// Image is the OCI image reference for the tool (Kubernetes 1.31+ image volume)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// MountPath is the path where the tool image will be mounted in the agent pod
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	MountPath string `json:"mountPath"`

	// BinPath is the directory containing executables to add to the agent's PATH
	// +optional
	BinPath string `json:"binPath,omitempty"`
}

// ClusterAgentToolStatus defines the observed state of ClusterAgentTool.
type ClusterAgentToolStatus struct {
	// conditions represent the current state of the ClusterAgentTool resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="MountPath",type=string,JSONPath=`.spec.mountPath`

// ClusterAgentTool is the Schema for the clusteragenttools API
type ClusterAgentTool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterAgentTool
	// +required
	Spec ClusterAgentToolSpec `json:"spec"`

	// status defines the observed state of ClusterAgentTool
	// +optional
	Status ClusterAgentToolStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAgentToolList contains a list of ClusterAgentTool
type ClusterAgentToolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAgentTool `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAgentTool{}, &ClusterAgentToolList{})
		return nil
	})
}
