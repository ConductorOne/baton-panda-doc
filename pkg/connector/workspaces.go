package connector

import (
	"context"
	"time"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type workspaceBuilder struct {
	resourceType *v2.ResourceType
	client       *client.PandaDocClient
}

const wsPageSize = 100

var permissionName = "member"

func (wb *workspaceBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return workspaceResourceType
}

func (wb *workspaceBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	var resources []*v2.Resource
	pToken := &opts.PageToken
	bag, pageToken, err := getToken(pToken, workspaceResourceType)

	if err != nil {
		return nil, nil, err
	}

	workspaces, nextPage, annotation, err := wb.client.ListWorkspaces(ctx, client.PageOptions{
		Page:  pageToken,
		Count: wsPageSize,
	})
	if err != nil {
		return nil, nil, err
	}

	err = bag.Next(nextPage)
	if err != nil {
		return nil, nil, err
	}
	nextPage, err = bag.Marshal()
	if err != nil {
		return nil, nil, err
	}

	if err := cacheWorkspacesPage(ctx, opts, workspaces); err != nil {
		return nil, nil, err
	}

	for _, workspace := range workspaces {
		workspaceResource, err := parseIntoWorkspaceResource(workspace)
		if err != nil {
			return nil, nil, err
		}
		resources = append(resources, workspaceResource)
	}

	return resources, &rs.SyncOpResults{NextPageToken: nextPage, Annotations: annotation}, nil
}

// This function parses a workspace from PandaDoc into a Workspace Resource.
func parseIntoWorkspaceResource(workspace client.Workspace) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"workspace_id": workspace.ID,
		"name":         workspace.Name,
		"owner":        workspace.Owner,
		"date_created": workspace.DateCreated.Format(time.RFC3339),
	}

	groupTraits := []rs.GroupTraitOption{}

	ret, err := rs.NewGroupResource(
		workspace.Name,
		workspaceResourceType,
		workspace.ID,
		groupTraits,
		rs.WithResourceProfile(profile),
	)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (wb *workspaceBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	var entitlements []*v2.Entitlement

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(resource.Description),
		entitlement.WithDisplayName(permissionName),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, permissionName, assigmentOptions...))

	return entitlements, nil, nil
}

func (wb *workspaceBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	var grants []*v2.Grant

	var workspaceId = resource.Id.Resource

	users, err := getUsersFromSession(ctx, wb.client, opts)
	if err != nil {
		return nil, nil, err
	}

	for _, user := range users {
		for _, workspace := range user.Workspaces {
			if workspace.WorkspaceID == workspaceId {
				userResource, _ := parseIntoUserResource(ctx, &user, resource.Id)
				membershipGrant := grant.NewGrant(resource, permissionName, userResource)
				grants = append(grants, membershipGrant)
			}
		}
	}
	return grants, nil, nil
}

func newWorkspaceBuilder(client *client.PandaDocClient) *workspaceBuilder {
	return &workspaceBuilder{
		resourceType: workspaceResourceType,
		client:       client,
	}
}
