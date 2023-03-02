package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MySQLServerSpec struct {
	Host     string `json:"host,omitempty"`
	Port     int32  `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type MySQLServerStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.host"
//+kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.port"
//+kubebuilder:printcolumn:name="Username",type="string",JSONPath=".spec.username"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type MySQLServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLServerSpec   `json:"spec,omitempty"`
	Status MySQLServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type MySQLServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQLServer{}, &MySQLServerList{})
}
