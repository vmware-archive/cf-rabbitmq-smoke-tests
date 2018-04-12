package smoke_tests

import (
	"encoding/json"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
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

		return []byte{}
	}, func([]byte) {
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		wfh.Teardown()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Lifecycle Suite")
}

func loadTestConfig(configPath string) TestConfig {
	configFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}

	defer configFile.Close()
	var testConfig TestConfig
	err = json.NewDecoder(configFile).Decode(&testConfig)
	if err != nil {
		panic(err)
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
