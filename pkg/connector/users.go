package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"

	"github.com/conductorone/baton-panda-doc/pkg/client"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.PandaDocClient
	connector    *Connector
}

func (ub *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
// Users are also written to the session store so that roles and workspaces
// builders can read them without re-fetching from the API.
func (ub *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	users, err := getUsersFromSession(ctx, ub.client, opts)
	if err != nil {
		return nil, nil, err
	}

	resources := make([]*v2.Resource, 0, len(users))
	for _, user := range users {
		userCopy := user
		userResource, err := parseIntoUserResource(ctx, &userCopy, nil)
		if err != nil {
			return nil, nil, err
		}
		resources = append(resources, userResource)
	}

	return resources, nil, nil
}

func parseIntoUserResource(_ context.Context, user *client.User, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	var userStatus = v2.UserTrait_Status_STATUS_ENABLED

	profile := map[string]interface{}{
		"user_id":    user.ID,
		"first_name": user.FirstName,
		"last_name":  user.Lastame,
		"email":      user.Email,
		"phone":      user.Phone,
		"license":    user.License,
		"owner":      user.IsOrganizationOwner,
	}

	userTraits := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithStatus(userStatus),
		rs.WithEmail(user.Email, true),
	}

	displayName := user.Email

	ret, err := rs.NewUserResource(
		displayName,
		userResourceType,
		user.ID,
		userTraits,
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// Entitlements always returns an empty slice for users.
func (ub *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (ub *userBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

func newUserBuilder(c *client.PandaDocClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       c,
	}
}
