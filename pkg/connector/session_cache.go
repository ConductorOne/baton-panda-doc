package connector

import (
	"context"
	"strconv"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/session"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sdk/pkg/types/sessions"
)

const (
	usersCachePrefix      = "users"
	workspacesCachePrefix = "workspaces"
)

// getUsersFromSession returns all users, reading from the session store when
// available and falling back to the API on a miss. On an API fetch the results
// are written back to the session store so subsequent callers get cache hits
// regardless of which builder runs first during the List phase.
func getUsersFromSession(ctx context.Context, c *client.PandaDocClient, opts rs.SyncOpAttrs) ([]client.User, error) {
	if opts.Session != nil {
		cached, err := session.GetAllJSON[client.User](ctx, opts.Session,
			sessions.WithSyncID(opts.SyncID),
			sessions.WithPrefix(usersCachePrefix),
		)
		if err != nil {
			return nil, err
		}
		if len(cached) > 0 {
			users := make([]client.User, 0, len(cached))
			for _, u := range cached {
				users = append(users, u)
			}
			return users, nil
		}
	}

	// Cache miss or no session: fetch from API.
	users, err := fetchAllUsers(ctx, c)
	if err != nil {
		return nil, err
	}

	if opts.Session != nil && len(users) > 0 {
		toCache := make(map[string]client.User, len(users))
		for _, u := range users {
			toCache[u.ID] = u
		}
		if err := session.SetManyJSON(ctx, opts.Session, toCache,
			sessions.WithSyncID(opts.SyncID),
			sessions.WithPrefix(usersCachePrefix),
		); err != nil {
			return nil, err
		}
	}

	return users, nil
}

func fetchAllUsers(ctx context.Context, c *client.PandaDocClient) ([]client.User, error) {
	var all []client.User
	pageOptions := client.PageOptions{
		Count: usersPageSize,
		Page:  1,
	}
	for {
		users, nextPageToken, _, err := c.ListUsers(ctx, pageOptions)
		if err != nil {
			return nil, err
		}
		all = append(all, users...)
		if nextPageToken == "" {
			break
		}
		pageToken, err := strconv.Atoi(nextPageToken)
		if err != nil {
			return nil, err
		}
		pageOptions.Page = pageToken
	}
	return all, nil
}

// getWorkspacesFromSession returns all workspaces, reading from the session
// store when available. On a miss it fetches all pages from the API and writes
// them back so subsequent callers hit the cache.
func getWorkspacesFromSession(ctx context.Context, c *client.PandaDocClient, opts rs.SyncOpAttrs) ([]client.Workspace, error) {
	if opts.Session != nil {
		cached, err := session.GetAllJSON[client.Workspace](ctx, opts.Session,
			sessions.WithSyncID(opts.SyncID),
			sessions.WithPrefix(workspacesCachePrefix),
		)
		if err != nil {
			return nil, err
		}
		if len(cached) > 0 {
			workspaces := make([]client.Workspace, 0, len(cached))
			for _, w := range cached {
				workspaces = append(workspaces, w)
			}
			return workspaces, nil
		}
	}

	// Cache miss or no session: fetch from API.
	workspaces, err := fetchAllWorkspaces(ctx, c)
	if err != nil {
		return nil, err
	}

	if opts.Session != nil && len(workspaces) > 0 {
		if err := cacheWorkspacesPage(ctx, opts, workspaces); err != nil {
			return nil, err
		}
	}

	return workspaces, nil
}

// cacheWorkspacesPage writes a slice of workspaces into the session store.
// Called from workspaceBuilder.List on each fetched page so the cache is
// populated incrementally as the List phase runs.
func cacheWorkspacesPage(ctx context.Context, opts rs.SyncOpAttrs, workspaces []client.Workspace) error {
	if opts.Session == nil || len(workspaces) == 0 {
		return nil
	}
	toCache := make(map[string]client.Workspace, len(workspaces))
	for _, w := range workspaces {
		toCache[w.ID] = w
	}
	return session.SetManyJSON(ctx, opts.Session, toCache,
		sessions.WithSyncID(opts.SyncID),
		sessions.WithPrefix(workspacesCachePrefix),
	)
}

func fetchAllWorkspaces(ctx context.Context, c *client.PandaDocClient) ([]client.Workspace, error) {
	var all []client.Workspace
	pageOptions := client.PageOptions{
		Count: wsPageSize,
		Page:  1,
	}
	for {
		workspaces, nextPageToken, _, err := c.ListWorkspaces(ctx, pageOptions)
		if err != nil {
			return nil, err
		}
		all = append(all, workspaces...)
		if nextPageToken == "" {
			break
		}
		pageToken, err := strconv.Atoi(nextPageToken)
		if err != nil {
			return nil, err
		}
		pageOptions.Page = pageToken
	}
	return all, nil
}
