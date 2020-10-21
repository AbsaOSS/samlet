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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	samletv1 "github.com/bison-cloud-platform/samlet/api/v1"
)

// Saml2AwsReconciler reconciles a Saml2Aws object
type Saml2AwsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var log = logf.Log.WithName("controller_saml")

// +kubebuilder:rbac:groups=samlet.absa.oss,resources=saml2aws,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=samlet.absa.oss,resources=saml2aws/status,verbs=get;update;patch

// Reconcile reconcile loop handler
func (r *Saml2AwsReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log = r.Log.WithValues("saml2aws", req.NamespacedName)

	var result *ctrl.Result

	saml := &samletv1.Saml2Aws{}
	err := r.Get(context.TODO(), req.NamespacedName, saml)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	result, err = r.createAWSCreds(req, saml)
	if result != nil {
		return *result, err
	}

	// your logic here
	return ctrl.Result{}, nil
}

// SetupWithManager sets up controller with manager
func (r *Saml2AwsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samletv1.Saml2Aws{}).
		Complete(r)
}
