package helper

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func CreateAndBindSecurityGroup(securityGroupName, orgName, spaceName string) {
	sgs := []struct {
		Protocol    string `json:"protocol"`
		Destination string `json:"destination"`
		Ports       string `json:"ports"`
	}{
		{"tcp", "0.0.0.0/0", "5671,5672,1883,8883,61613,61614,15672,15674"},
	}

	sgFile, err := ioutil.TempFile("", "smoke-test-security-group-")
	Expect(err).NotTo(HaveOccurred())
	defer sgFile.Close()
	defer os.Remove(sgFile.Name())

	err = json.NewEncoder(sgFile).Encode(sgs)
	Expect(err).NotTo(HaveOccurred(), `{"FailReason": "Failed to encode security groups"}`)

	Eventually(cf.Cf("create-security-group", securityGroupName, sgFile.Name()), ThirtySecondTimeout).Should(gexec.Exit(0))
	Eventually(cf.Cf("bind-security-group", securityGroupName, orgName, spaceName), ThirtySecondTimeout).Should(gexec.Exit(0))
}

func DeleteSecurityGroup(securityGroupName string) {
	Eventually(cf.Cf("delete-security-group", securityGroupName, "-f"), ThirtySecondTimeout).Should(gexec.Exit(0))
}

func PushAndBindApp(appName, serviceName, testAppPath string) string {
	Eventually(cf.Cf("push", "-f", filepath.Join(testAppPath, "manifest.yml"), "--no-start", "--random-route", appName), FiveMinuteTimeout).Should(gexec.Exit(0))
	Eventually(cf.Cf("bind-service", appName, serviceName), FiveMinuteTimeout).Should(gexec.Exit(0))
	Eventually(cf.Cf("start", appName), FiveMinuteTimeout).Should(gexec.Exit(0))

	appDetails := cf.Cf("app", appName)
	Eventually(appDetails, ThirtySecondTimeout).Should(gexec.Exit(0))

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

func findURL(cliOutput string) string {
	for _, line := range strings.Split(cliOutput, "\n") {
		if strings.HasPrefix(line, "routes:") {
			return strings.Fields(line)[1]
		}
	}
	return ""
}