package transit

import (
	"context"

	"github.com/hashicorp/go-hclog"
	vault "github.com/hashicorp/vault/api"
)

type TransitClient struct {
	logger hclog.Logger

	client *vault.Client

	ctx context.Context

	// key properties
	transitMount string
	keyName      string
}

func NewTransitClient(l hclog.Logger) (*TransitClient, error) {

	ctx := context.Background()

	client, err := NewVaultClient(l)
	if err != nil {
		return nil, err
	}

	return &TransitClient{
		logger: l,
		client: client,
		ctx:    ctx,
	}, nil

}

func (t *TransitClient) SetKeyProperties(transitMount, keyName string) {

	t.keyName = keyName
	t.transitMount = transitMount
}
