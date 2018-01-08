package lifecycle_tests

import (
	"encoding/json"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/services"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cf "github.com/pivotal-cf/cf-rabbitmq-smoke-tests/cf"

	"testing"
)

var (
	configPath        = os.Getenv("CONFIG_PATH")
	cfConfig          = loadCFConfig(configPath)
	rabbitmqConfig    = loadRabbitmqConfig(configPath)
	securityGroupName = "cf-rabbitmq-smoke-tests-security-group"
	quotaName         = "cf-rabbitmq-smoke-tests-quota"
)

func TestLifecycle(t *testing.T) {
	SynchronizedBeforeSuite(func() []byte {
		cf.Api(cfConfig.ApiEndpoint, cfConfig.SkipSSLValidation)
		cf.Auth(cfConfig.AdminUser, cfConfig.AdminPassword)
		cf.CreateOrg(cfConfig.OrgName)
		cf.CreateSpace(cfConfig.OrgName, rabbitmqConfig.SpaceName)
		cf.Target(cfConfig.OrgName, rabbitmqConfig.SpaceName)
		cf.CreateAndBindSecurityGroup(securityGroupName, cfConfig.OrgName, rabbitmqConfig.SpaceName)
		cf.CreateAndSetQuota(quotaName, cfConfig.OrgName)

		for _, testPlan := range rabbitmqConfig.TestPlans {
			cf.EnableServiceAccess(rabbitmqConfig.ServiceOffering, testPlan.Name, cfConfig.OrgName)
		}

		return []byte{}
	}, func([]byte) {
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		cf.Target(cfConfig.OrgName, rabbitmqConfig.SpaceName)

		cf.DeleteSpace(rabbitmqConfig.SpaceName)
		cf.DeleteOrg(cfConfig.OrgName)

		cf.DeleteQuota(quotaName)
		cf.DeleteSecurityGroup(securityGroupName)

		for _, testPlan := range rabbitmqConfig.TestPlans {
			cf.DisableServiceAccess(rabbitmqConfig.ServiceOffering, testPlan.Name, cfConfig.OrgName)
		}
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Lifecycle Suite")
}

func loadCFConfig(configPath string) services.Config {
	var err error
	config := services.Config{}

	if err = services.LoadConfig(configPath, &config); err != nil {
		panic(err)
	}

	if err = services.ValidateConfig(&config); err != nil {
		panic(err)
	}

	return config
}

func loadRabbitmqConfig(configPath string) RabbitMQTestConfig {
	config, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}

	defer config.Close()
	var rmqConfig RabbitMQTestConfig
	err = json.NewDecoder(config).Decode(&rmqConfig)
	if err != nil {
		panic(err)
	}

	return rmqConfig
}

type RabbitMQTestConfig struct {
	TestPlans       []TestPlan `json:"plans"`
	ServiceOffering string     `json:"service_offering"`
	AppType         string     `json:"app_type"`
	SpaceName       string     `json:"space_name"`
}

type TestPlan struct {
	Name            string          `json:"name"`
	UpdateToPlan    string          `json:"update_to_plan"`
	ArbitraryParams json.RawMessage `json:"arbitrary_params"`
}
