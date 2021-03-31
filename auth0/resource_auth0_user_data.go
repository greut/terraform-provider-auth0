package auth0

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"gopkg.in/auth0.v5/management"
)

func newUserData() *schema.Resource {
	return &schema.Resource{
		Create: createUserData,
		Read:   readUserData,
		Update: updateUserData,
		Delete: deleteUserData,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"user_metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.ValidateJsonString,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
			"app_metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.ValidateJsonString,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
		},
	}
}

func readUserData(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	u, err := api.User.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	userMeta, err := structure.FlattenJsonToString(u.UserMetadata)
	if err != nil {
		return err
	}
	d.Set("user_metadata", userMeta)

	appMeta, err := structure.FlattenJsonToString(u.AppMetadata)
	if err != nil {
		return err
	}
	d.Set("app_metadata", appMeta)

	return nil
}

func createUserData(d *schema.ResourceData, m interface{}) error {
	userID := d.Get("user_id").(string)
	d.SetId(userID)

	return updateUserData(d, m)
}

func updateUserData(d *schema.ResourceData, m interface{}) error {
	u, err := buildUserData(d)
	if err != nil {
		return err
	}

	api := m.(*management.Management)

	if userHasChange(u) {
		if err := api.User.Update(d.Id(), u); err != nil {
			return err
		}
	}

	return readUserData(d, m)
}

func deleteUserData(d *schema.ResourceData, m interface{}) error {
	u, err := buildUserData(d)
	if err != nil {
		return err
	}

	// Empty the Metadata maps as they are merged during Update.

	if u.AppMetadata != nil {
		for k := range u.AppMetadata {
			u.AppMetadata[k] = nil
		}
	}

	if u.UserMetadata != nil {
		for k := range u.UserMetadata {
			u.UserMetadata[k] = nil
		}
	}

	api := m.(*management.Management)

	return api.User.Update(d.Id(), u)
}

func buildUserData(d *schema.ResourceData) (u *management.User, err error) {
	u = new(management.User)

	u.UserMetadata, err = JSON(d, "user_metadata")
	if err != nil {
		return nil, err
	}

	u.AppMetadata, err = JSON(d, "app_metadata")
	if err != nil {
		return nil, err
	}

	return u, nil
}
