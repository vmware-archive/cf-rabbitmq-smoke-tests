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

const RETRY_LIMIT = 5
const COMMAND_TIMEOUT = 5 * time.Minute

func CfWithTimeout(timeout time.Duration, args ...string) *gexec.Session {
	session := cf.Cf(args...)

	select {
	case <-session.Exited:
	case <-time.After(timeout):
		session.Kill().Wait()
	}
	return session
}

func Cf(args ...string) *gexec.Session {
	var s *gexec.Session
	for i := 0; i < RETRY_LIMIT; i++ {
		s = CfWithTimeout(COMMAND_TIMEOUT, args...)
		if s.ExitCode() == 0 {
			return s
		}
		time.Sleep(5 * time.Second)
	}
	return s
}

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

	Eventually(Cf("create-security-group", securityGroupName, sgFile.Name()), ThirtySecondTimeout).Should(gexec.Exit(0))
	Eventually(Cf("bind-security-group", securityGroupName, orgName, spaceName), ThirtySecondTimeout).Should(gexec.Exit(0))
}

func DeleteSecurityGroup(securityGroupName string) {
	Eventually(Cf("delete-security-group", securityGroupName, "-f"), ThirtySecondTimeout).Should(gexec.Exit(0))
}

func GetAppEnv(appName string) string {
	appEnv := Cf("env", appName)
	Eventually(appEnv, ThirtySecondTimeout).Should(gexec.Exit(0))
	return string(appEnv.Buffer().Contents())
}

func PushAndBindApp(appName, serviceName, testAppPath string) string {
	Eventually(Cf("push", "-f", filepath.Join(testAppPath, "manifest.yml"), "--no-start", "--random-route", appName), FiveMinuteTimeout).Should(gexec.Exit(0))
	Eventually(Cf("bind-service", appName, serviceName), FiveMinuteTimeout).Should(gexec.Exit(0))
	Eventually(Cf("start", appName), FiveMinuteTimeout).Should(gexec.Exit(0))

	appDetails := Cf("app", appName)
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
	AwaitServiceAvailable(serviceName)
	return session
}

func UpdateService(serviceName, params string) {
	eventuallyWithTimeout(FiveMinuteTimeout, "update-service", serviceName, "-c", params)
	AwaitServiceAvailable(serviceName)
}

func UnbindService(appName, serviceName string) *gexec.Session {
	return eventuallyWithTimeout(FiveMinuteTimeout, "unbind-service", appName, serviceName)
}

func DeleteService(serviceName string) *gexec.Session {
	session := eventuallyWithTimeout(TenMinuteTimeout, "delete-service", serviceName, "-f")
	AwaitServiceDeletion(serviceName)
	return session
}

func CreateServiceKey(serviceName, keyName string) {
	eventuallyWithTimeout(FiveMinuteTimeout, "create-service-key", serviceName, keyName)
}

func DeleteServiceKey(serviceName, keyName string) {
	eventuallyWithTimeout(FiveMinuteTimeout, "delete-service-key", "-f", serviceName, keyName)
}

func GetServiceKey(serviceName, keyName string) []byte {
	getKey := Cf("service-key", serviceName, keyName)
	Eventually(getKey, FiveMinuteTimeout).Should(gexec.Exit(0))
	return getKey.Buffer().Contents()
}

func eventuallyWithTimeout(timeout time.Duration, args ...string) *gexec.Session {
	session := Cf(args...)
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
