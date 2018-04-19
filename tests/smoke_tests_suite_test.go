package smoke_tests

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/tests/helper"

	"testing"
)

var (
	configPath        = os.Getenv("CONFIG_PATH")
	testConfig        = loadTestConfig(configPath)
	spaceName         = generator.PrefixedRandomName(testConfig.NamePrefix, "space")
	securityGroupName = generator.PrefixedRandomName(testConfig.NamePrefix, "security-group")
)

func TestLifecycle(t *testing.T) {
	SynchronizedBeforeSuite(func() []byte {

		helper.Api(testConfig.ApiEndpoint, testConfig.SkipSSLValidation)
		helper.Auth(testConfig.AdminUser, testConfig.AdminPassword)
		helper.CreateSpace(testConfig.ExistingOrganization, spaceName)
		helper.Target(testConfig.ExistingOrganization, spaceName)
		helper.CreateAndBindSecurityGroup(securityGroupName, testConfig.ExistingOrganization, spaceName)

		return []byte{}
	}, func([]byte) {
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		helper.Target(testConfig.ExistingOrganization, spaceName)
		helper.DeleteSpace(spaceName)
		helper.DeleteSecurityGroup(securityGroupName)
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Smoke Tests Suite")
}

func loadTestConfig(configPath string) TestConfig {
	if configPath == "" {
		panic(fmt.Errorf("Path to config file is empty -- Did you set CONFIG_PATH?"))
	}
	configFile, err := os.Open(configPath)
	if err != nil {
		panic(fmt.Errorf("Could not open config file at %s --  ERROR %s", configPath, err.Error()))
	}

	defer configFile.Close()
	var testConfig TestConfig
	err = json.NewDecoder(configFile).Decode(&testConfig)
	if err != nil {
		panic(fmt.Errorf("Could not decode config json -- ERROR: %s", err.Error()))
	}

	return testConfig
}

type TestConfig struct {
	ApiEndpoint          string     `json:"api"`
	SkipSSLValidation    bool       `json:"skip_ssl_validation"`
	AdminUser            string     `json:"admin_user"`
	AdminPassword        string     `json:"admin_password"`
	ExistingOrganization string     `json:"existing_organization"`
	TestPlans            []TestPlan `json:"plans"`
	ServiceOffering      string     `json:"service_offering"`
	AppType              string     `json:"app_type"`
	NamePrefix           string     `json:"name_prefix"`
}

type TestPlan struct {
	Name string `json:"name"`
}
