package tests_test

import (
	"testing"

	"isp-service-template/assembly"
	"isp-service-template/conf"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/grpct"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetAllGrpc(t *testing.T) {
	t.Parallel()

	require, testDb, cli := prepareGrpcTest(t)

	result := make([]Object, 0)
	err := cli.Invoke("isp-service-template/object/all").
		JsonResponseBody(&result).
		Do(t.Context())
	require.NoError(err)
	require.Empty(result)

	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")
	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 2, "b")

	result = make([]Object, 0)
	err = cli.Invoke("isp-service-template/object/all").
		JsonResponseBody(&result).
		Do(t.Context())
	require.NoError(err)

	expected := []Object{{
		Name: "a",
	}, {
		Name: "b",
	}}
	require.EqualValues(expected, result)
}

type reqBody struct {
	Id int
}

func TestGetByIdGrpc(t *testing.T) {
	t.Parallel()
	require, testDb, cli := prepareGrpcTest(t)

	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")

	// empty req body
	result := Object{}
	err := cli.Invoke("isp-service-template/object/get_by_id").
		JsonResponseBody(&result).
		Do(t.Context())
	require.Error(err)
	require.EqualValues(codes.InvalidArgument, status.Code(err))

	// id is required
	result = Object{}
	err = cli.Invoke("isp-service-template/object/get_by_id").
		JsonResponseBody(&result).
		JsonRequestBody(reqBody{}).
		Do(t.Context())
	require.Error(err)
	require.EqualValues(codes.InvalidArgument, status.Code(err))

	// not found
	result = Object{}
	err = cli.Invoke("isp-service-template/object/get_by_id").
		JsonResponseBody(&result).
		JsonRequestBody(reqBody{Id: 2}).
		Do(t.Context())
	require.Error(err)
	require.EqualValues(codes.InvalidArgument, status.Code(err))
	businessError := apierrors.FromError(err)
	require.NotNil(businessError)
	require.EqualValues(800, businessError.ErrorCode)

	// happy path
	result = Object{}
	err = cli.Invoke("isp-service-template/object/get_by_id").
		JsonResponseBody(&result).
		JsonRequestBody(reqBody{Id: 1}).
		Do(t.Context())
	require.NoError(err)

	expected := Object{Name: "a"}
	require.EqualValues(expected, result)
}

func prepareGrpcTest(t *testing.T) (*require.Assertions, *dbt.TestDb, *client.Client) {
	t.Helper()

	test, require := test.New(t)
	testDb := dbt.New(test, dbx.WithMigrationRunner("../migrations", test.Logger()))

	locator := assembly.NewLocator(testDb, test.Logger())
	config := locator.Config(conf.Remote{})
	_, cli := grpct.TestServer(test, config.GrpcHandler)

	return require, testDb, cli
}
