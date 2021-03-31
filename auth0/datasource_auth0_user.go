package auth0

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"

	"gopkg.in/auth0.v5"
	"gopkg.in/auth0.v5/management"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"connection_name", "email"},
			},
			"connection_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_id"},
			},
			"email": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_id"},
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"family_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"given_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"nickname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"email_verified": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"verify_email": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"phone_verified": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"user_metadata": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"app_metadata": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"picture": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		Read: dataSourceUserRead,
	}
}

func dataSourceUserRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)

	userID, useUserID := d.GetOk("user_id")
	connectionName, useConnectionName := d.GetOk("connection_name")
	email, useEmail := d.GetOk("email")

	var user *management.User
	if useUserID {
		u, err := api.User.Read(userID.(string))
		if err != nil {
			return err
		}

		user = u
	} else if useConnectionName && useEmail {
		users, err := api.User.ListByEmail(email.(string))
		if err != nil {
			return err
		}
		for _, u := range users {
			if u.Connection != nil && *u.Connection == connectionName.(string) {
				user = u
				break
			}
		}
	} else {
		return fmt.Errorf("badly configured `auth0_user`")
	}

	if user == nil {
		return fmt.Errorf("Missing user.")
	}

	d.Set("user_id", user.ID)
	d.Set("username", user.Username)
	d.Set("name", user.Name)
	d.Set("family_name", user.FamilyName)
	d.Set("given_name", user.GivenName)
	d.Set("nickname", user.Nickname)
	d.Set("email", user.Email)
	d.Set("email_verified", user.EmailVerified)
	d.Set("verify_email", user.VerifyEmail)
	d.Set("phone_number", user.PhoneNumber)
	d.Set("phone_verified", user.PhoneVerified)
	d.Set("blocked", user.Blocked)
	d.Set("picture", user.Picture)

	userMeta, err := structure.FlattenJsonToString(user.UserMetadata)
	if err != nil {
		return err
	}
	d.Set("user_metadata", userMeta)

	appMeta, err := structure.FlattenJsonToString(user.AppMetadata)
	if err != nil {
		return err
	}
	d.Set("app_metadata", appMeta)

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
