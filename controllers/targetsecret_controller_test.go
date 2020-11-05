package controllers

import (
	"context"
	"time"

	samletv1 "github.com/bison-cloud-platform/samlet/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Samlet Controller", func() {
	const timeout = time.Second * 10
	const interval = time.Second * 1
	const awsEnvAccessKey = "AWS_SECRET_ACCESS_KEY"

	Context("Secrets", func() {
		var (
			samlMeta = types.NamespacedName{
				Name:      "test-saml",
				Namespace: "default",
			}
			targetSecretMeta = types.NamespacedName{
				Name:      "target-secret",
				Namespace: "default",
			}
			sourceSecretMeta = types.NamespacedName{
				Name:      "source-secret",
				Namespace: "default",
			}
			sourceSecret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sourceSecretMeta.Name,
					Namespace: sourceSecretMeta.Namespace,
				},
				StringData: map[string]string{
					"username": "foo",
					"password": "bar",
				},
			}
			targetSecret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      targetSecretMeta.Name,
					Namespace: targetSecretMeta.Namespace,
				},
				StringData: map[string]string{
					awsEnvAccessKey: "broken",
				},
			}
			obj = &samletv1.Saml2Aws{
				ObjectMeta: metav1.ObjectMeta{
					Name:      samlMeta.Name,
					Namespace: samlMeta.Namespace,
				},
				Spec: samletv1.Saml2AwsSpec{
					SecretFormat:     "envVariables",
					SecretName:       "source-secret",
					RoleARN:          "arn:aws:iam::000000000001:role/Production",
					TargetSecretName: "target-secret",
				},
			}
		)
		BeforeEach(func() {
			_ = k8sClient.Create(context.Background(), sourceSecret)
		})
		It("Creates target secret", func() {
			By("source secret should be created")
			Eventually(func() error {
				s := &v1.Secret{}
				return k8sClient.Get(context.Background(), sourceSecretMeta, s)
			}, timeout, interval).Should(Succeed())

			By("Creating test saml object")
			Expect(k8sClient.Create(context.Background(), obj)).Should(Succeed())

			By("Target secret is created")
			Eventually(func() error {
				s := &v1.Secret{}
				return k8sClient.Get(context.Background(), targetSecretMeta, s)
			}, timeout, interval).Should(Succeed())

			By("Secret Data matching expected result")
			s := &v1.Secret{}
			_ = k8sClient.Get(context.Background(), targetSecretMeta, s)
			Expect(string(s.Data[awsEnvAccessKey])).To(Equal("ACCESSECRETKEY"))
		})
		It("Updates expired secret", func() {
			// Break/Expire target secret
			_ = k8sClient.Update(context.Background(), targetSecret)

			By("Target secret is created")
			Eventually(func() error {
				s := &v1.Secret{}
				return k8sClient.Get(context.Background(), targetSecretMeta, s)
			}, timeout, interval).Should(Succeed())

			By("Updating status")
			updObj := &samletv1.Saml2Aws{}
			_ = k8sClient.Get(context.Background(), samlMeta, updObj)
			updObj.Status.ExpirationTime = metav1.Time{Time: time.Now().Add(time.Duration(-1) * time.Second)}
			Eventually(func() error {
				return k8sClient.Status().Update(context.Background(), updObj)
			}, timeout, interval).Should(Succeed())
			time.Sleep(time.Duration(10) * time.Second)

			By("Getting updated secret")
			s := &v1.Secret{}
			_ = k8sClient.Get(context.Background(), targetSecretMeta, s)
			Expect(string(s.Data[awsEnvAccessKey])).NotTo(Equal("broken"))

		})
	})
})
