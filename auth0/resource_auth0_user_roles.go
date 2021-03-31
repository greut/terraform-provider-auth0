package auth0

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"gopkg.in/auth0.v5"
	"gopkg.in/auth0.v5/management"
)

func newUserRoles() *schema.Resource {
	return &schema.Resource{
		Create: createUserRoles,
		Read:   readUserRoles,
		Update: updateUserRoles,
		Delete: deleteUserRoles,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func readUserRoles(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)

	l, err := api.User.Roles(d.Id())
	if err != nil {
		return err
	}
	d.Set("roles", func() (v []interface{}) {
		for _, role := range l.Roles {
			v = append(v, auth0.StringValue(role.ID))
		}
		return
	}())

	return nil
}

func createUserRoles(d *schema.ResourceData, m interface{}) error {
	userID := d.Get("user_id").(string)
	d.SetId(userID)

	d.Partial(true)
	err := assignUserRoles(d, m)
	if err != nil {
		return fmt.Errorf("failed creating user roles. %w", err)
	}
	d.Partial(false)

	return readUserRoles(d, m)
}

func updateUserRoles(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	err := assignUserRoles(d, m)
	if err != nil {
		return fmt.Errorf("failed assigning user roles. %w", err)
	}
	d.Partial(false)

	return readUserRoles(d, m)
}

func deleteUserRoles(d *schema.ResourceData, m interface{}) error {
	roles := d.Get("roles").(*schema.Set)
	if roles.Len() == 0 {
		return nil
	}

	rmRoles := make([]*management.Role, roles.Len())

	for i, rmRole := range roles.List() {
		rmRoles[i] = &management.Role{
			ID: auth0.String(rmRole.(string)),
		}
	}

	api := m.(*management.Management)

	err := api.User.RemoveRoles(d.Id(), rmRoles)
	if err != nil {
		// Ignore 404 errors as the role may have been deleted prior to
		// unassigning them from the user.
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() != http.StatusNotFound {
				return err
			}
		} else {
			return err
		}
	}

	return err
}
