package connector

import (
	"context"
	"io"
	"strconv"
	"sync"
	"time"

	cfg "github.com/conductorone/baton-panda-doc/pkg/config"
	"github.com/conductorone/baton-panda-doc/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

const TTL = 5 // in minutes
const usersPageSize = 50

type Connector struct {
	client         *client.PandaDocClient
	cachedUsers    []client.User
	cacheTimestamp time.Time
	usersMtx       sync.Mutex
}

func (c *Connector) cacheUsers(ctx context.Context) error {
	c.usersMtx.Lock()
	defer c.usersMtx.Unlock()

	if c.cachedUsers != nil && time.Since(c.cacheTimestamp) < TTL*time.Minute {
		return nil
	}

	var usersToCache []client.User
	pageOptions := client.PageOptions{
		Count: usersPageSize,
		Page:  1,
	}

	for {
		users, nextPageToken, _, err := c.client.ListUsers(ctx, pageOptions)
		if err != nil {
			return err
		}
		usersToCache = append(usersToCache, users...)

		if nextPageToken == "" {
			break
		} else {
			pageToken, err := strconv.Atoi(nextPageToken)
			if err != nil {
				return err
			}
			pageOptions.Page = pageToken
		}
	}

	c.cachedUsers = usersToCache
	c.cacheTimestamp = time.Now()
	return nil
}

// ResourceSyncers returns a ResourceSyncerV2 for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(d.client, d),
		newWorkspaceBuilder(d.client, d),
		newRolesBuilder(d.client, d),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "PandaDoc connector",
		Description: "Connector to sync users, workspaces, and roles from PandaDoc.",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, europeDomain bool, apiKey string, baseURL string) (*Connector, error) {
	pandaDocClient, err := client.New(
		ctx,
		client.WithEuropeDomain(europeDomain),
		client.WithBearerToken(apiKey),
		client.WithBaseURL(baseURL),
	)

	if err != nil {
		return nil, err
	}

	return &Connector{client: pandaDocClient}, nil
}

// NewLambdaConnector creates a new connector from config for use with lambda/containerized deployments.
func NewLambdaConnector(ctx context.Context, ac *cfg.PandaDoc, _ *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	cb, err := New(ctx, ac.EuropeDomain, ac.ApiKey, ac.BaseUrl)
	if err != nil {
		return nil, nil, err
	}
	return cb, nil, nil
}
