package connector

import (
	"context"
	"sync"
	"time"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
)

type workspaceBuilder struct {
	resourceType *v2.ResourceType
	client       *client.PandaDocClient
	users        []client.User
	usersMutex   sync.RWMutex
}

var permissionName = "member"

func (wb *workspaceBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return workspaceResourceType
}

func (wb *workspaceBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource
	bag, pageToken, err := getToken(pToken, workspaceResourceType)

	if err != nil {
		return nil, "", nil, err
	}

	workspaces, nextPage, annotation, err := wb.client.ListWorkspaces(ctx, client.PageOptions{
		Page:  pageToken,
		Count: pToken.Size,
	})
	if err != nil {
		return nil, "", nil, err
	}

	err = bag.Next(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	for _, workspace := range workspaces {
		workspaceResource, err := parseIntoWorkspaceResource(workspace)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, workspaceResource)
	}

	return resources, "", annotation, nil
}

// This function parses a workspace from PandaDoc into a Workspace Resource.
func parseIntoWorkspaceResource(workspace client.Workspace) (*v2.Resource, error) {

	profile := map[string]interface{}{
		"workspace_id": workspace.ID,
		"name":         workspace.Name,
		"owner":        workspace.Owner,
		"date_created": workspace.DateCreated.Format(time.RFC3339),
	}

	groupTraits := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}

	ret, err := resource.NewGroupResource(
		workspace.Name,
		workspaceResourceType,
		workspace.ID,
		groupTraits,
	)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (wb *workspaceBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(resource.Description),
		entitlement.WithDisplayName(permissionName),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, permissionName, assigmentOptions...))

	return entitlements, "", nil, nil
}

func (wb *workspaceBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant

	var workspaceId = resource.Id.Resource

	err := wb.GetUsers(ctx)

	if err != nil {
		return nil, "", nil, err
	}

	users := wb.users

	for _, user := range users {
		for _, workspace := range user.Workspaces {
			if workspace.WorkspaceID == workspaceId {
				userResource, _ := parseIntoUserResource(ctx, &user, resource.Id)
				membershipGrant := grant.NewGrant(resource, permissionName, userResource)
				grants = append(grants, membershipGrant)
			}
		}

	}

	return grants, "", nil, nil
}

func newWorkspaceBuilder(client *client.PandaDocClient) *workspaceBuilder {
	return &workspaceBuilder{
		resourceType: workspaceResourceType,
		client:       client,
	}
}

func (wb *workspaceBuilder) GetUsers(ctx context.Context) error {
	wb.usersMutex.RLock()
	defer wb.usersMutex.RUnlock()

	paginationToken := pagination.Token{
		Size:  50,
		Token: "",
	}

	if wb.users != nil {
		return nil
	}

	for {
		bag, pageToken, err := getToken(&paginationToken, userResourceType)
		if err != nil {
			return err
		}
		users, nextPageToken, _, err := wb.client.ListUsers(ctx, client.PageOptions{
			Count: paginationToken.Size,
			Page:  pageToken,
		})
		if err != nil {
			return err
		}
		err = bag.Next(nextPageToken)
		if err != nil {
			return err
		}

		wb.users = append(wb.users, users...)
		nextPageToken, err = bag.Marshal()
		if err != nil {
			return err
		}
		if nextPageToken == "" {
			break
		}
		paginationToken.Token = nextPageToken
	}

	return nil
}
