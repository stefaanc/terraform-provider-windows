//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package windows

import (
    "fmt"
    "log"
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

    "github.com/stefaanc/terraform-provider-windows/api"
    "github.com/stefaanc/terraform-provider-windows/windows/tfutil"
)

//------------------------------------------------------------------------------

func dataSourceWindowsNetworkAdapter() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "guid": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: tfutil.ValidateUUID(),
                StateFunc: tfutil.StateToUpper(),
            },

            "name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid" },
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "dns_client": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: dataSourceWindowsNetworkAdapterDNSClient(),
            },

            "admin_status": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "operational_status": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "connection_status": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "connection_speed": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "is_physical": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for persistent resources (similar to data-sources)
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },

        Read:   dataSourceWindowsNetworkAdapterRead,
    }
}

func dataSourceWindowsNetworkAdapterDNSClient() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "register_connection_address": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "register_connection_suffix": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceWindowsNetworkAdapterRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    guid            := d.Get("guid").(string)
    name            := d.Get("name").(string)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if guid    != "" { id = guid    } else
    if name    != "" { id = name    }
    id = fmt.Sprintf("//%s/network_adapters/%s", host, id)

    log.Printf("[INFO][terraform-provider-windows] reading windows_network_adapter %q\n", id)

    x_lifecycle := tfutil.GetResource(d, "x_lifecycle")

    // read
    naQuery := new(api.NetworkAdapter)
    naQuery.GUID    = d.Get("guid").(string)
    naQuery.Name    = d.Get("name").(string)

    networkAdapter, err := c.ReadNetworkAdapter(naQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        v, ok := x_lifecycle["ignore_error_if_not_exists"]
        if ok && v.(bool) && strings.Contains(err.Error(), "cannot find network_adapter") {
            log.Printf("[INFO][terraform-provider-windows] cannot import windows_network_adapter %q into terraform state\n", id)

            // set zeroed properties
            d.Set("guid", "")
            d.Set("name", "")
            d.Set("mac_address", "")
            d.Set("dns_client", nil)
            d.Set("admin_status", "")
            d.Set("operational_status", "")
            d.Set("connection_status", "")
            d.Set("connection_speed", "")
            d.Set("is_physical", false)

            // set computed lifecycle properties
            x_lifecycle["exists"] = false
            d.Set("x_lifecycle", []interface{}{ x_lifecycle })

            // set id
            d.SetId(id)

            log.Printf("[INFO][terraform-provider-windows] ignored error and added zeroed windows_network_adapter %q to terraform state\n", id)
            return nil
        }

        // no lifecycle customizations
        log.Printf("[INFO][terraform-provider-windows] cannot read windows_network_adapter %q\n", id)
        return err
    }

    // set computed lifecycle properties
    x_lifecycle["exists"] = true
    d.Set("x_lifecycle", []interface{}{ x_lifecycle })

    // set properties
    setDataNetworkAdapterProperties(d, networkAdapter)

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] read windows_network_adapter %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setDataNetworkAdapterProperties(d *schema.ResourceData, naProperties *api.NetworkAdapter) {
    d.Set("guid", naProperties.GUID)
    d.Set("name", naProperties.Name)
    d.Set("mac_address", naProperties.MACAddress)

    if len(naProperties.DNSClient) > 0 {
        dnsClient := make(map[string]interface{})
        dnsClient["register_connection_address"] = naProperties.DNSClient[0].RegisterConnectionAddress
        if naProperties.DNSClient[0].RegisterConnectionSuffix == "" {
            // replace "" with "<empty>" (because "" is a valid value and would get skipped in diff since it is equal to the zero value)
            dnsClient["register_connection_suffix"] = "<empty>"
        } else {
            dnsClient["register_connection_suffix"] = naProperties.DNSClient[0].RegisterConnectionSuffix
        }
        d.Set("dns_client", []interface{}{ dnsClient })
    } else {
        d.Set("dns_client", []interface{}{ })
    }

    d.Set("admin_status", naProperties.AdminStatus)
    d.Set("operational_status", naProperties.OperationalStatus)
    d.Set("connection_status", naProperties.ConnectionStatus)
    d.Set("connection_speed", naProperties.ConnectionSpeed)
    d.Set("is_physical", naProperties.IsPhysical)
}

//------------------------------------------------------------------------------
