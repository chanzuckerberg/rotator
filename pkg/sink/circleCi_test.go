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
	"github.com/jszwedko/go-circleci"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	circleAccount   = "testo_account"
	circleRepo      = "testo_repo"
	circleBadRepo   = "bad_testo_repo"
	circleEnvVar    = "foo"
	circleEnvVarVal = "bar"
)

type CircleTestSuite struct {
	suite.Suite

	ctx      context.Context
	client   *circleci.Client
	mux      *http.ServeMux
	teardown func()
}

func (ts *CircleTestSuite) TearDownTest() {
	ts.teardown()
}

func (ts *CircleTestSuite) SetupTest() {
	mux := http.NewServeMux()
	apiHandler := http.NewServeMux()
	apiHandler.Handle("/", mux)

	server := httptest.NewServer(apiHandler)
	u, _ := url.Parse(server.URL + "/")
	client := &circleci.Client{BaseURL: u}

	ts.ctx = context.Background()
	ts.client = client
	ts.mux = mux
	ts.teardown = server.Close

	t := ts.T()
	a := assert.New(t)

	mux.HandleFunc(
		fmt.Sprintf("/project/%s/%s/envvar", circleAccount, circleRepo),
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				a.Fail("invalid http method %s", r.Method)
				return
			}

			b, err := ioutil.ReadAll(r.Body)
			a.NoError(err)

			want := fmt.Sprintf(`{"name":"%s","value":"%s"}`, circleEnvVar, circleEnvVarVal)
			if want != string(b) {
				http.Error(w, "body doesn't match", http.StatusBadRequest)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, want)
			return
		},
	)
}

func (ts *CircleTestSuite) TestWriteToCircleCiSink() {
	t := ts.T()
	a := assert.New(t)
	sink := &sink.CircleCiSink{}
	sink.WithCircleClient(ts.client, circleAccount, circleRepo)
	err := sink.Write(ts.ctx, circleEnvVar, circleEnvVarVal)
	a.NoError(err)
}

func TestCircleCISuite(t *testing.T) {
	suite.Run(t, new(CircleTestSuite))
}
