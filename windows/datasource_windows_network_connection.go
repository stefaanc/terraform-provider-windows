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
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"

    "github.com/stefaanc/terraform-provider-windows/api"
    "github.com/stefaanc/terraform-provider-windows/windows/tfutil"
)

//------------------------------------------------------------------------------

func dataSourceWindowsNetworkConnection() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "guid": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: tfutil.ValidateUUID(),
                StateFunc: tfutil.StateToUpper(),
            },

            "ipv4_gateway_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid" },
                ValidateFunc: validation.SingleIP(),
            },
            "ipv6_gateway_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid", "ipv4_gateway_address" },
                ValidateFunc: validation.SingleIP(),
                StateFunc: tfutil.StateToLower(),
            },

            "name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "guid", "ipv4_gateway_address", "ipv6_gateway_address" },
            },

            "allow_disconnect": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
            },

            "connection_profile": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "ipv4_connectivity": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "ipv6_connectivity": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "network_adapter_names": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem:     &schema.Schema{ Type: schema.TypeString },
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for persistent resources (similar to data-sources)
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },

        Read:   dataSourceWindowsNetworkConnectionRead,
    }
}

//------------------------------------------------------------------------------

func dataSourceWindowsNetworkConnectionRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    guid               := d.Get("guid").(string)
    ipv4GatewayAddress := d.Get("ipv4_gateway_address").(string)
    ipv6GatewayAddress := d.Get("ipv6_gateway_address").(string)
    name               := d.Get("name").(string)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if guid               != "" { id = guid               } else
    if ipv4GatewayAddress != "" { id = ipv4GatewayAddress } else
    if ipv6GatewayAddress != "" { id = ipv6GatewayAddress } else
    if name               != "" { id = name               }
    id = fmt.Sprintf("//%s/network_connections/%s", host, id)

    log.Printf("[INFO][terraform-provider-windows] reading windows_network_connection %q\n", id)

    x_lifecycle := tfutil.GetResource(d, "x_lifecycle")

    // read
    ncQuery := new(api.NetworkConnection)
    ncQuery.GUID               = d.Get("guid").(string)
    ncQuery.IPv4GatewayAddress = d.Get("ipv4_gateway_address").(string)
    ncQuery.IPv6GatewayAddress = d.Get("ipv6_gateway_address").(string)
    ncQuery.Name               = d.Get("name").(string)
    ncQuery.AllowDisconnect    = d.Get("allow_disconnect").(bool)

    networkConnection, err := c.ReadNetworkConnection(ncQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        v, ok := x_lifecycle["ignore_error_if_not_exists"]
        if ok && v.(bool) && strings.Contains(err.Error(), "cannot find network_connection") {
            log.Printf("[INFO][terraform-provider-windows] cannot read windows_network_connection %q\n", id)

            // set zeroed properties
            d.Set("guid", "")
            d.Set("ipv4_gateway_address", "")
            d.Set("ipv6_gateway_address", "")
            d.Set("name", "")
            d.Set("allow_disconnect", false)
            d.Set("connection_profile", "")
            d.Set("ipv4_connectivity", "")
            d.Set("ipv6_connectivity", "")
            d.Set("network_adapter_names", nil)

            // set computed lifecycle properties
            x_lifecycle["exists"] = false
            d.Set("x_lifecycle", []interface{}{ x_lifecycle })

            // set id
            d.SetId(id)

            log.Printf("[INFO][terraform-provider-windows] ignored error and added zeroed windows_network_connection %q to terraform state\n", id)
            return nil
        }

        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot read windows_network_connection %q\n", id)
        return err
    }

    // set computed lifecycle properties
    x_lifecycle["exists"] = true
    d.Set("x_lifecycle", []interface{}{ x_lifecycle })

    // set properties
    setDataNetworkConnectionProperties(d, networkConnection)

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] read windows_network_connection %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setDataNetworkConnectionProperties(d *schema.ResourceData, ncProperties *api.NetworkConnection) {
    d.Set("guid", ncProperties.GUID)
    d.Set("ipv4_gateway_address", ncProperties.IPv4GatewayAddress)
    d.Set("ipv6_gateway_address", ncProperties.IPv6GatewayAddress)
    d.Set("name", ncProperties.Name)
    // d.Set("allow_disconnect", d.Get("allow_disconnect"))
    d.Set("connection_profile", ncProperties.ConnectionProfile)
    d.Set("ipv4_connectivity", ncProperties.IPv4Connectivity)
    d.Set("ipv6_connectivity", ncProperties.IPv6Connectivity)
    d.Set("network_adapter_names", ncProperties.NetworkAdapterNames)
}

//------------------------------------------------------------------------------
