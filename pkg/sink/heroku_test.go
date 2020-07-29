package sink

import (
	"context"
	"testing"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	herokuEnvVar    = "foo"
	herokuEnvVarVal = "bar"
)

type HerokuTestSuite struct {
	suite.Suite

	ctx    context.Context
	client *heroku.Service
}

func (hs *HerokuTestSuite) SetupTest() {
	client := heroku.NewService(nil)

	hs.ctx = context.Background()
	hs.client = client
}

func (hs *HerokuTestSuite) TestWriteToHerokuSink() {
	t := hs.T()
	a := assert.New(t)
	sink := &HerokuSink{}
	sink.WithHerokuClient(hs.client)
	err := sink.Write(hs.ctx, herokuEnvVar, herokuEnvVarVal)
	a.NoError(err)
}

func TestHerokuSuite(t *testing.T) {
	suite.Run(t, new(HerokuTestSuite))
}
