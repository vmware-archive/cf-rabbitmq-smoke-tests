package smoke_tests

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cf-rabbitmq-smoke-tests/tests/helper"

	"testing"
)

const (
	securityGroupName = "cf-rabbitmq-smoke-tests"
	quotaName         = "cf-rabbitmq-smoke-tests-quota"
)

var (
	configPath = os.Getenv("CONFIG_PATH")
	testConfig = loadTestConfig(configPath)
	wfh        *workflowhelpers.ReproducibleTestSuiteSetup
)

func TestLifecycle(t *testing.T) {
	SynchronizedBeforeSuite(func() []byte {
		wfh = workflowhelpers.NewTestSuiteSetup(&testConfig.Config)
		wfh.Setup()

		helper.CreateAndBindSecurityGroup(securityGroupName, wfh.GetOrganizationName())
		helper.CreateAndSetQuota(quotaName, wfh.GetOrganizationName())

		return []byte{}
	}, func([]byte) {
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		wfh.Teardown()

		helper.DeleteQuota(quotaName)
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
	config.Config

	TestPlans       []TestPlan `json:"plans"`
	ServiceOffering string     `json:"service_offering"`
	AppType         string     `json:"app_type"`
}

type TestPlan struct {
	Name string `json:"name"`
}
