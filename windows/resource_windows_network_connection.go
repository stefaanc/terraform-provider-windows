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

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"

    "github.com/stefaanc/terraform-provider-windows/api"
    "github.com/stefaanc/terraform-provider-windows/windows/tfutil"
)

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnection() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "ipv4_gateway_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                ForceNew: true,

                ConflictsWith: []string{ "name" },
                ValidateFunc: validation.SingleIP(),
            },
            "ipv6_gateway_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                ForceNew: true,

                ConflictsWith: []string{ "ipv4_gateway_address" },
                ValidateFunc: validation.SingleIP(),
                StateFunc: tfutil.StateToUpper(),
            },
            "allow_disconnect": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
            },

            "name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "ipv4_gateway_address", "ipv6_gateway_address" },
                StateFunc: tfutil.StateToUpper(),
            },

            "connection_profile": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: validation.StringInSlice([]string{ "public", "private" }, true),
                StateFunc:    tfutil.StateToCamel(),
            },

            "original": &schema.Schema{   // used to reset values on terraform destroy
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: resourceWindowsNetworkConnectionOriginal(),
            },
        },

        Create: resourceWindowsNetworkConnectionCreate,
        Read:   resourceWindowsNetworkConnectionRead,
        Update: resourceWindowsNetworkConnectionUpdate,
        Delete: resourceWindowsNetworkConnectionDelete,
    }
}

func resourceWindowsNetworkConnectionOriginal() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "connection_profile": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
        },
    }
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnectionCreate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    name               := d.Get("name").(string)
    ipv4GatewayAddress := d.Get("ipv4_gateway_address").(string)
    ipv6GatewayAddress := d.Get("ipv6_gateway_address").(string)
    connectionProfile  := d.Get("connection_profile").(string)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if        name != ""               { id = name
    } else if ipv4GatewayAddress != "" { id = ipv4GatewayAddress
    } else if ipv6GatewayAddress != "" { id = ipv6GatewayAddress
    }
    id = fmt.Sprintf("//%s/network_connections/%s", host, id)

    log.Printf(`[INFO][terraform-provider-windows] creating windows_network_connection %q
                    [INFO][terraform-provider-windows]     name:                 %#v
                    [INFO][terraform-provider-windows]     ipv4_gateway_address: %#v
                    [INFO][terraform-provider-windows]     ipv6_gateway_address: %#v
                    [INFO][terraform-provider-windows]     connection_profile:   %#v
`       ,
        id,
        name,
        ipv4GatewayAddress,
        ipv6GatewayAddress,
        connectionProfile,
    )

    // import network
    log.Printf("[INFO][terraform-provider-windows] importing windows_network_connection %q into terraform state\n", id)

    ncQuery := new(api.NetworkConnection)
    ncQuery.Name               = name
    ncQuery.IPv4GatewayAddress = ipv4GatewayAddress
    ncQuery.IPv6GatewayAddress = ipv6GatewayAddress

    networkConnection, err := c.ReadNetworkConnection(ncQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows] cannot import windows_network_connection %q into terraform state\n", id)
        return err
    }

    // save existing config
    setOriginalNetworkConnectionProperties(d, networkConnection)

    // update network
    if networkConnection.Name != name                           ||
       networkConnection.ConnectionProfile != connectionProfile {

        log.Printf("[INFO][terraform-provider-windows] updating windows_network_connection %q\n", id)

        ncProperties := new(api.NetworkConnection)
        expandNetworkConnectionProperties(ncProperties, d)

        err := c.UpdateNetworkConnection(networkConnection, ncProperties)
        if err != nil {
            log.Printf("[ERROR][terraform-provider-windows] cannot update windows_network_connection %q\n", id)
            log.Printf("[ERROR][terraform-provider-windows] cannot import windows_network_connection %q into terraform state\n", id)
            return err
        }
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] created windows_network_connection %q\n", id)
    return resourceWindowsNetworkConnectionRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnectionRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id                 := d.Id()
    name               := d.Get("name").(string)
    ipv4GatewayAddress := d.Get("ipv4_gateway_address").(string)
    ipv6GatewayAddress := d.Get("ipv6_gateway_address").(string)
    original           := tfutil.GetResource(d, "original")   // make sure new terraform state includes 'original' from the old terraform state when doing a terraform refresh

    log.Printf("[INFO][terraform-provider-windows] reading windows_network_connection %q\n", id)

    // read network connection
    ncQuery := new(api.NetworkConnection)
    ncQuery.Name               = name
    ncQuery.IPv4GatewayAddress = ipv4GatewayAddress
    ncQuery.IPv6GatewayAddress = ipv6GatewayAddress

    networkConnection, err := c.ReadNetworkConnection(ncQuery)
    if err != nil {
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot read windows_network_connection %q\n", id)

        // set id
        d.SetId("")

        log.Printf("[INFO][terraform-provider-windows] deleted windows_network_connection %q from terraform state\n", id)
        return nil   // don't return an error to allow terraform refresh to update state
    }

    // set properties
    setNetworkConnectionProperties(d, networkConnection)
    d.Set("original", []interface{}{ original })   // make sure new terraform state includes 'original' from the old terraform state when doing a terraform refresh

    log.Printf("[INFO][terraform-provider-windows] read windows_network_connection %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnectionUpdate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id                 := d.Id()
    name               := d.Get("name").(string)
    ipv4GatewayAddress := d.Get("ipv4_gateway_address").(string)
    ipv6GatewayAddress := d.Get("ipv6_gateway_address").(string)
    connectionProfile  := d.Get("connection_profile").(string)

    log.Printf(`[INFO][terraform-provider-windows] updating windows_network_connection %q
                    [INFO][terraform-provider-windows]     name:                 %#v
                    [INFO][terraform-provider-windows]     ipv4_gateway_address: %#v
                    [INFO][terraform-provider-windows]     ipv6_gateway_address: %#v
                    [INFO][terraform-provider-windows]     connection_profile:   %#v
`       ,
        id,
        name,
        ipv4GatewayAddress,
        ipv6GatewayAddress,
        connectionProfile,
    )

    // update network
    ncQuery := new(api.NetworkConnection)
    ncQuery.IPv4GatewayAddress = ipv4GatewayAddress
    ncQuery.IPv6GatewayAddress = ipv6GatewayAddress

    ncProperties := new(api.NetworkConnection)
    expandNetworkConnectionProperties(ncProperties, d)

    err := c.UpdateNetworkConnection(ncQuery, ncProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows] cannot update windows_network_connection %q\n", id)
        return err
    }

    log.Printf("[INFO][terraform-provider-windows] updated windows_network_connection %q\n", id)
    return resourceWindowsNetworkConnectionRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnectionDelete(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id             := d.Id()
    ipv4GatewayAddress := d.Get("ipv4_gateway_address").(string)
    ipv6GatewayAddress := d.Get("ipv6_gateway_address").(string)

    log.Printf("[INFO][terraform-provider-windows] deleting windows_network_connection %q\n", id)
    log.Printf("[INFO][terraform-provider-windows] restore original properties for windows_network_connection %q\n", id)

    // restore network
    ncQuery := new(api.NetworkConnection)
    ncQuery.IPv4GatewayAddress = ipv4GatewayAddress
    ncQuery.IPv6GatewayAddress = ipv6GatewayAddress

    ncProperties := new(api.NetworkConnection)
    expandOriginalNetworkConnectionProperties(ncProperties, d)

    err := c.UpdateNetworkConnection(ncQuery, ncProperties)
    if err != nil {
        log.Printf("[WARNING][terraform-provider-windows] cannot restore original properties for windows_network_connection %q\n", id)
    }

    // set id
    d.SetId("")

    log.Printf("[INFO][terraform-provider-windows] deleted windows_network_connection %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setNetworkConnectionProperties(d *schema.ResourceData, nProperties *api.NetworkConnection) {
    d.Set("name", nProperties.Name)
    d.Set("ipv4_gateway_address", nProperties.IPv4GatewayAddress)
    d.Set("ipv6_gateway_address", nProperties.IPv6GatewayAddress)
    d.Set("connection_profile", nProperties.ConnectionProfile)
}

func setOriginalNetworkConnectionProperties(d *schema.ResourceData, nProperties *api.NetworkConnection) {
    original := make(map[string]interface{})

    original["name"]                 = nProperties.Name
    original["ipv4_gateway_address"] = nProperties.IPv4GatewayAddress
    original["ipv6_gateway_address"] = nProperties.IPv6GatewayAddress
    original["connection_profile"]   = nProperties.ConnectionProfile

    d.Set("original", []interface{}{ original })
}

//------------------------------------------------------------------------------

func expandNetworkConnectionProperties(nProperties *api.NetworkConnection, d *schema.ResourceData) {
    if v, ok := d.GetOkExists("name"); ok {
        nProperties.Name = v.(string)
    }
    if v, ok := d.GetOkExists("connection_profile"); ok {
        nProperties.ConnectionProfile = v.(string)
    }
}

func expandOriginalNetworkConnectionProperties(nProperties *api.NetworkConnection, d *schema.ResourceData) {
    original := tfutil.GetResource(d, "original")

    nProperties.Name              = original["name"].(string)
    nProperties.ConnectionProfile = original["connection_profile"].(string)
}

//------------------------------------------------------------------------------
