package lifecycle_tests

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/cf"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/lifecycle_tests/cf_helpers"
)

var _ = Describe("The service broker lifecycle", func() {
	var (
		apps        map[string]string
		serviceName string
	)

	BeforeEach(func() {
		apps = map[string]string{
			"rmq-smoke-tests-ruby":   "../assets/rabbit-example-app",
			"rmq-smoke-tests-spring": "../assets/spring-example-app",
		}
		serviceName = fmt.Sprintf("rmq-smoke-test-instance-%s", uuid.New()[:18])
	})

	AfterEach(func() {
		for appName, _ := range apps {
			cf.DeleteApp(appName)
		}
	})

	lifecycle := func(testPlan TestPlan) {
		It(fmt.Sprintf("plan: '%s', with arbitrary params: '%s', will update to: '%s'", testPlan.Name, string(testPlan.ArbitraryParams), testPlan.UpdateToPlan), func() {
			cf.CreateService(config.ServiceOffering, testPlan.Name, serviceName, string(testPlan.ArbitraryParams))

			var wg sync.WaitGroup
			wg.Add(len(apps))

			for appName, appPath := range apps {
				go func(appName, appPath string) {
					appURL := cf_helpers.PushAndBindApp(appName, serviceName, appPath)

					testService(config.AppType, appURL, appName)

					if testPlan.UpdateToPlan != "" {
						updatePlan(serviceName, testPlan.UpdateToPlan)
						testService(config.AppType, appURL, appName)
					}

					cf.UnbindService(appName, serviceName)
					wg.Done()
				}(appName, appPath)
			}

			wg.Wait()

			cf.DeleteService(serviceName)
		})
	}

	for _, plan := range config.TestPlans {
		lifecycle(plan)
	}
})

func testService(exampleAppType, testAppURL, appName string) {
	switch exampleAppType {
	case "crud":
		testCrud(testAppURL)
	case "fifo":
		testFifo(testAppURL, appName)
	default:
		Fail(fmt.Sprintf("invalid example app type %s. valid types are: crud, fifo", exampleAppType))
	}
}

func testCrud(testAppURL string) {
	cf_helpers.PutToTestApp(testAppURL, "foo", "bar")
	Expect(cf_helpers.GetFromTestApp(testAppURL, "foo")).To(Equal("bar"))
}

func testFifo(testAppURL, appName string) {
	queue := fmt.Sprintf("%s-queue", appName)
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
