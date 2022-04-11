package secret_manager

import (
	"context"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	"github.com/googleapis/gax-go"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

type client interface {
	AccessSecretVersion(
		ctx context.Context,
		req *secretmanagerpb.AccessSecretVersionRequest,
		opts ...gax.CallOption,
	) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

type SecretManager struct {
	client  client
	project string
	cache   map[string]string
}

// NewSecretManager creates a new SecretManager instance to acess secrets.
func NewSecretManager() (*SecretManager, error) {
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)

	if err != nil {
		return nil, err
	}

	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeReadonlyScope)

	if err != nil {
		return nil, err
	}

	return &SecretManager{
		client:  client,
		project: credentials.ProjectID,
		cache:   map[string]string{},
	}, nil
}

// AccessSecrets gets the secrets from environment values.
// It obtains the secret of each environment value if it has his value equal to `@secret`.
func (sm *SecretManager) AccessSecrets() error {
	for _, element := range os.Environ() {
		env := strings.Split(element, "=")

		key := env[0]
		val := env[1]

		if val == "@secret" {
			data, _ := sm.AccessSecret(key)

			err := os.Setenv(key, data)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AccessSecret get a secret value from a name.
func (sm *SecretManager) AccessSecret(name string) (string, error) {
	if value, ok := sm.cache[name]; ok {
		return value, nil
	}

	sname := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sm.project, name)

	ctx := context.Background()
	vr := &secretmanagerpb.AccessSecretVersionRequest{Name: sname}

	secret, err := sm.client.AccessSecretVersion(ctx, vr)

	if err != nil {
		return "", err
	}

	value := string(secret.Payload.Data[:])

	sm.cache[name] = value

	return value, nil
}

// Close
func (sm *SecretManager) Close() error {
	return sm.client.Close()
}
