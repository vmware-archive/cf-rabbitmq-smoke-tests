package smoke_tests

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/tests/helper"
)

var _ = Describe("Smoke tests", func() {
	var (
		serviceName string
		appName     string
		appPath     string
	)

	BeforeEach(func() {
		serviceName = fmt.Sprintf("rmq-smoke-test-instance-%s", uuid.New()[:18])
		appName = "rmq-smoke-tests-ruby"
		appPath = "../assets/rabbit-example-app"
	})

	AfterEach(func() {
		helper.DeleteApp(appName)
	})

	for _, plan := range testConfig.TestPlans {
		It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ: plan '%s'", plan.Name), func() {

			helper.CreateService(testConfig.ServiceOffering, plan.Name, serviceName)

			// if it's TLS, then update the service to make TLS happen on the rabbit
			By("pushing and binding an app")
			appURL := helper.PushAndBindApp(appName, serviceName, appPath)
			if testConfig.TLSSupport != "disabled" {
				By("connecting to rabbit securely")
				appEnv := helper.GetAppEnv(appName)
				Expect(appEnv).To(ContainSubstring("amqps://"))
				Expect(appEnv).ToNot(ContainSubstring("amqp://"), "should not bind to app through amqp")
			}

			By("sending and receiving rabbit messages")
			queue := fmt.Sprintf("%s-queue", appName)

			helper.SendMessage(appURL, queue, "foo")
			helper.SendMessage(appURL, queue, "bar")
			Expect(helper.ReceiveMessage(appURL, queue)).To(Equal("foo"))
			Expect(helper.ReceiveMessage(appURL, queue)).To(Equal("bar"))

			helper.UnbindService(appName, serviceName)

			helper.DeleteService(serviceName)
		}, 300.0) // seconds
	}
})
