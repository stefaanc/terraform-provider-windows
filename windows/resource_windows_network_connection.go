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

func resourceWindowsNetworkConnection() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "guid": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                ForceNew: true,

                ValidateFunc: tfutil.ValidateUUID(),
                StateFunc: tfutil.StateToUpper(),
            },

            "ipv4_gateway_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                ForceNew: true,

                ConflictsWith: []string{ "guid" },
                ValidateFunc: validation.SingleIP(),
            },
            "ipv6_gateway_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                ForceNew: true,

                ConflictsWith: []string{ "guid", "ipv4_gateway_address" },
                ValidateFunc: validation.SingleIP(),
                StateFunc: tfutil.StateToLower(),
            },

            "name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                // ForceNew: true,   // handled in CustomizeDiff method to avoid unwanted activation of ForceNew-condition because of 'new-name'

                ConflictsWith: []string{ "guid", "ipv4_gateway_address", "ipv6_gateway_address" },
            },
            "old_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                ForceNew: true,

                ConflictsWith: []string{ "guid", "ipv4_gateway_address", "ipv6_gateway_address", "name" },
            },
            "new_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
            },

            "allow_disconnect": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
            },

            "connection_profile": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: validation.StringInSlice([]string{ "public", "private" }, true),
                StateFunc:    tfutil.StateToCamel(),
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

            // used to reset values on terraform destroy
            "original": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: resourceWindowsNetworkConnectionOriginal(),
            },
        },

        CustomizeDiff: resourceWindowsNetworkConnectionCustomizeDiff,

        Create: resourceWindowsNetworkConnectionCreate,
        Read:   resourceWindowsNetworkConnectionRead,
        Update: resourceWindowsNetworkConnectionUpdate,
        Delete: resourceWindowsNetworkConnectionDelete,
    }
}

func resourceWindowsNetworkConnectionOriginal() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "old_name": &schema.Schema{
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

func resourceWindowsNetworkConnectionCustomizeDiff(d *schema.ResourceDiff, m interface{}) error {
    // don't set 'ForceNew: true' in the schema for 'name', but handle the ForceNew-condition explicitly in CustomizeDiff.
    // avoid that the ForceNew-condition is also activated when using SetNewComputed('name') for 'new_name'
    // activate only when 'name' is different in config vs state
    if d.HasChange("name") {
        d.ForceNew("name")
    }

    if d.HasChange("new_name") {
        if v, ok := d.GetOk("new_name"); ok && ( d.Get("name").(string) != v.(string) ) {
            d.SetNewComputed("name")
        }
    }

    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnectionCreate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    guid               := d.Get("guid").(string)
    ipv4GatewayAddress := d.Get("ipv4_gateway_address").(string)
    ipv6GatewayAddress := d.Get("ipv6_gateway_address").(string)
    name               := d.Get("name").(string)
    oldName            := d.Get("old_name").(string)
    newName            := d.Get("new_name").(string)
    connectionProfile  := d.Get("connection_profile").(string)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if guid               != "" { id = guid               } else
    if ipv4GatewayAddress != "" { id = ipv4GatewayAddress } else
    if ipv6GatewayAddress != "" { id = ipv6GatewayAddress } else
    if name               != "" { id = name               } else
    if oldName            != "" { id = oldName            }
    id = fmt.Sprintf("//%s/network_connections/%s", host, id)

    log.Printf(`[INFO][terraform-provider-windows] creating windows_network_connection %q
                    [INFO][terraform-provider-windows]     guid:                 %#v
                    [INFO][terraform-provider-windows]     ipv4_gateway_address: %#v
                    [INFO][terraform-provider-windows]     ipv6_gateway_address: %#v
                    [INFO][terraform-provider-windows]     name:                 %#v
                    [INFO][terraform-provider-windows]     old_name:             %#v
                    [INFO][terraform-provider-windows]     new_name:             %#v
                    [INFO][terraform-provider-windows]     connection_profile:   %#v
`       ,
        id,
        guid,
        ipv4GatewayAddress,
        ipv6GatewayAddress,
        name,
        oldName,
        newName,
        connectionProfile,
    )

    // import
    log.Printf("[INFO][terraform-provider-windows] importing windows_network_connection %q into terraform state\n", id)

    x_lifecycle := tfutil.GetResource(d, "x_lifecycle")

    ncQuery := new(api.NetworkConnection)
    ncQuery.GUID               = guid
    ncQuery.IPv4GatewayAddress = ipv4GatewayAddress
    ncQuery.IPv6GatewayAddress = ipv6GatewayAddress
    ncQuery.Name               = name
    ncQuery.OldName            = oldName
    ncQuery.AllowDisconnect    = d.Get("allow_disconnect").(bool)

    networkConnection, err := c.ReadNetworkConnection(ncQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        v, ok := x_lifecycle["ignore_error_if_not_exists"]
        if ok && v.(bool) && strings.Contains(err.Error(), "cannot find network_connection") {
            log.Printf("[INFO][terraform-provider-windows] cannot import windows_network_connection %q into terraform state\n", id)

            // set zeroed properties
            d.Set("guid", "")
            d.Set("ipv4_gateway_address", "")
            d.Set("ipv6_gateway_address", "")
            d.Set("name", "")
            d.Set("old_name", "")
            d.Set("new_name", "")
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
        log.Printf("[ERROR][terraform-provider-windows] cannot import windows_network_connection %q into terraform state\n", id)
        return err
    }

    // set computed lifecycle properties
    x_lifecycle["exists"] = true
    d.Set("x_lifecycle", []interface{}{ x_lifecycle })

    // save original config
    setOriginalNetworkConnectionProperties(d, networkConnection)

    if !diffNetworkConnectionProperties(d, networkConnection) {
        // no update required

        // set properties
        setNetworkConnectionProperties(d, networkConnection)

        // set id
        d.SetId(id)

        log.Printf("[INFO][terraform-provider-windows] created windows_network_connection %q\n", id)
        return nil

    } else {
        // update
        log.Printf("[INFO][terraform-provider-windows] updating windows_network_connection %q\n", id)

        // set principal identifying property, so it can be found after possible rename
        d.Set("guid", networkConnection.GUID)

        ncProperties := new(api.NetworkConnection)
        expandNetworkConnectionProperties(ncProperties, d)

        err := c.UpdateNetworkConnection(networkConnection, ncProperties)
        if err != nil {
            log.Printf("[ERROR][terraform-provider-windows] cannot update windows_network_connection %q\n", id)
            return err
        }

        // set id
        d.SetId(id)

        log.Printf("[INFO][terraform-provider-windows] created windows_network_connection %q\n", id)
        return resourceWindowsNetworkConnectionRead(d, m)
    }
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnectionRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id                 := d.Id()

    log.Printf("[INFO][terraform-provider-windows] reading windows_network_connection %q\n", id)

    x_lifecycle := tfutil.GetResource(d, "x_lifecycle")

    // read
    ncQuery := new(api.NetworkConnection)
    ncQuery.GUID               = d.Get("guid").(string)
    ncQuery.IPv4GatewayAddress = d.Get("ipv4_gateway_address").(string)
    ncQuery.IPv6GatewayAddress = d.Get("ipv6_gateway_address").(string)
    ncQuery.Name               = d.Get("name").(string)
    ncQuery.OldName            = d.Get("old_name").(string)
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
            d.Set("old_name", "")
            d.Set("new_name", "")
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

        // set id
        d.SetId("")

        log.Printf("[INFO][terraform-provider-windows] deleted windows_network_connection %q from terraform state\n", id)
        return nil   // don't return an error to allow terraform refresh to update state
    }

    // set properties
    setNetworkConnectionProperties(d, networkConnection)

    log.Printf("[INFO][terraform-provider-windows] read windows_network_connection %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkConnectionUpdate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id                 := d.Id()
    guid               := d.Get("guid").(string)
    ipv4GatewayAddress := d.Get("ipv4_gateway_address").(string)
    ipv6GatewayAddress := d.Get("ipv6_gateway_address").(string)
    name               := d.Get("name").(string)
    oldName            := d.Get("old_name").(string)
    newName            := d.Get("new_name").(string)
    connectionProfile  := d.Get("connection_profile").(string)

    log.Printf(`[INFO][terraform-provider-windows] updating windows_network_connection %q
                    [INFO][terraform-provider-windows]     guid:                 %#v
                    [INFO][terraform-provider-windows]     ipv4_gateway_address: %#v
                    [INFO][terraform-provider-windows]     ipv6_gateway_address: %#v
                    [INFO][terraform-provider-windows]     name:                 %#v
                    [INFO][terraform-provider-windows]     old_name:             %#v
                    [INFO][terraform-provider-windows]     new_name:             %#v
                    [INFO][terraform-provider-windows]     connection_profile:   %#v
`       ,
        id,
        guid,
        ipv4GatewayAddress,
        ipv6GatewayAddress,
        name,
        oldName,
        newName,
        connectionProfile,
    )

    // update
    ncQuery := new(api.NetworkConnection)
    ncQuery.GUID = guid

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

    id   := d.Id()

    log.Printf("[INFO][terraform-provider-windows] deleting windows_network_connection %q from terraform state\n", id)
    log.Printf("[INFO][terraform-provider-windows] restoring original properties for windows_network_connection %q\n", id)

    // restore original config
    ncQuery := new(api.NetworkConnection)
    ncQuery.GUID = d.Get("guid").(string)

    ncProperties := new(api.NetworkConnection)
    expandOriginalNetworkConnectionProperties(ncProperties, d)

    err := c.UpdateNetworkConnection(ncQuery, ncProperties)
    if err != nil {
        log.Printf("[WARNING][terraform-provider-windows] cannot restore original properties for windows_network_connection %q\n", id)
    }

    // set id
    d.SetId("")

    log.Printf("[INFO][terraform-provider-windows] deleted windows_network_connection %q from terraform state\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setNetworkConnectionProperties(d *schema.ResourceData, ncProperties *api.NetworkConnection) {
    d.Set("guid", ncProperties.GUID)

    d.Set("ipv4_gateway_address", ncProperties.IPv4GatewayAddress)
    d.Set("ipv6_gateway_address", ncProperties.IPv6GatewayAddress)

    d.Set("name", ncProperties.Name)
    // d.Set("old_name", d.Get("old_name"))
    // d.Set("new_name", d.Get("new_name"))

    // d.Set("allow_disconnect", d.Get("allow_disconnect"))

    d.Set("connection_profile", ncProperties.ConnectionProfile)

    d.Set("ipv4_connectivity", ncProperties.IPv4Connectivity)
    d.Set("ipv6_connectivity", ncProperties.IPv6Connectivity)

    d.Set("network_adapter_names", ncProperties.NetworkAdapterNames)
}

func setOriginalNetworkConnectionProperties(d *schema.ResourceData, ncProperties *api.NetworkConnection) {
    original := make(map[string]interface{})

    original["old_name"]           = ncProperties.Name
    original["connection_profile"] = ncProperties.ConnectionProfile

    d.Set("original", []interface{}{ original })
}

//------------------------------------------------------------------------------

func diffNetworkConnectionProperties(d *schema.ResourceData, ncProperties *api.NetworkConnection) bool {
    if v, ok := d.GetOk("new_name"); ok && ( ncProperties.Name != v.(string) ) {
        return true
    }

    if v, ok := d.GetOk("connection_profile"); ok && ( ncProperties.ConnectionProfile != v.(string) ) {
        return true
    }

    return false
}

func expandNetworkConnectionProperties(ncProperties *api.NetworkConnection, d *schema.ResourceData) {
    ncProperties.NewName           = d.Get("new_name").(string)
    ncProperties.ConnectionProfile = d.Get("connection_profile").(string)
}

func expandOriginalNetworkConnectionProperties(ncProperties *api.NetworkConnection, d *schema.ResourceData) {
    original := tfutil.GetResource(d, "original")

    ncProperties.NewName           = original["old_name"].(string)
    ncProperties.ConnectionProfile = original["connection_profile"].(string)
}

//------------------------------------------------------------------------------
