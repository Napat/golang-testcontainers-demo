package event

import (
	"context"
	"fmt"
	"testing"

	"github.com/IBM/sarama"
	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/event"
	"github.com/Napat/golang-testcontainers-demo/test/integration"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

type ProducerTestSuite struct {
    integration.BaseTestSuite
    container testcontainers.Container
    ctx context.Context
    producer  sarama.SyncProducer
    repo      *event.Producer
}


// TestProducer runs the ProducerTestSuite.
//
// It uses Testcontainers to spin up a Kafka container,
// applies the necessary configuration scripts, and tests the
// ProducerRepository type.
func TestProducer(t *testing.T) {
    suite.Run(t, new(ProducerTestSuite))
}

// SetupSuite prepares the test environment for the ProducerTestSuite.
// It sets up a Kafka container using testcontainers, configures the container to
// log verbosely, uses a custom configuration file, and waits for the container to
// be ready to accept connections.
// 
// The method:
// - Initializes the base test suite
// - Creates a new context
// - Starts a Kafka container with specified configuration
// - Establishes a connection to the Kafka instance
// - Creates a new SyncProducer instance
// - Creates a test topic
// - Initializes the ProducerRepository instance
// 
// It doesn't return any values, but it populates the suite's fields with the necessary objects for testing.

// SetupSuite prepares the test environment for the ProducerTestSuite.
//
// It performs the following tasks:
// - Initializes the base test suite
// - Creates a new context
// - Starts a Kafka container with specified configuration
// - Establishes a connection to the Kafka instance
// - Configures and initializes a SyncProducer
// - Creates a test topic in Kafka
// - Initializes the ProducerRepository instance with the producer and test topic
//
// This method populates the suite's fields with necessary objects for testing
// and does not return any values.

func (s *ProducerTestSuite) SetupSuite() {
    s.BaseTestSuite.SetupSuite()

    s.ctx = context.Background()
    kafkaContainer, err := kafka.Run(s.ctx,
        "confluentinc/cp-kafka:7.8.0",
        kafka.WithClusterID("test-cluster"),
        // testcontainers.WithWaitStrategy(
        //     wait.ForLog("started (kafka.server.KafkaServer)").
        //         WithStartupTimeout(2*time.Minute),
        // ),
    )
    s.Require().NoError(err)
    s.container = kafkaContainer

    // Get connection details
    host, err := kafkaContainer.Host(s.ctx)
    s.Require().NoError(err)
    port, err := kafkaContainer.MappedPort(s.ctx, "9093")
    s.Require().NoError(err)

    // Configure Kafka producer
    config := sarama.NewConfig()
    config.Producer.Return.Successes = true
    producer, err := sarama.NewSyncProducer(
        []string{fmt.Sprintf("%s:%s", host, port.Port())},
        config,
    )
    s.Require().NoError(err)
    s.producer = producer

    // Create test topic
    brokerAddress := fmt.Sprintf("%s:%s", host, port.Port())
    admin, err := sarama.NewClusterAdmin([]string{brokerAddress}, config)
    s.Require().NoError(err)
    defer admin.Close()
    
    err = admin.CreateTopic("test-topic", &sarama.TopicDetail{
        NumPartitions:     1,
        ReplicationFactor: 1,
    }, false)
    s.Require().NoError(err)

    // Initialize repository
    s.repo = event.NewProducerRepository(producer, "test-topic")
}

// TearDownSuite tears down the test environment for the ProducerTestSuite.
// 
// The method:
// - Closes the Kafka producer connection
// - Cleans up the Kafka container
//
// The method doesn't take any parameters as it operates on the suite's fields.
// It doesn't return any values.
func (s *ProducerTestSuite) TearDownSuite() {
    if (s.producer != nil) {
        s.producer.Close()
    }
    if (s.container != nil) {
        s.CleanupContainer(s.container)
    }
}

// TestSendMessage tests the SendMessage method of the ProducerRepository.
//
// The test creates a sample User instance and uses SendMessage to send it as a message to the Kafka topic.
// Then, it verifies that the message is sent successfully without any errors.
func (s *ProducerTestSuite) TestSendMessage() {
    // Test data
    user := &model.User{
        ID:       1,
        Username: "testuser",
        Email:    "test@example.com",
    }

    // Test sending message
    err := s.repo.SendMessage("user-1", user)
    s.Require().NoError(err)
}
