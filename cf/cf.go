package cf

import (
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
	eventuallyWithTimeout(cf_helpers.FiveMinuteTimeout, "delete-service", serviceName, "-f")
	AssertProgress(serviceName, "delete")

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
