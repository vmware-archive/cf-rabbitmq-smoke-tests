package service_test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testConfig struct {
	services.Config

	ServiceName string   `json:"service_name"`
	PlanNames   []string `json:"plan_names"`
}

func loadConfig(path string) (cfg testConfig) {
	if path == "" {
		fatal("Path to config file is empty", fmt.Errorf("CONFIG_PATH not set"))
	}
	configFile, err := os.Open(path)
	if err != nil {
		fatal(fmt.Sprintf("Failed while opening config file at %s", path), err)
	}

	decoder := json.NewDecoder(configFile)
	if err = decoder.Decode(&cfg); err != nil {
		fatal("Cannot decode config json", err)
	}

	return
}

var (
	config = loadConfig(os.Getenv("CONFIG_PATH"))
	ctx    services.Context
)

func fatal(message string, err error) {
	log.Fatalf("%s -- ERROR: %s", message, err.Error())
}

func TestService(t *testing.T) {
	if err := services.ValidateConfig(&config.Config); err != nil {
		fatal("Invalid config", err)
	}

	ctx = services.NewContext(config.Config, "rabbitmq-smoke-tests")

	RegisterFailHandler(Fail)

	RunSpecs(t, "RabbitMQ Smoke Tests")
}

var _ = BeforeEach(func() {
	ctx.Setup()
})

var _ = AfterEach(func() {
	ctx.Teardown()
})
