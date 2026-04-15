package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// LeaseInfo holds metadata about a Vault secret lease.
type LeaseInfo struct {
	LeaseID   string
	Path      string
	ExpiresAt time.Time
	TTL       time.Duration
}

// Client wraps the Vault API client.
type Client struct {
	api *vaultapi.Client
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	api.SetToken(token)

	return &Client{api: api}, nil
}

// LookupLease retrieves lease information for the given lease ID.
func (c *Client) LookupLease(leaseID string) (*LeaseInfo, error) {
	secret, err := c.api.Sys().Lookup(leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup lease %q: %w", leaseID, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data returned for lease %q", leaseID)
	}

	ttlRaw, ok := secret.Data["ttl"]
	if !ok {
		return nil, fmt.Errorf("ttl not found in lease data for %q", leaseID)
	}

	ttlFloat, ok := ttlRaw.(float64)
	if !ok {
		return nil, fmt.Errorf("unexpected ttl type for lease %q", leaseID)
	}

	ttl := time.Duration(ttlFloat) * time.Second
	expiresAt := time.Now().Add(ttl)

	return &LeaseInfo{
		LeaseID:   leaseID,
		ExpiresAt: expiresAt,
		TTL:       ttl,
	}, nil
}

// IsHealthy checks whether the Vault server is reachable and unsealed.
func (c *Client) IsHealthy() error {
	health, err := c.api.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}
	return nil
}
