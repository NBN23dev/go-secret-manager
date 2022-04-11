package secret_manager

import (
	"context"
	"fmt"
	"os"
	"testing"

	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/googleapis/gax-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedClient struct {
	mock.Mock
}

func (mc *mockedClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	args := mc.Called(ctx, req)

	return args.Get(0).(*secretmanagerpb.AccessSecretVersionResponse), args.Error(1)
}

func (mc *mockedClient) Close() error {
	args := mc.Called()

	return args.Error(0)
}

// TestAccessSecret
func TestAccessSecret(t *testing.T) {
	mc := mockedClient{}
	mp := "test-dev"

	mk := "foo"
	mv := "bar"

	mName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", mp, mk)

	asvr := secretmanagerpb.AccessSecretVersionResponse{
		Name: mName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(mv),
		},
	}

	mctx := context.Background()
	mReq := secretmanagerpb.AccessSecretVersionRequest{
		Name: mName,
	}

	mc.On("AccessSecretVersion", mctx, &mReq).Return(&asvr, nil)

	sm := SecretManager{client: &mc, project: mp, cache: map[string]string{}}

	// Tested function
	secret, err := sm.AccessSecret(mk)

	assert.NoError(t, err)
	assert.Equal(t, secret, mv)

	mc.AssertExpectations(t)
}

// TestAccessSecrets
func TestAccessSecrets(t *testing.T) {
	t.Cleanup(func() {
		os.Clearenv()
	})

	mc := mockedClient{}
	mp := "test-dev"

	mk := "foo"
	mv := "bar"

	os.Setenv(mk, "@secret")

	mName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", mp, mk)

	asvr := secretmanagerpb.AccessSecretVersionResponse{
		Name: mName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(mv),
		},
	}

	mctx := context.Background()
	mReq := secretmanagerpb.AccessSecretVersionRequest{
		Name: mName,
	}

	mc.On("AccessSecretVersion", mctx, &mReq).Return(&asvr, nil)

	sm := SecretManager{client: &mc, project: mp, cache: map[string]string{}}

	// Tested function
	err := sm.AccessSecrets()

	assert.NoError(t, err)
	assert.Equal(t, os.Getenv(mk), mv)

	fmt.Println(os.Getenv(mk))

	mc.AssertExpectations(t)
}

// TestClose
func TestClose(t *testing.T) {
	mc := mockedClient{}

	mc.On("Close").Return(nil)

	sm := SecretManager{client: &mc}

	// Tested function
	err := sm.Close()

	assert.NoError(t, err)

	mc.AssertExpectations(t)
}
