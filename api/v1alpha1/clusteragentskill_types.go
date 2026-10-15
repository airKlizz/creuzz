package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterAgentSkillSpec defines the desired state of ClusterAgentSkill
type ClusterAgentSkillSpec struct {
	// FileName is the name of the skill file to be mounted in ~/.config/opencode/skills/
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+$`
	FileName string `json:"fileName"`

	// Content is the markdown content of the skill (context, system prompts, architecture guidelines)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Content string `json:"content"`
}

// ClusterAgentSkillStatus defines the observed state of ClusterAgentSkill.
type ClusterAgentSkillStatus struct {
	// conditions represent the current state of the ClusterAgentSkill resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="FileName",type=string,JSONPath=`.spec.fileName`

// ClusterAgentSkill is the Schema for the clusteragentskills API
type ClusterAgentSkill struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterAgentSkill
	// +required
	Spec ClusterAgentSkillSpec `json:"spec"`

	// status defines the observed state of ClusterAgentSkill
	// +optional
	Status ClusterAgentSkillStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAgentSkillList contains a list of ClusterAgentSkill
type ClusterAgentSkillList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAgentSkill `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAgentSkill{}, &ClusterAgentSkillList{})
		return nil
	})
}
