package v1alpha1

type Status struct {
	Ready              bool   `json:"ready"`
	ObservedGeneration int64  `json:"observedGeneration,omitempty"`
	Message            string `json:"message,omitempty"`
}
