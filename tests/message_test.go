package tests_test

import (
	"testing"
	"time"

	"isp-service-template/assembly"
	"isp-service-template/entity"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/fake"
	"github.com/txix-open/isp-kit/test/grmqt"
)

type messageSuite struct {
	suite.Suite

	mq *grmqt.Client
	db *dbt.TestDb
}

func TestMessage(t *testing.T) {
	t.Parallel()
	suite.Run(t, &messageSuite{})
}

func (s *messageSuite) SetupTest() {
	ispTest, _ := test.New(s.T())
	s.mq = grmqt.New(ispTest)
	s.db = dbt.New(ispTest, dbx.WithMigrationRunner("../migrations", ispTest.Logger()))

	cfg := getRemoteConfig()
	locator := assembly.NewLocator(s.db, ispTest.Logger())
	locatorCfg := locator.Config(cfg)

	s.mq.Upgrade(grmqx.NewConfig(
		cfg.Consumer.Client.Url(),
		grmqx.WithConsumers(locatorCfg.RmqHandler),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(cfg.Consumer.Config)),
	))
}

func (s *messageSuite) Test_Consuming_InvalidMessage_Dlq() {
	s.mq.Publish("", testQueue, amqp091.Publishing{Body: []byte("invalid json")})
	time.Sleep(time.Second)
	s.Require().EqualValues(1, s.mq.QueueLength(testQueue+".DLQ"))
}

func (s *messageSuite) Test_Consuming_InsertNewMessage() {
	expected := fake.It[entity.Message]()
	s.mq.PublishJson("", testQueue, expected)
	time.Sleep(time.Second)

	actual := s.selectMessage(expected.Id)
	s.Require().EqualValues(0, s.mq.QueueLength(testQueue))
	s.Require().EqualValues(expected, actual)
}

func (s *messageSuite) Test_Consuming_RepeatVersion_Ignored() {
	msg := fake.It[entity.Message]()
	s.mq.PublishJson("", testQueue, msg)
	time.Sleep(time.Second)

	s.mq.PublishJson("", testQueue, msg)
	time.Sleep(time.Second)

	actual := s.selectMessage(msg.Id)
	s.Require().EqualValues(0, s.mq.QueueLength(testQueue))
	s.Require().EqualValues(msg, actual)
}

func (s *messageSuite) Test_Consuming_NewVersion_Updated() {
	msg := fake.It[entity.Message]()
	s.mq.PublishJson("", testQueue, msg)
	time.Sleep(time.Second)

	msg.Version++
	s.mq.PublishJson("", testQueue, msg)
	time.Sleep(time.Second)

	actual := s.selectMessage(msg.Id)
	s.Require().EqualValues(0, s.mq.QueueLength(testQueue))
	s.Require().EqualValues(msg, actual)
}

func (s *messageSuite) selectMessage(id int64) entity.Message {
	actual := entity.Message{}
	s.db.Must().SelectRow(&actual, "select * from message where id = $1", id)
	return actual
}
