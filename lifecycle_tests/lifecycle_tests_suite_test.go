package lifecycle_tests

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/cf"

	"testing"
)

var (
	configPath        = os.Getenv("CONFIG_PATH")
	config            = loadRabbitmqConfig(configPath)
	securityGroupName = "cf-rabbitmq-smoke-tests-security-group"
	quotaName         = "cf-rabbitmq-smoke-tests-quota"
)

func TestLifecycle(t *testing.T) {
	SynchronizedBeforeSuite(func() []byte {
		cf.Api(config.ApiEndpoint, config.SkipSSLValidation)
		cf.Auth(config.AdminUser, config.AdminPassword)
		cf.CreateOrg(config.OrgName)
		cf.CreateSpace(config.OrgName, config.SpaceName)
		cf.Target(config.OrgName, config.SpaceName)
		cf.CreateAndBindSecurityGroup(securityGroupName, config.OrgName, config.SpaceName)
		cf.CreateAndSetQuota(quotaName, config.OrgName)

		for _, testPlan := range config.TestPlans {
			cf.EnableServiceAccess(config.ServiceOffering, testPlan.Name, config.OrgName)
		}

		return []byte{}
	}, func([]byte) {
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		cf.Target(config.OrgName, config.SpaceName)

		for _, testPlan := range config.TestPlans {
			cf.DisableServiceAccess(config.ServiceOffering, testPlan.Name, config.OrgName)
		}

		cf.DeleteSpace(config.SpaceName)
		cf.DeleteOrg(config.OrgName)

		cf.DeleteQuota(quotaName)
		cf.DeleteSecurityGroup(securityGroupName)
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Lifecycle Suite")
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
	TestPlans                     []TestPlan `json:"plans"`
	ServiceOffering               string     `json:"service_offering"`
	AppType                       string     `json:"app_type"`
	SpaceName                     string     `json:"space_name"`
	AppsDomain                    string     `json:"apps_domain"`
	ApiEndpoint                   string     `json:"api"`
	AdminUser                     string     `json:"admin_user"`
	AdminPassword                 string     `json:"admin_password"`
	CreatePermissiveSecurityGroup bool       `json:"create_permissive_security_group"`
	SkipSSLValidation             bool       `json:"skip_ssl_validation"`
	TimeoutScale                  float64    `json:"timeout_scale"`
	OrgName                       string     `json:"org_name"`
}

type TestPlan struct {
	Name            string          `json:"name"`
	UpdateToPlan    string          `json:"update_to_plan"`
	ArbitraryParams json.RawMessage `json:"arbitrary_params"`
}
