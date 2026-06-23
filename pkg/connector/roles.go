package connector

import (
	"context"
	"fmt"
	"slices"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.PandaDocClient
}

func (rb *roleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return rb.resourceType
}

// There is no endpoint for Roles.
// There are 4 system roles and custom roles can be created. We'll retrieve custom roles names from the users list.
func (rb *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	var rolesResource []*v2.Resource

	for _, role := range systemRoles {
		roleCopy := role
		roleResource, err := parseIntoRoleResource(ctx, &roleCopy, nil)
		if err != nil {
			return nil, nil, err
		}
		rolesResource = append(rolesResource, roleResource)
	}

	users, err := getUsersFromSession(ctx, rb.client, opts)
	if err != nil {
		return nil, nil, err
	}

	systemRoleNames := []string{"Admin", "Collaborator", "Member", "Manager"}

	for _, user := range users {
		for _, workspace := range user.Workspaces {
			if !slices.Contains(systemRoleNames, workspace.Role) {
				newRole := client.Role{
					Name:        workspace.Role,
					IsSystem:    false,
					Description: "Custom role",
				}
				roleResource, err := parseIntoRoleResource(ctx, &newRole, nil)
				if err != nil {
					return nil, nil, err
				}
				rolesResource = append(rolesResource, roleResource)
			}
		}
	}

	return rolesResource, nil, nil
}

func (rb *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	var rv []*v2.Entitlement

	workspaces, err := getWorkspacesFromSession(ctx, rb.client, opts)
	if err != nil {
		return nil, nil, err
	}

	for _, workspace := range workspaces {
		permissionName := fmt.Sprintf("assigned in workspace %s", workspace.Name)

		assigmentOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDescription(resource.Description),
			entitlement.WithDisplayName(resource.DisplayName),
		}
		rv = append(rv, entitlement.NewPermissionEntitlement(resource, permissionName, assigmentOptions...))
	}

	return rv, nil, nil
}

func (rb *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	var grants []*v2.Grant

	users, err := getUsersFromSession(ctx, rb.client, opts)
	if err != nil {
		return nil, nil, err
	}

	workspaces, err := getWorkspacesFromSession(ctx, rb.client, opts)
	if err != nil {
		return nil, nil, err
	}
	wsNames := make(map[string]string, len(workspaces))
	for _, w := range workspaces {
		wsNames[w.ID] = w.Name
	}

	for _, user := range users {
		for _, workspace := range user.Workspaces {
			if workspace.Role == resource.Id.Resource {
				entitlementName := fmt.Sprintf("assigned in workspace %s", wsNames[workspace.WorkspaceID])
				userResource, _ := parseIntoUserResource(ctx, &user, resource.Id)
				membershipGrant := grant.NewGrant(resource, entitlementName, userResource, grant.WithAnnotation(&v2.V1Identifier{
					Id: fmt.Sprintf("workspace-grant:%s:%s:%s", resource.Id.Resource, workspace.MembershipID, workspace.Role),
				}))
				grants = append(grants, membershipGrant)
			}
		}
	}

	return grants, nil, nil
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

