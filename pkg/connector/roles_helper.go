package connector

import (
	"github.com/conductorone/baton-panda-doc/pkg/client"
)

var systemRoles = []client.Role{
	{
		Name:     "Member",
		IsSystem: true,
	},
	{
		Name:     "Manager",
		IsSystem: true,
	},
	{
		Name:     "Admin",
		IsSystem: true,
	},
	{
		Name:     "Collaborator",
		IsSystem: true,
	},
}
