package helper

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/craigfurman/herottp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func PushAndBindApp(appName, serviceName, testAppPath string) string {
	Eventually(cf.Cf("push", "-f", filepath.Join(testAppPath, "manifest.yml"), "--no-start", appName), FiveMinuteTimeout).Should(gexec.Exit(0))
	Eventually(cf.Cf("bind-service", appName, serviceName), FiveMinuteTimeout).Should(gexec.Exit(0))
	Eventually(cf.Cf("start", appName), FiveMinuteTimeout).Should(gexec.Exit(0))

	appDetails := cf.Cf("app", appName)
	Eventually(appDetails, FiveSecondTimeout).Should(gexec.Exit(0))

	appDetailsOutput := string(appDetails.Buffer().Contents())
	testAppURL := findURL(appDetailsOutput)
	Expect(testAppURL).NotTo(BeEmpty())

	return testAppURL
}

func DeleteApp(appName string) *gexec.Session {
	return eventuallyWithTimeout(ThirtySecondTimeout, "delete", appName, "-f", "-r")
}

func CreateService(serviceOffering, servicePlan, serviceName string) *gexec.Session {
	session := eventuallyWithTimeout(FiveMinuteTimeout, "create-service", serviceOffering, servicePlan, serviceName)
	AwaitServiceCreation(serviceName)
	return session
}

func UnbindService(appName, serviceName string) *gexec.Session {
	return eventuallyWithTimeout(FiveMinuteTimeout, "unbind-service", appName, serviceName)
}

func DeleteService(serviceName string) *gexec.Session {
	session := eventuallyWithTimeout(TenMinuteTimeout, "delete-service", serviceName, "-f")
	AwaitServiceDeletion(serviceName)
	return session
}

func eventuallyWithTimeout(timeout time.Duration, args ...string) *gexec.Session {
	session := cf.Cf(args...)
	Eventually(session, timeout).Should(gexec.Exit(0))
	return session
}

func SendMessage(testAppURL, queueName, message string) {
	postReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s/queues/%s", testAppURL, queueName),
		strings.NewReader(message),
	)
	postReq.Header.Add("Content-Type", "text/plain")
	Expect(err).ToNot(HaveOccurred())
	makeAndCheckHttpRequest(postReq)
}

func ReceiveMessage(testAppURL, queueName string) string {
	getReq, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/queues/%s", testAppURL, queueName),
		nil,
	)
	getReq.Header.Add("Content-Type", "text/plain")
	Expect(err).ToNot(HaveOccurred())
	return makeAndCheckHttpRequest(getReq)
}

func makeAndCheckHttpRequest(req *http.Request) string {
	certIgnoringHTTPClient := herottp.New(herottp.Config{
		DisableTLSCertificateVerification: true,
		Timeout: ThirtySecondTimeout,
	})

	resp, err := certIgnoringHTTPClient.Do(req)
	defer resp.Body.Close()
	Expect(err).ToNot(HaveOccurred())

	bodyContent, err := ioutil.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())

	fmt.Fprintf(GinkgoWriter,
		"response from %s %s: %d\n------------------------------\n%s\n------------------------------\n",
		req.Method,
		req.URL.String(),
		resp.StatusCode,
		bodyContent,
	)
	Expect(resp.StatusCode).To(BeNumerically("<", 300))

	return string(bodyContent)
}

func findURL(cliOutput string) string {
	for _, line := range strings.Split(cliOutput, "\n") {
		if strings.HasPrefix(line, "routes:") {
			return strings.Fields(line)[1]
		}
	}
	return ""
}
