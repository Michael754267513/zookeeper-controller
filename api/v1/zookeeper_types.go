/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type ZookeeperMeta struct {
	Port int `json:"port"`
	Name string `json:"name"`
	Myid int `json:"myid"`
	ServiceType apiv1.ServiceType `json:"ServiceType,omitempty"`
	FlowerPort int `json:"flowerPort,omitempty"`
	LeaderPort int `json:"leaderPort,omitempty"`
}
// ZookeeperSpec defines the desired state of Zookeeper
type ZookeeperSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Zookeeper. Edit Zookeeper_types.go to remove/update
	Image string `json:"image"`
	NodeList  []ZookeeperMeta `json:"nodeList"`
	LogDir string `json:"logDir,omitempty"`
	DataDir string `json:"dataDir,omitempty"`
	NodeLogDir string `json:"nodeLogDir,omitempty"`
	NodeDataDir string `json:"nodeDataDir,omitempty"`
	CPU string `json:"CPU,omitempty"`
	Memory string `json:"Memory,omitempty"`

}
// ZookeeperStatus defines the observed state of Zookeeper
type ZookeeperStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Zookeeper is the Schema for the zookeepers API
type Zookeeper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZookeeperSpec   `json:"spec,omitempty"`
	Status ZookeeperStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ZookeeperList contains a list of Zookeeper
type ZookeeperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zookeeper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zookeeper{}, &ZookeeperList{})
}
