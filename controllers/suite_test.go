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
	idpEndpointKey     = "IDP_ENDPOINT"
	awsEndpointKey     = "AWS_ENDPOINT"
	awsRegionKey       = "AWS_REGION"
	durationKey        = "DurationSeconds"
	sessionDurationKey = "SESSION_DURATION"
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
	err := tpl.Execute(w, b64Data)
	if err != nil {
		Fail("Fail to process samlResponse template")
	}
}
func returnLoginPage(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadFile("../testdata/loginpage.html")
	if err != nil {
		Fail("Fail to read loginpage.html")
	}
	_, err = w.Write(data)
	if err != nil {
		Fail("Fail to write response")
	}
}
func returnSamlPage(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadFile("../testdata/saml.html")
	if err != nil {
		Fail("Fail to read saml.html")
	}
	_, err = w.Write(data)
	if err != nil {
		Fail("Fail to write response")
	}
}
func returnAWSCreds(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		Fail("Fail to parse request")
	}
	duration, err := strconv.Atoi(req.Form.Get(durationKey))
	if err != nil {
		Fail("Fail to get DurationSeconds")
	}
	expireTime := time.Now().UTC().Add(time.Duration(duration) * time.Second)
	w.Header().Set("Content-Type", "text/xml")
	tpl := template.Must(template.ParseFiles("../testdata/awsResponse.tmpl"))
	err = tpl.Execute(w, expireTime.Format(time.RFC3339))
	if err != nil {
		Fail("Fail to process awsResponse template")
	}
}

func startHttp() {
	http.HandleFunc("/adfs/ls/idpinitiatedsignon", returnAssertion)
	http.HandleFunc("/adfs/", returnLoginPage)
	http.HandleFunc("/saml/", returnSamlPage)
	http.HandleFunc("/aws/", returnAWSCreds)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		Fail("Fail to start http server on port 3000")
	}
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

	err = os.Setenv(sessionDurationKey, "1h")
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