package tests

import (
	"context"
	"encoding/json"
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
		Client: testMq.ConnectionConfig(),
		Config: grmqx.Consumer{
			Queue: "test",
			Dlq:   true,
		},
	}
	brokerConfig := locator.BrokerConfig(config)

	mqCli := grmqx.New(test.Logger())
	test.T().Cleanup(func() {
		mqCli.Close()
	})
	err := mqCli.Upgrade(context.Background(), brokerConfig)
	require.NoError(err)

	//invalid message
	testMq.Publish("", "test", amqp091.Publishing{Body: []byte("invalid json")})
	time.Sleep(1 * time.Second)
	require.EqualValues(1, testMq.QueueLength("test.DLQ"))

	//insert new message
	expected := entity.Message{
		Id:      1,
		Version: 2,
		Data:    entity.MessageData{Text: "a"},
	}
	data, err := json.Marshal(expected)
	require.NoError(err)
	testMq.Publish("", "test", amqp091.Publishing{Body: data})
	time.Sleep(1 * time.Second)
	actual := entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)

	//ignore message
	oldMsg := entity.Message{
		Id:      1,
		Version: 1,
		Data:    entity.MessageData{Text: "b"},
	}
	data, err = json.Marshal(oldMsg)
	require.NoError(err)
	testMq.Publish("", "test", amqp091.Publishing{Body: data})
	time.Sleep(1 * time.Second)
	actual = entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)

	//update message
	expected = entity.Message{
		Id:      1,
		Version: 3,
		Data:    entity.MessageData{Text: "b"},
	}
	data, err = json.Marshal(expected)
	require.NoError(err)
	testMq.Publish("", "test", amqp091.Publishing{Body: data})
	time.Sleep(1 * time.Second)
	actual = entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)
}
