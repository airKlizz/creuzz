package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EnvVariable defines a static environment variable
type EnvVariable struct {
	// Name is the environment variable name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Value is the environment variable value
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// SecretEnvVariable defines an environment variable sourced from a Kubernetes Secret
type SecretEnvVariable struct {
	// Name is the environment variable name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// SecretRef is the name of the Kubernetes Secret in the target namespace
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	SecretRef string `json:"secretRef"`

	// Key is the key within the Secret to use as the value
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
}

// ClusterAgentEnvSpec defines the desired state of ClusterAgentEnv
type ClusterAgentEnvSpec struct {
	// Variables is a list of static environment variables to inject
	// +optional
	// +listType=map
	// +listMapKey=name
	Variables []EnvVariable `json:"variables,omitempty"`

	// Secrets is a list of environment variables sourced from Kubernetes Secrets
	// +optional
	// +listType=map
	// +listMapKey=name
	Secrets []SecretEnvVariable `json:"secrets,omitempty"`
}

// ClusterAgentEnvStatus defines the observed state of ClusterAgentEnv.
type ClusterAgentEnvStatus struct {
	// conditions represent the current state of the ClusterAgentEnv resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Variables",type=integer,JSONPath=`.spec.variables[*].name`

// ClusterAgentEnv is the Schema for the clusteragentenvs API
type ClusterAgentEnv struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterAgentEnv
	// +required
	Spec ClusterAgentEnvSpec `json:"spec"`

	// status defines the observed state of ClusterAgentEnv
	// +optional
	Status ClusterAgentEnvStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAgentEnvList contains a list of ClusterAgentEnv
type ClusterAgentEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAgentEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAgentEnv{}, &ClusterAgentEnvList{})
		return nil
	})
}
