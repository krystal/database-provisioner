package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MySQLDatabaseSpec struct {
	ServerName                  string `json:"serverName"`
	ConnectionDetailsSecretName string `json:"connectionDetailsSecretName,omitempty"`
}

type MySQLDatabaseStatus struct {
	Created bool   `json:"created,omitempty"`
	Error   string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Server",type="string",JSONPath=".spec.serverName"
//+kubebuilder:printcolumn:name="Secret",type="string",JSONPath=".spec.connectionDetailsSecretName"
//+kubebuilder:printcolumn:name="Created",type="boolean",JSONPath=".status.created"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// MySQLDatabase is the Schema for the MySQLdatabases API
type MySQLDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLDatabaseSpec   `json:"spec,omitempty"`
	Status MySQLDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MySQLDatabaseList contains a list of MySQLDatabase
type MySQLDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQLDatabase{}, &MySQLDatabaseList{})
}
