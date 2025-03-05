package connector

import (
	"context"
	"sync"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.PandaDocClient
	users        []client.User
	usersMutex   sync.RWMutex
}

func (ub *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (ub *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	nextPageToken, annotation, err := ub.GetUsers(ctx, pToken)

	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range ub.users {
		userCopy := user
		userResource, err := parseIntoUserResource(ctx, &userCopy, nil)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, userResource)
	}

	return resources, nextPageToken, annotation, nil
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

	userTraits := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithStatus(userStatus),
		resource.WithEmail(user.Email, true),
	}

	displayName := user.Email

	ret, err := resource.NewUserResource(
		displayName,
		userResourceType,
		user.ID,
		userTraits,
		resource.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// Entitlements always returns an empty slice for users.
func (ub *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (ub *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(c *client.PandaDocClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       c,
	}
}

func (ub *userBuilder) GetUsers(ctx context.Context, pToken *pagination.Token) (string, annotations.Annotations, error) {
	ub.usersMutex.RLock()
	defer ub.usersMutex.RUnlock()

	if ub.users != nil {
		return "", nil, nil
	}

	bag, pageToken, err := getToken(pToken, userResourceType)
	if err != nil {
		return "", nil, err
	}

	users, nextPageToken, _, err := ub.client.ListUsers(ctx, client.PageOptions{
		Count: pToken.Size,
		Page:  pageToken,
	})
	if err != nil {
		return "", nil, err
	}
	err = bag.Next(nextPageToken)
	if err != nil {
		return "", nil, err
	}

	ub.users = users
	nextPageToken, err = bag.Marshal()
	if err != nil {
		return "", nil, err
	}

	return nextPageToken, nil, nil
}
