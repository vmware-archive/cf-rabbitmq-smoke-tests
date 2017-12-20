package cf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	helpersCF "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/lifecycle_tests/cf_helpers"
)

func Api(endpoint string, skipSSLValidation bool) *gexec.Session {
	apiCmd := []string{"api", endpoint}

	if skipSSLValidation {
		apiCmd = append(apiCmd, "--skip-ssl-validation")
	}

	return eventually(apiCmd...)
}

func Auth(username, password string) *gexec.Session {
	return eventually("auth", username, password)
}

func Target(orgName, spaceName string) *gexec.Session {
	return eventually("target", "-o", orgName, "-s", spaceName)
}

func CreateAndBindSecurityGroup(securityGroupName, orgName, spaceName string) *gexec.Session {
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

	eventually("create-security-group", securityGroupName, sgFile.Name())
	return eventually("bind-security-group", securityGroupName, orgName, spaceName)
}

func CreateAndSetQuota(quotaName, orgName string) *gexec.Session {
	eventually("create-quota", quotaName, "-m", "10G", "-r", "1000", "-s", "100", "--allow-paid-service-plans")

	return eventually("set-quota", orgName, quotaName)
}

func EnableServiceAccess(serviceOffering, testPlan, orgName string) *gexec.Session {
	return eventually("enable-service-access", serviceOffering, "-p", testPlan, "-o", orgName)
}

func CreateOrg(orgName string) *gexec.Session {
	return eventually("create-org", orgName)
}

func CreateSpace(orgName, spaceName string) *gexec.Session {
	return eventually("create-space", spaceName, "-o", orgName)
}

func DeleteOrg(orgName string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.ThirtySecondTimeout, "delete-org", orgName, "-f")
}

func DeleteSpace(spaceName string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.ThirtySecondTimeout, "delete-space", spaceName, "-f")
}

func DeleteApp(appName string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.ThirtySecondTimeout, "delete", appName, "-f", "-r")
}

func DeleteSecurityGroup(securityGroupName string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.ThirtySecondTimeout, "delete-security-group", securityGroupName, "-f")
}

func DeleteQuota(quotaName string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.ThirtySecondTimeout, "delete-quota", quotaName, "-f")
}

func CreateService(serviceOffering, servicePlan, serviceName, arbitraryParams string) *gexec.Session {
	args := []string{"create-service", serviceOffering, servicePlan, serviceName}
	if arbitraryParams != "" {
		args = append(args, "-c", arbitraryParams)
	}

	session := eventuallyWithTimeout(cf_helpers.FiveMinuteTimeout, args...)
	cf_helpers.AwaitServiceCreation(serviceName)
	return session
}

func UpdateService(serviceName, planName string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.FiveMinuteTimeout, "update-service", serviceName, "-p", planName)
}

func UnbindService(appName, serviceName string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.FiveMinuteTimeout, "unbind-service", appName, serviceName)
}

func AssertProgress(serviceName, operation string) {
	session := eventuallyWithTimeout(cf_helpers.FiveMinuteTimeout, "service", serviceName)
	Eventually(session).Should(gbytes.Say(operation + " in progress"))
}

func DeleteService(serviceName string) {
	eventuallyWithTimeout(cf_helpers.TenMinuteTimeout, "delete-service", serviceName, "-f")

	cf_helpers.AwaitServiceDeletion(serviceName)
}

func eventually(args ...string) *gexec.Session {
	return eventuallyWithTimeout(cf_helpers.FiveSecondTimeout, args...)
}

func eventuallyWithTimeout(timeout time.Duration, args ...string) *gexec.Session {
	session := helpersCF.Cf(args...)
	Eventually(session, timeout).Should(gexec.Exit(0))
	return session
}
