package connector

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type roleBuilder struct {
	resourceType    *v2.ResourceType
	client          *client.PandaDocClient
	users           []client.User
	usersMutex      sync.RWMutex
	workspaces      []client.Workspace
	workspacesMutex sync.RWMutex
}

func (rb *roleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return rb.resourceType
}

// There is no endpoint for Roles.
// There are 4 system roles and custom roles can be created. We'll retrieve custom roles names from the users list.
func (rb *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rolesResource []*v2.Resource

	for _, role := range systemRoles {
		roleCopy := role
		roleResource, err := parseIntoRoleResource(ctx, &roleCopy, nil)
		if err != nil {
			return nil, "", nil, err
		}
		rolesResource = append(rolesResource, roleResource)
	}

	err := rb.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	systemRoles := []string{"Admin", "Collaborator", "Member", "Manager"}

	users := rb.users
	for _, user := range users {
		for _, workspace := range user.Workspaces {
			if !slices.Contains(systemRoles, workspace.Role) {
				newRole := client.Role{
					Name:        workspace.Role,
					IsSystem:    false,
					Description: "Custom role",
				}
				roleResource, err := parseIntoRoleResource(ctx, &newRole, nil)
				if err != nil {
					return nil, "", nil, err
				}
				rolesResource = append(rolesResource, roleResource)
			}
		}
	}

	return rolesResource, "", nil, nil
}

func (rb *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	err := rb.GetWorkspaces(ctx)

	if err != nil {
		return nil, "", nil, err
	}

	workspaces := rb.workspaces

	for _, workspace := range workspaces {
		permissionName := fmt.Sprintf("assigned in workspace %s", workspace.Name)

		assigmentOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDescription(resource.Description),
			entitlement.WithDisplayName(resource.DisplayName),
		}
		rv = append(rv, entitlement.NewPermissionEntitlement(resource, permissionName, assigmentOptions...))
	}

	return rv, "", nil, nil
}

func (rb *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant

	err := rb.GetUsers(ctx)

	if err != nil {
		return nil, "", nil, err
	}

	users := rb.users
	for _, user := range users {
		for _, workspace := range user.Workspaces {
			if workspace.Role == resource.Id.Resource {
				workspaceName, err := rb.GetWorkspaceName(ctx, workspace.WorkspaceID)
				if err != nil {
					return nil, "", nil, err
				}
				entitlementName := fmt.Sprintf("assigned in workspace %s", workspaceName)
				userResource, _ := parseIntoUserResource(ctx, &user, resource.Id)
				membershipGrant := grant.NewGrant(resource, entitlementName, userResource, grant.WithAnnotation(&v2.V1Identifier{
					Id: fmt.Sprintf("workspace-grant:%s:%s:%s", resource.Id.Resource, workspace.MembershipID, workspace.Role),
				}))
				grants = append(grants, membershipGrant)
			}
		}
	}

	return grants, "", nil, nil
}

func newRolesBuilder(client *client.PandaDocClient) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       client,
	}
}

func parseIntoRoleResource(_ context.Context, role *client.Role, _ *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":        role.Name,
		"name":      role.Name,
		"is_system": role.IsSystem,
	}

	roleTraits := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	ret, err := rs.NewRoleResource(role.Name, roleResourceType, role.Name, roleTraits)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (rb *roleBuilder) GetUsers(ctx context.Context) error {
	rb.usersMutex.RLock()
	defer rb.usersMutex.RUnlock()

	paginationToken := pagination.Token{
		Size:  50,
		Token: "",
	}

	if rb.users != nil {
		return nil
	}

	for {
		bag, pageToken, err := getToken(&paginationToken, userResourceType)
		if err != nil {
			return err
		}
		users, nextPageToken, _, err := rb.client.ListUsers(ctx, client.PageOptions{
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

		rb.users = append(rb.users, users...)
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

func (rb *roleBuilder) GetWorkspaces(ctx context.Context) error {
	rb.workspacesMutex.RLock()
	defer rb.workspacesMutex.RUnlock()

	paginationToken := pagination.Token{
		Size:  50,
		Token: "",
	}

	if rb.workspaces != nil {
		return nil
	}

	for {
		bag, pageToken, err := getToken(&paginationToken, workspaceResourceType)
		if err != nil {
			return err
		}
		workspaces, nextPageToken, _, err := rb.client.ListWorkspaces(ctx, client.PageOptions{
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

		rb.workspaces = append(rb.workspaces, workspaces...)
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

func (rb *roleBuilder) GetWorkspaceName(ctx context.Context, workspaceID string) (string, error) {
	err := rb.GetWorkspaces(ctx)

	if err != nil {
		return "", err
	}
	workspaceMap := make(map[string]string)
	for _, w := range rb.workspaces {
		workspaceMap[w.ID] = w.Name
	}
	if name, exists := workspaceMap[workspaceID]; exists {
		return name, nil
	} else {
		return "", fmt.Errorf("provided workspace id %s does not exists", workspaceID)
	}
}
