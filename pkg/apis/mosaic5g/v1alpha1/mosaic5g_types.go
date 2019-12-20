package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Mosaic5gSpec defines the desired state of Mosaic5g
// +k8s:openapi-gen=true
type Mosaic5gSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Size                   int32  `json:"size" `
	CNImage                string `json:"cnImage" yaml:"cnImage"`
	RANImage               string `json:"ranImage" yaml:"ranImage"`
	FlexRANImage           string `json:"flexRANImage" yaml:"flexRANImage"`
	MCC                    string `json:"mcc" yaml:"mcc"`
	MNC                    string `json:"mnc" yaml:"mnc"`
	EutraBand              string `json:"eutraBand" yaml:"eutraBand"`
	DownlinkFrequency      string `json:"downlinkFrequency" yaml:"downlinkFrequency"`
	UplinkFrequencyOffset  string `json:"uplinkFrequencyOffset" yaml:"uplinkFrequencyOffset"`
	FlexRAN                bool   `json:"flexRAN" yaml:"flexRAN"`
	Elasticsearch          bool   `json:"elasticsearch" yaml:"elasticsearch"`
	Kibana                 bool   `json:"kibana" yaml:"kibana"`
	DroneStore             bool   `json:"droneStore" yaml:"droneStore"`
	RRMKPIStore            bool   `json:"rrmkpiStore" yaml:"rrmkpiStore"`
	ConfigurationPathofCN  string `json:"configurationPathofCN" yaml:"configurationPathofCN"`
	ConfigurationPathofRAN string `json:"configurationPathofRAN" yaml:"configurationPathofRAN"`
	SnapBinaryPath         string `json:"snapBinaryPath" yaml:"snapBinaryPath"`
	DNS                    string `json:"dns" yaml:"dns"`
	HssDomainName          string `json:"hssDomainName" yaml:"hssDomainName"`
	MmeDomainName          string `json:"mmeDomainName" yaml:"mmeDomainName"`
	SpgwDomainName         string `json:"spgwDomainName" yaml:"spgwDomainName"`
	MysqlDomainName        string `json:"mysqlDomainName" yaml:"mysqlDomainName"`
	FlexRANDomainName      string `json:"flexRANDomainName" yaml:"flexRANDomainName"`
}

// Mosaic5gStatus defines the observed state of Mosaic5g
// +k8s:openapi-gen=true
type Mosaic5gStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mosaic5g is the Schema for the mosaic5gs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Mosaic5g struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Mosaic5gSpec   `json:"spec,omitempty"`
	Status Mosaic5gStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mosaic5gList contains a list of Mosaic5g
type Mosaic5gList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mosaic5g `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mosaic5g{}, &Mosaic5gList{})
}
