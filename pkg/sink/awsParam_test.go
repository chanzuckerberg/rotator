package sink_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	awsMocks "github.com/chanzuckerberg/go-misc/aws/mocks"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	parName     = "test-parameter"
	parValue    = "value"
	fakeParName = "non-existing parameter"
)

type TestSuite struct {
	suite.Suite

	ctx context.Context

	// aws
	awsClient          *cziAws.Client
	mockSSM            *awsMocks.MockSSMAPI
	mockSecretsManager *awsMocks.MockSecretsManagerAPI
	sink               sink.Sink

	// cleanup
	server *httptest.Server
}

func (ts *TestSuite) TearDownTest() {
	ts.server.Close()
}

func (ts *TestSuite) SetupTest() {
	ts.ctx = context.Background()

	sess, server := cziAws.NewMockSession()
	ts.server = server

	controller := gomock.NewController(ts.T())
	ts.awsClient = cziAws.New(sess)
	ts.awsClient, ts.mockSSM = ts.awsClient.WithMockSSM(controller)
	ts.awsClient, ts.mockSecretsManager = ts.awsClient.WithMockSecretsManager(controller)

	// mock PutParameterWithContext
	out := &ssm.PutParameterOutput{}
	ts.mockSSM.EXPECT().PutParameterWithContext(gomock.Any(), gomock.Any()).Return(out, nil)
}

func (ts *TestSuite) TestWriteToAwsParamSinkFakeParam() {
	t := ts.T()
	r := require.New(t)

	// mock GetParameterWithContext for non-existing parameter
	in := &ssm.GetParameterInput{}
	in.SetName(fakeParName)
	out := &ssm.GetParameterOutput{}
	errNotFound := awserr.New(ssm.ErrCodeParameterNotFound, "", nil)
	ts.mockSSM.EXPECT().GetParameterWithContext(gomock.Any(), gomock.Eq(in)).Return(out, errNotFound)

	// write secret to sink
	ts.sink = &sink.AwsParamSink{Client: ts.awsClient}
	err := ts.sink.Write(ts.ctx, fakeParName, parValue)
	r.NotNil(err)
}

func (ts *TestSuite) TestWriteToAwsParamSink() {
	t := ts.T()
	r := require.New(t)

	// mock GetParameterWithContext for existing parameter
	in := &ssm.GetParameterInput{}
	in.SetName(parName)
	par := &ssm.Parameter{}
	par.SetName(parName).SetValue(parValue)
	out := &ssm.GetParameterOutput{}
	out.SetParameter(par)
	ts.mockSSM.EXPECT().GetParameterWithContext(gomock.Any(), gomock.Eq(in)).Return(out, nil)

	// write secret to sink
	ts.sink = &sink.AwsParamSink{Client: ts.awsClient}
	err := ts.sink.Write(ts.ctx, parName, parValue)
	r.Nil(err)
}

func TestProviderSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
