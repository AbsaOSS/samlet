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
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	samletv1 "github.com/bison-cloud-platform/samlet/api/v1"
	configreader "github.com/bison-cloud-platform/samlet/controllers/config"
	// +kubebuilder:scaffold:imports
)

const (
	idpEndpointKey = "IDP_ENDPOINT"
	awsEndpointKey = "AWS_ENDPOINT"
	awsRegionKey   = "AWS_REGION"
	durationKey    = "DurationSeconds"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var config *rest.Config
var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment
var err error

func returnAssertion(w http.ResponseWriter, req *http.Request) {
	rawData, _ := ioutil.ReadFile("../testdata/assertion.xml")
	b64Data := base64.StdEncoding.EncodeToString([]byte(rawData))
	tpl := template.Must(template.ParseFiles("../testdata/samlResponse.tmpl"))
	_ = tpl.Execute(w, b64Data)
}
func returnLoginPage(w http.ResponseWriter, req *http.Request) {
	data, _ := ioutil.ReadFile("../testdata/loginpage.html")
	_, _ = w.Write(data)
}
func returnSamlPage(w http.ResponseWriter, req *http.Request) {
	data, _ := ioutil.ReadFile("../testdata/saml.html")
	_, _ = w.Write(data)
}
func returnAWSCreds(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()
	duration, _ := strconv.Atoi(req.Form.Get(durationKey))
	expireTime := time.Now().UTC().Add(time.Duration(duration) * time.Second)
	w.Header().Set("Content-Type", "text/xml")
	tpl := template.Must(template.ParseFiles("../testdata/awsResponse.tmpl"))
	_ = tpl.Execute(w, expireTime.Format(time.RFC3339))
}

func startHttp() {
	http.HandleFunc("/adfs/ls/idpinitiatedsignon", returnAssertion)
	http.HandleFunc("/adfs/", returnLoginPage)
	http.HandleFunc("/saml/", returnSamlPage)
	http.HandleFunc("/aws/", returnAWSCreds)
	_ = http.ListenAndServe(":3000", nil)
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	go startHttp()
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	config, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(config).ToNot(BeNil())

	err = samletv1.AddToScheme(scheme.Scheme)
	Expect(err).ToNot(HaveOccurred())

	// +kubebuilder:scaffold:scheme
	k8sManager, err = ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = os.Setenv(idpEndpointKey, "http://localhost:3000")
	Expect(err).ToNot(HaveOccurred())

	err = os.Setenv(awsEndpointKey, "http://localhost:3000/aws")
	Expect(err).ToNot(HaveOccurred())

	// when setting aws endpoint region becomes mandatory
	err = os.Setenv(awsRegionKey, "us-west-1")
	Expect(err).ToNot(HaveOccurred())

	ctrlConf, _ := configreader.GetConfig()
	err = (&Saml2AwsReconciler{
		Client: k8sManager.GetClient(),
		Log:    ctrl.Log.WithName("controller_saml"),
		Config: ctrlConf,
		Scheme: scheme.Scheme,
	}).SetupWithManager(k8sManager)
	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()
	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
