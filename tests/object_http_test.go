package tests_test

import (
	"net/http"
	"testing"

	"isp-service-template/assembly"
	"isp-service-template/domain"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/httpt"
)

type objectHttpSuite struct {
	suite.Suite

	db  *dbt.TestDb
	api *httpcli.Client
}

func TestObjectHttp(t *testing.T) {
	t.Parallel()
	suite.Run(t, &objectHttpSuite{})
}

func (s *objectHttpSuite) SetupTest() {
	ispTest, _ := test.New(s.T())
	s.db = dbt.New(ispTest, dbx.WithMigrationRunner("../migrations", ispTest.Logger()))

	locator := assembly.NewLocator(s.db, ispTest.Logger())
	locatorCfg := locator.Config(getRemoteConfig())

	_, s.api = httpt.TestServer(ispTest, locatorCfg.HttpHandler)
}

func (s *objectHttpSuite) Test_GetAll_Empty() {
	result := make([]domain.Object, 0)
	err := s.api.Post("/isp-service-template/object/all").
		JsonResponseBody(&result).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
	s.Require().Empty(result)
}

func (s *objectHttpSuite) Test_GetAll_HappyPath() {
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 2, "b")

	result := make([]domain.Object, 0)
	err := s.api.Post("/isp-service-template/object/all").
		JsonResponseBody(&result).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)

	expected := []domain.Object{{Name: "a"}, {Name: "b"}}
	s.Require().EqualValues(expected, result)
}

func (s *objectHttpSuite) Test_GetById_EmptyBody() {
	err := s.api.Post("/isp-service-template/object/get_by_id").
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	expectApiError(s.Require(), err, http.StatusBadRequest, 400)
}

func (s *objectHttpSuite) Test_GetById_MissingId() {
	err := s.api.Post("/isp-service-template/object/get_by_id").
		JsonRequestBody(domain.ByIdRequest{}).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	expectApiError(s.Require(), err, http.StatusBadRequest, 400)
}

func (s *objectHttpSuite) Test_GetById_NotFound() {
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")

	err := s.api.Post("/isp-service-template/object/get_by_id").
		JsonRequestBody(domain.ByIdRequest{Id: 2}).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	expectApiError(s.Require(), err, http.StatusInternalServerError, 900)
}

func (s *objectHttpSuite) Test_GetById_HappyPath() {
	s.db.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")

	result := domain.Object{}
	err := s.api.Post("/isp-service-template/object/get_by_id").
		JsonRequestBody(domain.ByIdRequest{Id: 1}).
		JsonResponseBody(&result).
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)

	expected := domain.Object{Name: "a"}
	s.Require().Equal(expected, result)
}
