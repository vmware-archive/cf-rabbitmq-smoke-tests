package smoke_tests

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/tests/helper"
)

var _ = Describe("Smoke tests", func() {

	const appName = "rmq-smoke-tests-ruby"
	const appPath = "../assets/rabbit-example-app"

	AfterEach(func() {
		helper.DeleteApp(appName)
	})

	smokeTestForPlan := func(planName string) func() {
		return func() {
			serviceName := fmt.Sprintf("rmq-smoke-test-instance-%s", uuid.New()[:18])
			helper.CreateService(testConfig.ServiceOffering, planName, serviceName)

			if useTLS {
				By("enabling TLS")
				helper.EnableTLSForODB(serviceName)
			}

			By("pushing and binding an app")
			appURL := helper.PushAndBindApp(appName, serviceName, appPath)

			By("RabbitMQ protocol")
			appEnv := helper.GetAppEnv(appName)
			if useTLS {
				Expect(appEnv).To(ContainSubstring("amqps://"), "bind should expose amqps protocol")
				Expect(appEnv).ToNot(ContainSubstring("amqp://"), "bind should not expose amqp protocol")
			} else {
				Expect(appEnv).ToNot(ContainSubstring("amqps://"), "bind should not expose amqps protocol")
				Expect(appEnv).To(ContainSubstring("amqp://"), "bind should expose amqp protocol")
			}

			By("sending and receiving rabbit messages")
			queue := fmt.Sprintf("%s-queue", appName)

			helper.SendMessage(appURL, queue, "foo")
			helper.SendMessage(appURL, queue, "bar")
			Expect(helper.ReceiveMessage(appURL, queue)).To(Equal("foo"))
			Expect(helper.ReceiveMessage(appURL, queue)).To(Equal("bar"))

			helper.UnbindService(appName, serviceName)

			helper.DeleteService(serviceName)
		}
	}

	for _, plan := range testConfig.TestPlans {
		It(fmt.Sprintf("pushes an app, sends, and reads a message from RabbitMQ: plan '%s'", plan.Name),
			smokeTestForPlan(plan.Name), 300.0) // seconds
	}
})
