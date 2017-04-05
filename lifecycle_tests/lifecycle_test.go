package lifecycle_tests

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/lifecycle_tests/cf_helpers"
)

var _ = Describe("The service broker lifecycle", func() {
	AfterEach(func() {
		Eventually(cf.Cf("delete", testAppName, "-f", "-r"), cf_helpers.ThirtySecondTimeout).Should(gexec.Exit())
	})

	newServiceName := func() string {
		return fmt.Sprintf("instance-%s", uuid.New()[:18])
	}

	createService := func(servicePlan, serviceName, arbitraryParams string) {
		cfArgs := []string{"create-service", serviceOffering, servicePlan, serviceName}
		if arbitraryParams != "" {
			cfArgs = append(cfArgs, "-c", arbitraryParams)
		}

		Eventually(cf.Cf(cfArgs...), cf_helpers.FiveMinuteTimeout).Should(gexec.Exit(0))
		cf_helpers.AwaitServiceCreation(serviceName)
	}

	unbindService := func(serviceName string) {
		Eventually(cf.Cf("unbind-service", testAppName, serviceName), cf_helpers.FiveMinuteTimeout).Should(gexec.Exit(0))
	}

	assertProgress := func(serviceName, operation string) {
		session := cf.Cf("service", serviceName)
		Eventually(session, cf_helpers.FiveMinuteTimeout).Should(gbytes.Say(operation + " in progress"))
		Eventually(session).Should(gexec.Exit(0))
	}

	deleteService := func(serviceName string) {
		Eventually(cf.Cf("delete-service", serviceName, "-f"), cf_helpers.FiveMinuteTimeout).Should(gexec.Exit(0))
		assertProgress(serviceName, "delete")
		cf_helpers.AwaitServiceDeletion(serviceName)
	}

	testCrud := func(testAppURL string) {
		cf_helpers.PutToTestApp(testAppURL, "foo", "bar")
		Expect(cf_helpers.GetFromTestApp(testAppURL, "foo")).To(Equal("bar"))
	}

	testFifo := func(testAppURL string) {
		queue := "a-test-queue"
		cf_helpers.PushToTestAppQueue(testAppURL, queue, "foo")
		cf_helpers.PushToTestAppQueue(testAppURL, queue, "bar")
		Expect(cf_helpers.PopFromTestAppQueue(testAppURL, queue)).To(Equal("foo"))
		Expect(cf_helpers.PopFromTestAppQueue(testAppURL, queue)).To(Equal("bar"))
	}

	updatePlan := func(serviceName, updatedPlanName string) {
		Eventually(cf.Cf("update-service", serviceName, "-p", updatedPlanName), cf_helpers.FiveMinuteTimeout).Should(gexec.Exit(0))
		assertProgress(serviceName, "update")
		cf_helpers.AwaitServiceUpdate(serviceName)
	}

	testServiceWithExampleApp := func(exampleAppType, testAppURL string) {
		switch exampleAppType {
		case "crud":
			testCrud(testAppURL)
		case "fifo":
			testFifo(testAppURL)
		default:
			Fail(fmt.Sprintf("invalid example app type %s. valid types are: crud, fifo", exampleAppType))
		}
	}

	lifecycle := func(t LifecycleTest) {
		It(fmt.Sprintf("plan: '%s', with arbitrary params: '%s', will update to: '%s'", t.Plan, string(t.ArbitraryParams), t.UpdateToPlan), func() {
			serviceName := newServiceName()
			createService(t.Plan, serviceName, string(t.ArbitraryParams))
			testAppURL := cf_helpers.PushAndBindApp(testAppName, serviceName, exampleAppPath)

			testServiceWithExampleApp(exampleAppType, testAppURL)

			if t.UpdateToPlan != "" {
				updatePlan(serviceName, t.UpdateToPlan)
				testServiceWithExampleApp(exampleAppType, testAppURL)
			}

			unbindService(serviceName)
			deleteService(serviceName)
		})
	}

	for _, test := range tests {
		lifecycle(test)
	}
})

type GinkgoFirehosePrinter struct{}

func (c GinkgoFirehosePrinter) Print(title, dump string) {
	fmt.Fprintf(GinkgoWriter, "firehose: %s\n---%s\n---\n", title, dump)
}
