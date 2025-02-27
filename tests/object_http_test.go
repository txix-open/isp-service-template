package tests_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/dbx"
	client "github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/httpt"
	"msp-service-template/assembly"
	"msp-service-template/conf"
)

type Object struct {
	Name string
}

func TestGetAllHttp(t *testing.T) {
	t.Parallel()
	assert, testDb, cli := prepareHttpTest(t)

	result := make([]Object, 0)
	_, err := cli.Post("/object/all").
		JsonResponseBody(&result).
		Do(context.Background())
	assert.NoError(err)
	assert.Empty(result)

	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")
	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 2, "b")

	result = make([]Object, 0)
	_, err = cli.Post("/object/all").
		JsonResponseBody(&result).
		Do(context.Background())
	assert.NoError(err)

	expected := []Object{{
		Name: "a",
	}, {
		Name: "b",
	}}
	assert.EqualValues(expected, result)
}

func TestGetByIdHttp(t *testing.T) {
	t.Parallel()
	assert, testDb, cli := prepareHttpTest(t)

	testDb.Must().Exec("insert into object (id, name) values ($1, $2)", 1, "a")

	type reqBody struct {
		Id int
	}

	// empty req body
	resp, err := cli.Post("/object/get_by_id").
		Do(context.Background())
	assert.NoError(err)
	assert.EqualValues(http.StatusBadRequest, resp.StatusCode())

	// id is required
	_, err = cli.Post("/object/get_by_id").
		JsonRequestBody(reqBody{}).
		Do(context.Background())
	assert.NoError(err)
	assert.EqualValues(http.StatusBadRequest, resp.StatusCode())

	// not found
	_, err = cli.Post("/object/get_by_id").
		JsonRequestBody(reqBody{Id: 2}).
		Do(context.Background())
	assert.NoError(err)
	assert.EqualValues(http.StatusBadRequest, resp.StatusCode())

	// happy path
	okResult := Object{}
	_, err = cli.Post("/object/get_by_id").
		JsonResponseBody(&okResult).
		JsonRequestBody(reqBody{Id: 1}).
		Do(context.Background())
	assert.NoError(err)

	expected := Object{Name: "a"}
	assert.EqualValues(expected, okResult)
}

func prepareHttpTest(t *testing.T) (*require.Assertions, *dbt.TestDb, *client.Client) {
	t.Helper()
	test, assert := test.New(t)
	testDb := dbt.New(test, dbx.WithMigrationRunner("../migrations", test.Logger()))

	locator := assembly.NewLocator(testDb, test.Logger())
	h := locator.Handlers(conf.Remote{})
	_, cli := httpt.TestServer(test, h.HttpHandler)

	return assert, testDb, cli
}
