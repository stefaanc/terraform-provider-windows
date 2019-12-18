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
    "strconv"
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"

    "github.com/stefaanc/terraform-provider-windows/api"
    "github.com/stefaanc/terraform-provider-windows/windows/tfutil"
)

//------------------------------------------------------------------------------

func dataSourceWindowsLinkIPInterface() *schema.Resource {
    return &schema.Resource{
        Read:   dataSourceWindowsLinkIPInterfaceRead,

        Schema: map[string]*schema.Schema{
            "index": &schema.Schema{
                Type:     schema.TypeInt,   // uint32
                Optional: true,
                Computed: true,

                ValidateFunc: validation.IntBetween(0, 4294967295),
            },
            "alias": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index" },
            },
            "description": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "alias" },
            },
            "guid": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "alias", "description" },
                ValidateFunc: tfutil.ValidateUUID(),
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "alias", "description", "guid" },
                ValidateFunc: tfutil.ValidateSingleMAC(),
            },
            "network_adapter_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "alias", "description", "guid", "mac_address" },
            },
            "vnetwork_adapter_name": &schema.Schema{   // only for a net_adapter of the management_os
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "alias", "description", "guid", "mac_address","network_adapter_name" },
            },

            "network_connection_names": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem:     &schema.Schema{ Type: schema.TypeString },
            },
            "vswitch_name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "computer_name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for data sources
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceWindowsLinkIPInterfaceRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    index               := uint32(d.Get("index").(int))
    alias               := d.Get("alias").(string)
    description         := d.Get("description").(string)
    guid                := d.Get("guid").(string)
    macAddress          := d.Get("mac_address").(string)
    networkAdapterName  := d.Get("network_adapter_name").(string)
    vnetworkAdapterName := d.Get("vnetwork_adapter_name").(string)
    x_lifecycle         := tfutil.GetResource(d, "x_lifecycle")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if        index != 0                { id = strconv.FormatUint(uint64(index), 10)
    } else if alias != ""               { id = alias
    } else if description != ""         { id = description
    } else if guid != ""                { id = guid
    } else if macAddress != ""          { id = macAddress
    } else if networkAdapterName != ""  { id = networkAdapterName
    } else if vnetworkAdapterName != "" { id = vnetworkAdapterName
    }
    id = fmt.Sprintf("//%s/interfaces/%s", host, id)

    log.Printf("[INFO][terraform-provider-windows] reading windows_link_ip_interface %q\n", id)

    // read interface
    iQuery := new(api.LinkIPInterface)
    iQuery.Index               = index
    iQuery.Alias               = alias
    iQuery.Description         = description
    iQuery.GUID                = guid
    iQuery.MACAddress          = macAddress
    iQuery.NetworkAdapterName  = networkAdapterName
    iQuery.VNetworkAdapterName = vnetworkAdapterName

    iProperties, err := c.ReadLinkIPInterface(iQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        if x_lifecycle != nil {
            ignore_error_if_not_exists := x_lifecycle["ignore_error_if_not_exists"].(bool)
            if ignore_error_if_not_exists && strings.Contains(err.Error(), "cannot find link_ip_interface") {
                log.Printf("[INFO][terraform-provider-windows] cannot read windows_link_ip_interface %q\n", id)

                // set zeroed properties
                d.Set("index", 0)
                d.Set("alias", "")
                d.Set("description", "")
                d.Set("guid", "")
                d.Set("mac_address", "")
                d.Set("network_adapter_name", "")
                d.Set("vnetwork_adapter_name", "")
                d.Set("network_connection_names", nil)
                d.Set("vswitch_name", "")
                d.Set("computer_name", "")

                // set computed lifecycle properties
                x_lifecycle["exists"] = false
                d.Set("x_lifecycle", []interface{}{ x_lifecycle })

                // set id
                d.SetId(id)

                log.Printf("[INFO][terraform-provider-windows] ignored error and added zeroed windows_link_ip_interface %q to terraform state\n", id)
                return nil
            }
        }

        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot read windows_link_ip_interface %q\n", id)
        return err
    }

    // set properties
    d.Set("index", iProperties.Index)
    d.Set("alias", iProperties.Alias)
    d.Set("description", iProperties.Description)
    d.Set("guid", iProperties.GUID)
    d.Set("mac_address", iProperties.MACAddress)
    d.Set("network_adapter_name", iProperties.NetworkAdapterName)
    d.Set("vnetwork_adapter_name", iProperties.VNetworkAdapterName)
    d.Set("network_connection_names", iProperties.NetworkConnectionNames)
    d.Set("vswitch_name", iProperties.VSwitchName)
    d.Set("computer_name", iProperties.ComputerName)

    // set computed lifecycle properties
    if x_lifecycle != nil {
        x_lifecycle["exists"] = true
        d.Set("x_lifecycle", []interface{}{ x_lifecycle })
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] read windows_link_ip_interface %q\n", id)
    return nil
}

//------------------------------------------------------------------------------
