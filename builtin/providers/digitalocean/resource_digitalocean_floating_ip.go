package digitalocean

import (
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDigitalOceanFloatingIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceFloatingIPCreate,
		Read:   resourceFloatingIPRead,
		Update: resourceFloatingIPUpdate,
		Delete: resourceFloatingIPDelete,

		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"droplet_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceFloatingIPCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*godo.Client)

	region := d.Get("region").(string)
	dropletID := d.Get("droplet_id").(int)

	if region != "" && dropletID > 0 {
		return errors.New("Region and droplet_id are mutually exclusive")
	}

	var fcr godo.FloatingIPCreateRequest

	if region != "" {
		fcr.Region = region
	} else {
		fcr.DropletID = dropletID
	}

	fip, _, err := client.FloatingIPs.Create(&fcr)
	if err != nil {
		return err
	}

	d.SetId(fip.IP)
	d.Set("region", fip.Region.Slug)
	if fip.Droplet != nil {
		d.Set("droplet_id", fip.Droplet.ID)
	}
	return nil
}

func resourceFloatingIPRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*godo.Client)

	fip, _, err := client.FloatingIPs.Get(d.Id())
	if err != nil {
		// If fip does not exist, mark it as gone.
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving floating ip: %s", err)
	}

	d.Set("region", fip.Region.Slug)

	if fip.Droplet != nil {
		d.Set("droplet_id", fip.Droplet.ID)
	}

	return nil
}

func resourceFloatingIPUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*godo.Client)

	if d.HasChange("droplet_id") {
		_, newDropletID := d.GetChange("droplet_id")
		if id := newDropletID.(int); id > 0 {
			_, _, err := client.FloatingIPActions.Assign(d.Id(), id)
			if err != nil {
				return fmt.Errorf("Error assigning droplet (%s) to %s", newDropletID, d.Id())
			}
		} else {
			_, _, err := client.FloatingIPActions.Unassign(d.Id())
			if err != nil {
				return fmt.Errorf("Error unassigning droplet from %s", d.Id())
			}
		}
	}

	return nil
}

func resourceFloatingIPDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*godo.Client)

	_, err := client.FloatingIPs.Delete(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Floating IP: %s", err)
	}

	return nil
}
