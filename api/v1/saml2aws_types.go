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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Saml2AwsSpec defines the desired state of Saml2Aws
type Saml2AwsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	SecretName       string `json:"secretName"`
	RoleARN          string `json:"roleARN"`
	SecretFormat     string `json:"secretFormat"`
	TargetSecretName string `json:"targetSecretName"`
	SessionDuration  string `json:"sessionDuration,omitempty"`
	IDPEndpoint      string `json:"idpEndpoint,omitempty"`
}

// Saml2AwsStatus defines the observed state of Saml2Aws
type Saml2AwsStatus struct {
	RoleARN        string      `json:"roleARN,omitempty"`
	ExpirationTime metav1.Time `json:"expirationTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Saml2Aws is the Schema for the saml2aws API
type Saml2Aws struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Saml2AwsSpec   `json:"spec,omitempty"`
	Status Saml2AwsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// Saml2AwsList contains a list of Saml2Aws
type Saml2AwsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Saml2Aws `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Saml2Aws{}, &Saml2AwsList{})
}
