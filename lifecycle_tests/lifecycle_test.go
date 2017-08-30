package lifecycle_tests

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/cf"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/lifecycle_tests/cf_helpers"
)

var _ = Describe("The service broker lifecycle", func() {
	var (
		appName     string
		appPath     string
		serviceName string
	)

	BeforeEach(func() {
		appName = fmt.Sprintf("rmq-smoke-test-app-%d", GinkgoParallelNode())
		appPath = "../assets/rabbit-example-app"
		serviceName = fmt.Sprintf("rmq-smoke-test-instance-%s", uuid.New()[:18])
	})

	AfterEach(func() {
		cf.DeleteApp(appName)
	})

	lifecycle := func(t TestPlan) {
		It(fmt.Sprintf("plan: '%s', with arbitrary params: '%s', will update to: '%s'", t.Name, string(t.ArbitraryParams), t.UpdateToPlan), func() {
			cf.CreateService(rabbitmqConfig.ServiceOffering, t.Name, serviceName, string(t.ArbitraryParams))

			appURL := cf_helpers.PushAndBindApp(appName, serviceName, appPath)

			testService(rabbitmqConfig.AppType, appURL)

			if t.UpdateToPlan != "" {
				updatePlan(serviceName, t.UpdateToPlan)
				testService(rabbitmqConfig.AppType, appURL)
			}

			cf.UnbindService(appName, serviceName)

			cf.DeleteService(serviceName)
		})
	}

	for _, plan := range rabbitmqConfig.TestPlans {
		lifecycle(plan)
	}
})

func testService(exampleAppType, testAppURL string) {
	switch exampleAppType {
	case "crud":
		testCrud(testAppURL)
	case "fifo":
		testFifo(testAppURL)
	default:
		Fail(fmt.Sprintf("invalid example app type %s. valid types are: crud, fifo", exampleAppType))
	}
}

func testCrud(testAppURL string) {
	cf_helpers.PutToTestApp(testAppURL, "foo", "bar")
	Expect(cf_helpers.GetFromTestApp(testAppURL, "foo")).To(Equal("bar"))
}

func testFifo(testAppURL string) {
	queue := "a-test-queue"
	cf_helpers.PushToTestAppQueue(testAppURL, queue, "foo")
	cf_helpers.PushToTestAppQueue(testAppURL, queue, "bar")
	Expect(cf_helpers.PopFromTestAppQueue(testAppURL, queue)).To(Equal("foo"))
	Expect(cf_helpers.PopFromTestAppQueue(testAppURL, queue)).To(Equal("bar"))
}

func updatePlan(serviceName, updatedPlanName string) {
	cf.UpdateService(serviceName, updatedPlanName)
	cf.AssertProgress(serviceName, "update")
	cf_helpers.AwaitServiceUpdate(serviceName)
}
