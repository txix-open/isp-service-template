package tests

import (
	"context"
	"testing"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/dbt"
	"github.com/integration-system/isp-kit/test/grpct"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"msp-service-template/assembly"
)

type Object struct {
	Name string
}

func TestGetAll(t *testing.T) {
	assert, testDb, cli := prepareTest(t)

	result := make([]Object, 0)
	err := cli.Invoke("msp-service-template/object/all").
		ReadJsonResponse(&result).
		Do(context.Background())
	assert.NoError(err)
	assert.Empty(result)

	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")
	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 2, "b")

	result = make([]Object, 0)
	err = cli.Invoke("msp-service-template/object/all").
		ReadJsonResponse(&result).
		Do(context.Background())
	assert.NoError(err)

	expected := []Object{{
		Name: "a",
	}, {
		Name: "b",
	}}
	assert.EqualValues(expected, result)
}

func TestGetById(t *testing.T) {
	assert, testDb, cli := prepareTest(t)

	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")

	type reqBody struct {
		Id int
	}

	// empty req body
	result := Object{}
	err := cli.Invoke("msp-service-template/object/get_by_id").
		ReadJsonResponse(&result).
		Do(context.Background())
	assert.Error(err)
	assert.EqualValues(codes.InvalidArgument, status.Code(err))

	// id is required
	result = Object{}
	err = cli.Invoke("msp-service-template/object/get_by_id").
		ReadJsonResponse(&result).
		JsonRequestBody(reqBody{}).
		Do(context.Background())
	assert.Error(err)
	assert.EqualValues(codes.InvalidArgument, status.Code(err))

	// not found
	result = Object{}
	err = cli.Invoke("msp-service-template/object/get_by_id").
		ReadJsonResponse(&result).
		JsonRequestBody(reqBody{Id: 2}).
		Do(context.Background())
	assert.Error(err)
	assert.EqualValues(codes.NotFound, status.Code(err))

	// happy path
	result = Object{}
	err = cli.Invoke("msp-service-template/object/get_by_id").
		ReadJsonResponse(&result).
		JsonRequestBody(reqBody{Id: 1}).
		Do(context.Background())
	assert.NoError(err)

	expected := Object{Name: "a"}
	assert.EqualValues(expected, result)
}

func prepareTest(t *testing.T) (*require.Assertions, *dbt.TestDb, *client.Client) {
	test, assert := test.New(t)
	testDb := dbt.New(test, dbx.WithMigration("../migrations"))

	locator := assembly.NewLocator(testDb, test.Logger())
	_, cli := grpct.TestServer(test, locator.Handler())

	return assert, testDb, cli
}
