package tests_test

import (
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/fake"
	"github.com/txix-open/isp-kit/test/grmqt"
	"msp-service-template/assembly"
	"msp-service-template/conf"
	"msp-service-template/entity"
)

func TestConsuming(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	testMq := grmqt.New(test)
	testDb := dbt.New(test, dbx.WithMigrationRunner("../migrations", test.Logger()))

	locator := assembly.NewLocator(testDb, test.Logger())
	config := conf.Consumer{
		Config: grmqx.Consumer{
			Queue: "test",
			Dlq:   true,
		},
	}
	h := locator.Handlers(conf.Remote{Consumer: config})
	testMq.Upgrade(grmqx.NewConfig(
		config.Client.Url(),
		grmqx.WithConsumers(h.RmqHandler),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(config.Config)),
	))

	// invalid message
	testMq.Publish("", "test", amqp091.Publishing{Body: []byte("invalid json")})
	time.Sleep(1 * time.Second)
	require.EqualValues(1, testMq.QueueLength("test.DLQ"))

	// insert new message
	expected := fake.It[entity.Message]()

	testMq.PublishJson("", "test", expected)
	time.Sleep(1 * time.Second)
	actual := entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)

	// ignore message
	oldMsg := fake.It[entity.Message]()
	oldMsg.Version = expected.Version
	testMq.PublishJson("", "test", oldMsg)
	time.Sleep(1 * time.Second)
	actual = entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)

	// update message
	oldVersion := expected.Version
	expected = fake.It[entity.Message]()
	expected.Version = oldVersion + 1
	testMq.PublishJson("", "test", expected)
	time.Sleep(1 * time.Second)
	actual = entity.Message{}
	testDb.Must().SelectRow(&actual, "select * from message where id = $1", expected.Id)
	require.EqualValues(0, testMq.QueueLength("test"))
	require.EqualValues(expected, actual)
}
