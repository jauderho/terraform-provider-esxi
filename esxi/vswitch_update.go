package esxi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVSWITCHUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	log.Println("[resourceVSWITCHUpdate]")

	var uplinks []string
	var err error
	var i int

	name := d.Get("name").(string)
	ports := d.Get("ports").(int)
	mtu := d.Get("mtu").(int)
	link_discovery_mode := d.Get("link_discovery_mode").(string)
	promiscuous_mode := d.Get("promiscuous_mode").(bool)
	mac_changes := d.Get("mac_changes").(bool)
	forged_transmits := d.Get("forged_transmits").(bool)

	// Validate variables
	if ports == 0 {
		ports = 128
	}

	if mtu == 0 {
		mtu = 1500
	}

	if link_discovery_mode == "" {
		link_discovery_mode = "listen"
	}

	if link_discovery_mode != "down" && link_discovery_mode != "listen" &&
		link_discovery_mode != "advertise" && link_discovery_mode != "both" {
		return fmt.Errorf("link_discovery_mode must be one of down, listen, adertise or both")
	}

	uplinkCount, ok := d.Get("uplink.#").(int)
	if !ok {
		uplinkCount = 0
		uplinks[0] = ""
	}
	if uplinkCount > 32 {
		uplinkCount = 32
	}
	for i = 0; i < uplinkCount; i++ {
		prefix := fmt.Sprintf("uplink.%d.", i)

		if attr, ok := d.Get(prefix + "name").(string); ok && attr != "" {
			uplinks = append(uplinks, d.Get(prefix+"name").(string))
		}
	}

	// Do update
	err = vswitchUpdate(c, name, ports, mtu, uplinks, link_discovery_mode, promiscuous_mode, mac_changes, forged_transmits)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Failed to refresh vswitch: %s\n", err)
	}

	// Refresh
	ports, mtu, uplinks, link_discovery_mode, promiscuous_mode, mac_changes, forged_transmits, err = vswitchRead(c, name)
	if err != nil {
		d.SetId("")
		return nil
	}

	// Change uplinks (list) to map
	log.Printf("[resourceVSWITCHUpdate] uplinks: %q\n", uplinks)
	uplink := make([]map[string]interface{}, 0, 1)

	if len(uplinks) == 0 {
		uplink = nil
	} else {
		for i, _ := range uplinks {
			out := make(map[string]interface{})
			out["name"] = uplinks[i]
			uplink = append(uplink, out)
		}
	}
	d.Set("uplink", uplink)

	d.Set("ports", ports)
	d.Set("mtu", mtu)
	d.Set("link_discovery_mode", link_discovery_mode)
	d.Set("promiscuous_mode", promiscuous_mode)
	d.Set("mac_changes", mac_changes)
	d.Set("forged_transmits", forged_transmits)

	return nil
}
