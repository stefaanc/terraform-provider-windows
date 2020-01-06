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

func dataSourceWindowsNetworkInterface() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "guid": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: tfutil.ValidateUUID(),
                StateFunc: tfutil.StateToUpper(),
            },
            "index": &schema.Schema{
                Type:     schema.TypeInt,   // uint32
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid" },
                ValidateFunc: validation.IntBetween(0, 4294967295),
            },
            "alias": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid", "index" },
            },
            "description": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{"guid", "index", "alias" },
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid", "index", "alias", "description" },
                ValidateFunc: tfutil.ValidateSingleMAC(),
                StateFunc: tfutil.StateToUpper(),
            },
            "network_adapter_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid", "index", "alias", "description", "mac_address" },
            },
            "vnetwork_adapter_name": &schema.Schema{   // only for a net_adapter of the management_os
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid", "index", "alias", "description", "mac_address","network_adapter_name" },
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

        Read:   dataSourceWindowsNetworkInterfaceRead,
    }
}

//------------------------------------------------------------------------------

func dataSourceWindowsNetworkInterfaceRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    guid                := d.Get("guid").(string)
    index               := uint32(d.Get("index").(int))
    alias               := d.Get("alias").(string)
    description         := d.Get("description").(string)
    macAddress          := d.Get("mac_address").(string)
    networkAdapterName  := d.Get("network_adapter_name").(string)
    vnetworkAdapterName := d.Get("vnetwork_adapter_name").(string)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if guid                != "" { id = guid                                  } else
    if index               != 0  { id = strconv.FormatUint(uint64(index), 10) } else
    if alias               != "" { id = alias                                 } else
    if description         != "" { id = description                           } else
    if macAddress          != "" { id = macAddress                            } else
    if networkAdapterName  != "" { id = networkAdapterName                    } else
    if vnetworkAdapterName != "" { id = vnetworkAdapterName                   }
    id = fmt.Sprintf("//%s/network_interfaces/%s", host, id)

    log.Printf("[INFO][terraform-provider-windows] reading windows_network_interface %q\n", id)

    x_lifecycle := tfutil.GetResource(d, "x_lifecycle")

    // read
    niQuery := new(api.NetworkInterface)
    niQuery.GUID                = guid
    niQuery.Index               = index
    niQuery.Alias               = alias
    niQuery.Description         = description
    niQuery.MACAddress          = macAddress
    niQuery.NetworkAdapterName  = networkAdapterName
    niQuery.VNetworkAdapterName = vnetworkAdapterName

    networkInterface, err := c.ReadNetworkInterface(niQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        v, ok := x_lifecycle["ignore_error_if_not_exists"]
        if ok && v.(bool) && strings.Contains(err.Error(), "cannot find network_interface") {
            log.Printf("[INFO][terraform-provider-windows] cannot read windows_network_interface %q\n", id)

            // set zeroed properties
            d.Set("guid", "")
            d.Set("index", 0)
            d.Set("alias", "")
            d.Set("description", "")
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

            log.Printf("[INFO][terraform-provider-windows] ignored error and added zeroed windows_network_interface %q to terraform state\n", id)
            return nil
        }

        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot read windows_network_interface %q\n", id)
        return err
    }

    // set computed lifecycle properties
    x_lifecycle["exists"] = true
    d.Set("x_lifecycle", []interface{}{ x_lifecycle })

    // set properties
    setDataNetworkInterfaceProperties(d, networkInterface)

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] read windows_network_interface %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setDataNetworkInterfaceProperties(d *schema.ResourceData, niProperties *api.NetworkInterface) {
    d.Set("guid", niProperties.GUID)
    d.Set("index", niProperties.Index)
    d.Set("alias", niProperties.Alias)
    d.Set("description", niProperties.Description)
    d.Set("mac_address", niProperties.MACAddress)
    d.Set("network_adapter_name", niProperties.NetworkAdapterName)
    d.Set("vnetwork_adapter_name", niProperties.VNetworkAdapterName)
    d.Set("network_connection_names", niProperties.NetworkConnectionNames)
    d.Set("vswitch_name", niProperties.VSwitchName)
    d.Set("computer_name", niProperties.ComputerName)
}

//------------------------------------------------------------------------------
