package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EgressPort defines a port and protocol for egress rules
type EgressPort struct {
	// Protocol is the network protocol (TCP, UDP, SCTP)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	// +kubebuilder:default=TCP
	Protocol string `json:"protocol,omitempty"`

	// Port is the port number
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// ClusterAgentEgressRuleSpec defines the desired state of ClusterAgentEgressRule
type ClusterAgentEgressRuleSpec struct {
	// Description is a human-readable description of the egress rule
	// +optional
	Description string `json:"description,omitempty"`

	// Domains is a list of fully qualified domain names (FQDN) allowed for egress
	// +optional
	// +listType=set
	Domains []string `json:"domains,omitempty"`

	// Ports is a list of ports and protocols allowed for egress
	// +optional
	// +listType=map
	// +listMapKey=port
	Ports []EgressPort `json:"ports,omitempty"`
}

// ClusterAgentEgressRuleStatus defines the observed state of ClusterAgentEgressRule.
type ClusterAgentEgressRuleStatus struct {
	// conditions represent the current state of the ClusterAgentEgressRule resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.spec.description`

// ClusterAgentEgressRule is the Schema for the clusteragentegressrules API
type ClusterAgentEgressRule struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterAgentEgressRule
	// +required
	Spec ClusterAgentEgressRuleSpec `json:"spec"`

	// status defines the observed state of ClusterAgentEgressRule
	// +optional
	Status ClusterAgentEgressRuleStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAgentEgressRuleList contains a list of ClusterAgentEgressRule
type ClusterAgentEgressRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAgentEgressRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAgentEgressRule{}, &ClusterAgentEgressRuleList{})
		return nil
	})
}
