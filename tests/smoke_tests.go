package smoke_tests

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/cf"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/tests/cf_helpers"
)

var _ = Describe("Smoke tests", func() {
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

	for _, plan := range testConfig.TestPlans {
		It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ: plan '%s'", plan.Name), func() {
			cf.CreateService(testConfig.ServiceOffering, plan.Name, serviceName)

			var wg sync.WaitGroup
			wg.Add(len(apps))

			for appName, appPath := range apps {
				go func(appName, appPath string) {
					appURL := cf_helpers.PushAndBindApp(appName, serviceName, appPath)

					queue := fmt.Sprintf("%s-queue", appName)
					cf_helpers.PushToTestAppQueue(appURL, queue, "foo")
					cf_helpers.PushToTestAppQueue(appURL, queue, "bar")
					Expect(cf_helpers.PopFromTestAppQueue(appURL, queue)).To(Equal("foo"))
					Expect(cf_helpers.PopFromTestAppQueue(appURL, queue)).To(Equal("bar"))

					cf.UnbindService(appName, serviceName)
					wg.Done()
				}(appName, appPath)
			}

			wg.Wait()

			cf.DeleteService(serviceName)
		}, 300.0) // seconds
	}
})
