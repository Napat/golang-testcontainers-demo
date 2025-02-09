package integration

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ElasticTestSuite struct {
    BaseTestSuite
    ESContainer testcontainers.Container
    ESClient    *elasticsearch.Client
}

// SetupElasticsearch initializes and starts an Elasticsearch container for testing purposes.
// It configures the container with specific settings, starts it, and creates an Elasticsearch client.
//
// The function performs the following steps:
// 1. Creates a container request with Elasticsearch image and configuration.
// 2. Starts the container using testcontainers.
// 3. Retrieves the host and port of the running container.
// 4. Creates and configures an Elasticsearch client.
//
// The function modifies the ElasticTestSuite struct, setting the ESContainer and ESClient fields.
// It uses s.Require() for asserting that no errors occur during the setup process.
func (s *ElasticTestSuite) SetupElasticsearch() {
    req := testcontainers.ContainerRequest{
        Image:        "docker.elastic.co/elasticsearch/elasticsearch:8.17.1",
        ExposedPorts: []string{"9200/tcp"},
        Env: map[string]string{
            "discovery.type":         "single-node",
            "xpack.security.enabled": "false",
            "ES_JAVA_OPTS":          "-Xms512m -Xmx512m",
            "bootstrap.memory_lock":  "false",
        },
        WaitingFor: wait.ForHTTP("/").WithPort("9200"),
        HostConfigModifier: func(hostConfig *container.HostConfig) {
            hostConfig.Resources = container.Resources{
                Memory:            2 * 1024 * 1024 * 1024,
                MemoryReservation: 1024 * 1024 * 1024,
            }
        },
    }

    container, err := testcontainers.GenericContainer(
        context.Background(),
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:         true,
        },
    )
    s.Require().NoError(err)
    s.ESContainer = container

    host, err := container.Host(context.Background())
    s.Require().NoError(err)
    port, err := container.MappedPort(context.Background(), "9200")
    s.Require().NoError(err)

    client, err := elasticsearch.NewClient(elasticsearch.Config{
        Addresses: []string{
            fmt.Sprintf("http://%s:%s", host, port.Port()),
        },
    })
    s.Require().NoError(err)
    s.ESClient = client
}

// TearDownElasticsearch terminates the Elasticsearch container if it exists.
// This function is typically used to clean up resources after tests have been run.
//
// It performs the following operation:
// - Checks if the ESContainer field of ElasticTestSuite is not nil.
// - If a container exists, it calls the Terminate method to stop and remove the container.
//
// The function does not take any parameters as it operates on the ElasticTestSuite struct.
// It does not return any value.
func (s *ElasticTestSuite) TearDownElasticsearch() {
    if s.ESContainer != nil {
        s.ESContainer.Terminate(context.Background())
    }
}
