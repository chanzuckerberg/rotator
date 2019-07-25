// Code for mocking the Travis CI API Client (e.g. SetupTest(), setup(), testMethod(), and testBody())
// is adapted from: https://github.com/shuheiktgw/go-travis/blob/master/env_vars_test.go (accessed July 24, 2019)

package sink_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/shuheiktgw/go-travis"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testRepoSlug = "chanzuckerberg/rotator-test"
	name         = "TEST"
	value        = "test"
	id           = "test-12345"
	public       = false
	fakeName     = "TEST non-existing"
)

type TravisTestSuite struct {
	suite.Suite

	ctx      context.Context
	client   *travis.Client
	mux      *http.ServeMux
	teardown func()

	sink *sink.TravisCiSink
}

func (ts *TravisTestSuite) TearDownTest() {
	ts.teardown()
}

func (ts *TravisTestSuite) SetupTest() {
	ts.ctx = context.Background()
	ts.client, ts.mux, _, ts.teardown = setup()
	t := ts.T()

	// mock ListByRepoSlug()
	ts.mux.HandleFunc(fmt.Sprintf("/repo/%s/env_vars", testRepoSlug), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"env_vars": [{"id":"%s","name":"%s","value":"","public":%t}]}`, id, name, public)
	})

	// mock UpdateByRepoSlug()
	ts.mux.HandleFunc(fmt.Sprintf("/repo/%s/env_var/%s", testRepoSlug, id), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		testBody(t, r, fmt.Sprintf(`{"env_var.name":"%s","env_var.value":"%s","env_var.public":%t}`, name, value, public)+"\n")
		fmt.Fprintf(w, `{"id":"%s","name":"%s","value":"%s","public":%t}`, id, name, value, public)
	})

	// create Travis CI sink
	ts.sink = &sink.TravisCiSink{
		RepoSlug: testRepoSlug,
		Client:   ts.client,
	}
}

func (ts *TravisTestSuite) TestWriteToTravisCiSink() {
	t := ts.T()
	r := require.New(t)
	err := ts.sink.Write(ts.ctx, name, value)
	r.Nil(err)
}

func (ts *TravisTestSuite) TestWriteToTravisCiSink_FakeEnvVar() {
	t := ts.T()
	r := require.New(t)
	err := ts.sink.Write(ts.ctx, fakeName, value)
	r.Error(err, fmt.Sprintf("env var %s does not exist in Travis CI for repo %s", fakeName, testRepoSlug))
}

func TestTravisProviderSuite(t *testing.T) {
	suite.Run(t, new(TravisTestSuite))
}

func setup() (client *travis.Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle("/", mux)

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the GitHub client being tested and is
	// configured to use test server.
	client = travis.NewClient("", "")
	u, _ := url.Parse(server.URL + "/")
	client.BaseURL = u

	return client, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testBody(t *testing.T, r *http.Request, want string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Error reading request body: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("request Body is %s, want %s", got, want)
	}
}
