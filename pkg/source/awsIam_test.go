package source_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	userName = "test-user"
)

type TestSuite struct {
	suite.Suite

	ctx context.Context

	// aws
	awsClient *cziAws.Client
	mockIAM   *cziAws.MockIAMSvc
	src       *source.AwsIamSource

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

	ts.awsClient = cziAws.New(sess)
	ts.awsClient, ts.mockIAM = ts.awsClient.WithMockIAM()
	ts.src = source.NewAwsIamSource().WithUserName(userName).WithAwsClient(ts.awsClient)

	// mock aws request functionalities
	key := &iam.AccessKey{}
	key.SetAccessKeyId("newAccessKeyId")
	key.SetSecretAccessKey("newSecretAccessKey")
	keyOut := &iam.CreateAccessKeyOutput{}
	keyOut.SetAccessKey(key)
	ts.mockIAM.On("CreateAccessKeyWithContext", mock.Anything).Return(keyOut, nil)

	delOut := &iam.DeleteAccessKeyOutput{}
	ts.mockIAM.On("DeleteAccessKeyWithContext", mock.Anything).Return(delOut, nil)
}

func (ts *TestSuite) TestAwsIamRotateNoKey() {
	t := ts.T()
	r := require.New(t)

	// mock aws list access keys functionality
	keys := &iam.ListAccessKeysOutput{}
	ts.mockIAM.On("ListAccessKeysWithContext", mock.Anything).Return(keys, nil)

	// rotate keys
	newKey, err := ts.src.RotateKeys(ts.ctx)
	r.Nil(err)
	r.NotNil(newKey)
}

func (ts *TestSuite) TestAwsIamRotateOneKey() {
	t := ts.T()
	r := require.New(t)

	// mock aws list access keys functionality
	key := &iam.AccessKeyMetadata{}
	key.SetAccessKeyId("accessKeyId")
	keys := &iam.ListAccessKeysOutput{}
	keys.SetAccessKeyMetadata([]*iam.AccessKeyMetadata{
		key,
	})
	ts.mockIAM.On("ListAccessKeysWithContext", mock.Anything).Return(keys, nil)

	// rotate keys
	newKey, err := ts.src.RotateKeys(ts.ctx)
	r.Nil(err)
	r.NotNil(newKey)
	r.NotEqual(*key.AccessKeyId, *newKey.AccessKeyId)
}

func (ts *TestSuite) TestAwsIamRotateTwoKeysBothOlder() {
	t := ts.T()
	r := require.New(t)

	// mock aws list access keys functionality
	key1 := &iam.AccessKeyMetadata{}
	key1.SetAccessKeyId("accessKeyId1")
	key1.SetCreateDate(time.Now().Add(-1000 * time.Minute))
	key2 := &iam.AccessKeyMetadata{}
	key2.SetAccessKeyId("accessKeyId2")
	key2.SetCreateDate(time.Now().Add(-500 * time.Minute))
	keys := &iam.ListAccessKeysOutput{}
	keys.SetAccessKeyMetadata([]*iam.AccessKeyMetadata{
		key1,
		key2,
	})
	ts.mockIAM.On("ListAccessKeysWithContext", mock.Anything).Return(keys, nil)

	// rotate keys - a new key should be returned
	newKey, err := ts.src.RotateKeys(ts.ctx)
	r.Nil(err)
	r.NotNil(newKey)
	r.NotEqual(*key1.AccessKeyId, *newKey.AccessKeyId)
}

func (ts *TestSuite) TestAwsIamRotateTwoKeysOneOlder() {
	t := ts.T()
	r := require.New(t)

	// mock aws list access keys functionality
	key1 := &iam.AccessKeyMetadata{}
	key1.SetAccessKeyId("accessKeyId1")
	key1.SetCreateDate(time.Now().Add(-1000 * time.Minute))
	key2 := &iam.AccessKeyMetadata{}
	key2.SetAccessKeyId("accessKeyId2")
	key2.SetCreateDate(time.Now())
	keys := &iam.ListAccessKeysOutput{}
	keys.SetAccessKeyMetadata([]*iam.AccessKeyMetadata{
		key1,
		key2,
	})
	ts.mockIAM.On("ListAccessKeysWithContext", mock.Anything).Return(keys, nil)

	// rotate keys - a new key should be returned
	newKey, err := ts.src.RotateKeys(ts.ctx)
	r.Nil(err)
	r.Nil(newKey)
}

func (ts *TestSuite) TestAwsIamRotateTwoKeysNoneOlder() {
	t := ts.T()
	r := require.New(t)

	// mock aws list access keys functionality
	key1 := &iam.AccessKeyMetadata{}
	key1.SetAccessKeyId("accessKeyId1")
	key1.SetCreateDate(time.Now())
	key2 := &iam.AccessKeyMetadata{}
	key2.SetAccessKeyId("accessKeyId2")
	key2.SetCreateDate(time.Now())
	keys := &iam.ListAccessKeysOutput{}
	keys.SetAccessKeyMetadata([]*iam.AccessKeyMetadata{
		key1,
		key2,
	})
	ts.mockIAM.On("ListAccessKeysWithContext", mock.Anything).Return(keys, nil)

	// rotate keys - no key should be createdd
	newKey, err := ts.src.RotateKeys(ts.ctx)
	r.Nil(err)
	r.Nil(newKey)
}

func TestProviderSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
