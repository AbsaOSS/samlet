package controllers

import (
	samletv1 "github.com/bison-cloud-platform/samlet/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func needsUpdate(saml *samletv1.Saml2Aws) bool {
	now := &metav1.Time{Time: time.Now()}

	if saml.Status.ExpirationTime.Before(now) {
		return true
	}
	if saml.Status.RoleARN != saml.Spec.RoleARN {
		return true
	}
	return false
}
