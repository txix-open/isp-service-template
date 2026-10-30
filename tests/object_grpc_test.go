package tests_test

import (
	"testing"

	"isp-service-template/assembly"
	"isp-service-template/domain"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/grpct"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type objectGrpcSuite struct {
	suite.Suite

	db  *dbt.TestDb
	cli *client.Client
}

func TestObjectGrpc(t *testing.T) {
	t.Parallel()
	suite.Run(t, &objectGrpcSuite{})
}

func (s *objectGrpcSuite) SetupTest() {
	ispTest, _ := test.New(s.T())
	s.db = dbt.New(ispTest, dbx.WithMigrationRunner("../migrations", ispTest.Logger()))

	locator := assembly.NewLocator(s.db, ispTest.Logger())
	locatorCfg := locator.Config(getRemoteConfig())

	_, s.cli = grpct.TestServer(ispTest, locatorCfg.GrpcHandler)
}

func (s *objectGrpcSuite) Test_GetAll_Empty() {
	result := make([]domain.Object, 0)
	err := s.cli.Invoke("isp-service-template/object/all").
		JsonResponseBody(&result).
		Do(s.T().Context())
	s.Require().NoError(err)
	s.Require().Empty(result)
}

func (s *objectGrpcSuite) Test_GetAll_HappyPath() {
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 2, "b")

	result := make([]domain.Object, 0)
	err := s.cli.Invoke("isp-service-template/object/all").
		JsonResponseBody(&result).
		Do(s.T().Context())
	s.Require().NoError(err)

	expected := []domain.Object{{Name: "a"}, {Name: "b"}}
	s.Require().EqualValues(expected, result)
}

func (s *objectGrpcSuite) Test_GetById_EmptyBody() {
	err := s.cli.Invoke("isp-service-template/object/get_by_id").
		Do(s.T().Context())
	s.Require().Error(err)
	s.Require().EqualValues(codes.InvalidArgument, status.Code(err))
}

func (s *objectGrpcSuite) Test_GetById_MissingId() {
	err := s.cli.Invoke("isp-service-template/object/get_by_id").
		JsonRequestBody(domain.ByIdRequest{}).
		Do(s.T().Context())
	s.Require().Error(err)
	s.Require().EqualValues(codes.InvalidArgument, status.Code(err))
}

func (s *objectGrpcSuite) Test_GetById_NotFound() {
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")

	err := s.cli.Invoke("isp-service-template/object/get_by_id").
		JsonRequestBody(domain.ByIdRequest{Id: 2}).
		Do(s.T().Context())
	s.Require().Error(err)
	s.Require().EqualValues(codes.InvalidArgument, status.Code(err))

	businessError := apierrors.FromError(err)
	s.Require().NotNil(businessError)
	s.Require().EqualValues(domain.ErrCodeObjectNotFound, businessError.ErrorCode)
}

func (s *objectGrpcSuite) Test_GetById_HappyPath() {
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")

	result := domain.Object{}
	err := s.cli.Invoke("isp-service-template/object/get_by_id").
		JsonRequestBody(domain.ByIdRequest{Id: 1}).
		JsonResponseBody(&result).
		Do(s.T().Context())
	s.Require().NoError(err)

	expected := domain.Object{Name: "a"}
	s.Require().Equal(expected, result)
}
