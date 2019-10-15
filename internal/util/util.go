package util

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

// LabelsForMosaic5g returns the labels for selecting the resources
// belonging to the given Mosaic5g CR name.
func LabelsForMosaic5g(name string) map[string]string {
	return map[string]string{"app": "mosaic5g", "mosaic5g_cr": name}
}

// GetPodNames returns the pod names of the array of pods passed in
func GetPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// NewTrue returns a bool pointer
func NewTrue() *bool {
	b := true
	return &b
}

// NewHostPathType returns a pointer of NewHostPathType
func NewHostPathType(input string) *corev1.HostPathType {
	var a corev1.HostPathType
	a = corev1.HostPathType(input)

	return &a
}

// GenAffinity will return a v1.Affinity pointer with specified tag value
func GenAffinity(tagValue string) *v1.Affinity {
	affinity := v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{{
				Weight: 10,
				Preference: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "oai",
						Operator: v1.NodeSelectorOperator("In"),
						Values:   []string{tagValue},
					}},
				}},
			},
		},
	}
	return &affinity
}
