package tests

import (
	"testing"
	"time"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grmqx"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/dbt"
	"github.com/integration-system/isp-kit/test/grmqt"
	"github.com/rabbitmq/amqp091-go"
	"msp-service-template/assembly"
	"msp-service-template/conf"
	"msp-service-template/entity"
)

func TestConsuming(t *testing.T) {
	test, require := test.New(t)
	testMq := grmqt.New(test)
	testDb := dbt.New(test, dbx.WithMigration("../migrations"))

	locator := assembly.NewLocator(testDb, test.Logger())
	config := conf.Consumer{
		Config: grmqx.Consumer{
			Queue: "test",
			Dlq:   true,
		},
	}
	brokerConfig := locator.BrokerConfig(config)
	testMq.Upgrade(brokerConfig)

	// invalid message
	testMq.Publish("", "test", amqp091.Publishing{Body: []byte("invalid json")})
	time.Sleep(1 * time.Second)
	require.EqualValues(1, testMq.QueueLength("test.DLQ"))

	// insert new message
	expected := entity.Message{
		Id:      1,
		Version: 2,
		Data:    entity.MessageData{Text: "a"},
	}

	testMq.PublishJson("", "test", expected)
	time.Sleep(1 * time.Second)
	actual := entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)

	// ignore message
	oldMsg := entity.Message{
		Id:      1,
		Version: 1,
		Data:    entity.MessageData{Text: "b"},
	}
	testMq.PublishJson("", "test", oldMsg)
	time.Sleep(1 * time.Second)
	actual = entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)

	// update message
	expected = entity.Message{
		Id:      1,
		Version: 3,
		Data:    entity.MessageData{Text: "b"},
	}
	testMq.PublishJson("", "test", expected)
	time.Sleep(1 * time.Second)
	actual = entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)
}
