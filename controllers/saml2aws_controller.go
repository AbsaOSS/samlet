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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

const requeueTime = 10

var log = logf.Log.WithName("controller_saml")

// +kubebuilder:rbac:groups=samlet.absa.oss,resources=saml2aws,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=samlet.absa.oss,resources=saml2aws/status,verbs=get;update;patch

// Reconcile reconcile loop handler
func (r *Saml2AwsReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log = r.Log.WithValues("saml2aws", req.NamespacedName)

	saml := &samletv1.Saml2Aws{}
	err := r.Get(ctx, req.NamespacedName, saml)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if needsUpdate(saml) {
		creds, profile, err := r.createAWSCreds(saml)
		if err != nil {
			return ctrl.Result{}, err
		}
		secret, err := r.targetSecret(saml)
		if err != nil {
			return ctrl.Result{}, err
		}
		switch f := saml.Spec.SecretFormat; f {
		case "credentialsFile":
			secret, err = generateCredentiasFile(profile, creds, secret)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to generate ini file")
			}
		case "envVariables":
			secret = generateEnvVariables(creds, secret)
		default:
			return ctrl.Result{}, fmt.Errorf("invalid secret format")

		}

		err = r.updateSecret(saml.Spec.TargetSecretName, saml.Namespace, secret)
		if err != nil {
			return ctrl.Result{}, err
		}

		saml.Status.ExpirationTime = metav1.Time{Time: creds.Expires}
		saml.Status.RoleARN = saml.Spec.RoleARN
		err = r.Status().Update(ctx, saml)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	reconcileTime := saml.Status.ExpirationTime.Add(time.Duration(-requeueTime) * time.Minute)
	log.Info("Reconcile", "Requeue at", reconcileTime)
	return ctrl.Result{RequeueAfter: time.Until(reconcileTime)}, nil
}

// SetupWithManager sets up controller with manager
func (r *Saml2AwsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samletv1.Saml2Aws{}).
		Complete(r)
}
