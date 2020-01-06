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

func resourceWindowsNetworkAdapter() *schema.Resource {
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

            "name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                // ForceNew: true,   // handled in CustomizeDiff method to avoid unwanted activation of ForceNew-condition because of 'new-name'

                ConflictsWith: []string{ "guid" },
            },
            "old_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                ForceNew: true,

                ConflictsWith: []string{ "guid", "name" },
            },
            "new_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: tfutil.ValidateSingleMAC(),
                StateFunc: tfutil.StateAll(
                    tfutil.StateToUpper(),
                    tfutil.StateAcceptEmptyString(),   // workaround for ?terraform bug?, replaces "" with "<empty>"
                ),
            },
            "permanent_mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "dns_client": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Optional: true,
                Computed: true,
                Elem: resourceWindowsNetworkAdapterDNSClient(),
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

            // used to reset values on terraform destroy
            "original": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: resourceWindowsNetworkAdapterOriginal(),
            },
        },

        CustomizeDiff: resourceWindowsNetworkAdapterCustomizeDiff,

        Create: resourceWindowsNetworkAdapterCreate,
        Read:   resourceWindowsNetworkAdapterRead,
        Update: resourceWindowsNetworkAdapterUpdate,
        Delete: resourceWindowsNetworkAdapterDelete,
    }
}

func resourceWindowsNetworkAdapterDNSClient() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "register_connection_address": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Computed: true,
            },
            "register_connection_suffix": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                StateFunc: tfutil.StateAll(
                    tfutil.StateToLower(),
                    tfutil.StateAcceptEmptyString(),   // workaround for ?terraform bug?, replaces "" with "<empty>"
                ),
            },
        },
    }
}

func resourceWindowsNetworkAdapterOriginal() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "old_name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "dns_client": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: &schema.Resource{
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
                },
            },
        },
    }
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterCustomizeDiff(d *schema.ResourceDiff, m interface{}) error {
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

func resourceWindowsNetworkAdapterCreate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    guid            := d.Get("guid").(string)
    name            := d.Get("name").(string)
    oldName         := d.Get("old_name").(string)
    newName         := d.Get("new_name").(string)
    macAddress      := d.Get("mac_address").(string)
    dnsClient       := tfutil.GetResource(d, "dns_client")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if guid    != "" { id = guid    } else
    if name    != "" { id = name    } else
    if oldName != "" { id = oldName }
    id = fmt.Sprintf("//%s/network_adapters/%s", host, id)

    log.Printf(`[INFO][terraform-provider-windows] creating windows_network_adapter %q
                    [INFO][terraform-provider-windows]     guid:        %#v
                    [INFO][terraform-provider-windows]     name:        %#v
                    [INFO][terraform-provider-windows]     old_name:    %#v
                    [INFO][terraform-provider-windows]     new_name:    %#v
                    [INFO][terraform-provider-windows]     mac_address: %#v
                    [INFO][terraform-provider-windows]     dns_client {
                    [INFO][terraform-provider-windows]         register_connection_address: %#v
                    [INFO][terraform-provider-windows]         register_connection_suffix:  %#v
                    [INFO][terraform-provider-windows]     }
`       ,
        id,
        guid,
        name,
        oldName,
        newName,
        macAddress,
        dnsClient["register_connection_address"],
        dnsClient["register_connection_suffix"],
    )

    // import
    log.Printf("[INFO][terraform-provider-windows] importing windows_network_adapter %q into terraform state\n", id)

    x_lifecycle := tfutil.GetResource(d, "x_lifecycle")

    naQuery := new(api.NetworkAdapter)
    naQuery.GUID    = guid
    naQuery.Name    = name
    naQuery.OldName = oldName

    networkAdapter, err := c.ReadNetworkAdapter(naQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        v, ok := x_lifecycle["ignore_error_if_not_exists"]
        if ok && v.(bool) && strings.Contains(err.Error(), "cannot find network_adapter") {
            log.Printf("[INFO][terraform-provider-windows] cannot import windows_network_adapter %q into terraform state\n", id)

            // set zeroed properties
            d.Set("guid", "")
            d.Set("name", "")
            d.Set("old_name", "")
            d.Set("new_name", "")
            d.Set("mac_address", "")
            d.Set("permanent_mac_address", "")
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
        log.Printf("[ERROR][terraform-provider-windows] cannot import windows_network_adapter %q into terraform state\n", id)
        return err
    }

    // set computed lifecycle properties
    x_lifecycle["exists"] = true
    d.Set("x_lifecycle", []interface{}{ x_lifecycle })

    // save original config
    setOriginalNetworkAdapterProperties(d, networkAdapter)

    // check diff
    if !diffNetworkAdapterProperties(d, networkAdapter) {
        // no update required

        // set properties
        setNetworkAdapterProperties(d, networkAdapter)

        // set id
        d.SetId(id)

        log.Printf("[INFO][terraform-provider-windows] created windows_network_adapter %q\n", id)
        return nil

    } else {
        // update
        log.Printf("[INFO][terraform-provider-windows] updating windows_network_adapter %q\n", id)

        // set principal identifying property, so it can be found after possible rename
        d.Set("guid", networkAdapter.GUID)

        naProperties := new(api.NetworkAdapter)
        expandNetworkAdapterProperties(naProperties, d)

        err := c.UpdateNetworkAdapter(networkAdapter, naProperties)
        if err != nil {
            log.Printf("[ERROR][terraform-provider-windows] cannot update windows_network_adapter %q\n", id)
            return err
        }

        // set id
        d.SetId(id)

        log.Printf("[INFO][terraform-provider-windows] created windows_network_adapter %q\n", id)
        return resourceWindowsNetworkAdapterRead(d, m)
    }
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id       := d.Id()

    log.Printf("[INFO][terraform-provider-windows] reading windows_network_adapter %q\n", id)

    x_lifecycle := tfutil.GetResource(d, "x_lifecycle")

    // read
    naQuery := new(api.NetworkAdapter)
    naQuery.GUID    = d.Get("guid").(string)
    naQuery.Name    = d.Get("name").(string)
    naQuery.OldName = d.Get("old_name").(string)

    networkAdapter, err := c.ReadNetworkAdapter(naQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        v, ok := x_lifecycle["ignore_error_if_not_exists"]
        if ok && v.(bool) && strings.Contains(err.Error(), "cannot find network_adapter") {
            log.Printf("[INFO][terraform-provider-windows] cannot import windows_network_adapter %q into terraform state\n", id)

            // set zeroed properties
            d.Set("guid", "")
            d.Set("name", "")
            d.Set("old_name", "")
            d.Set("new_name", "")
            d.Set("mac_address", "")
            d.Set("permanent_mac_address", "")
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

        // set id
        d.SetId("")

        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] deleted windows_network_adapter %q from terraform state\n", id)
        return nil   // don't return an error to allow terraform refresh to update state
    }

    // set properties
    setNetworkAdapterProperties(d, networkAdapter)

    log.Printf("[INFO][terraform-provider-windows] read windows_network_adapter %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterUpdate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id         := d.Id()
    guid       := d.Get("guid").(string)
    name       := d.Get("name").(string)
    oldName    := d.Get("old_name").(string)
    newName    := d.Get("new_name").(string)
    macAddress := d.Get("mac_address").(string)
    dnsClient  := tfutil.GetResource(d, "dns_client")

    log.Printf(`[INFO][terraform-provider-windows] updating windows_network_adapter %q
                    [INFO][terraform-provider-windows]     guid:        %#v
                    [INFO][terraform-provider-windows]     name:        %#v
                    [INFO][terraform-provider-windows]     old_name:    %#v
                    [INFO][terraform-provider-windows]     new_name:    %#v
                    [INFO][terraform-provider-windows]     mac_address: %#v
                    [INFO][terraform-provider-windows]     dns_client {
                    [INFO][terraform-provider-windows]         register_connection_address: %#v
                    [INFO][terraform-provider-windows]         register_connection_suffix:  %#v
                    [INFO][terraform-provider-windows]     }
`       ,
        id,
        guid,
        name,
        oldName,
        newName,
        macAddress,
        dnsClient["register_connection_address"],
        dnsClient["register_connection_suffix"],
    )

    // update
    naQuery := new(api.NetworkAdapter)
    naQuery.GUID = guid

    naProperties := new(api.NetworkAdapter)
    expandNetworkAdapterProperties(naProperties, d)

    err := c.UpdateNetworkAdapter(naQuery, naProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows] cannot update windows_network_adapter %q\n", id)
        return err
    }

    log.Printf("[INFO][terraform-provider-windows] updated windows_network_adapter %q\n", id)
    return resourceWindowsNetworkAdapterRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterDelete(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id       := d.Id()

    log.Printf("[INFO][terraform-provider-windows] deleting windows_network_adapter %q from terraform state\n", id)
    log.Printf("[INFO][terraform-provider-windows] restore original config for windows_network_adapter %q\n", id)

    // restore original config
    naQuery := new(api.NetworkAdapter)
    naQuery.GUID = d.Get("guid").(string)

    naProperties := new(api.NetworkAdapter)
    expandOriginalNetworkAdapterProperties(naProperties, d)

    err := c.UpdateNetworkAdapter(naQuery, naProperties)
    if err != nil {
        log.Printf("[WARNING][terraform-provider-windows] cannot restore original config for windows_network_adapter %q\n", id)
    }

    // set id
    d.SetId("")

    log.Printf("[INFO][terraform-provider-windows] deleted windows_network_adapter %q from terraform state\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setNetworkAdapterProperties(d *schema.ResourceData, naProperties *api.NetworkAdapter) {
    d.Set("guid", naProperties.GUID)

    d.Set("name", naProperties.Name)
    // d.Set("old_name", d.Get("old_name"))
    // d.Set("new_name", d.Get("new_name"))

    d.Set("mac_address", naProperties.MACAddress)
    d.Set("permanent_mac_address", naProperties.PermanentMACAddress)

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

func setOriginalNetworkAdapterProperties(d *schema.ResourceData, naProperties *api.NetworkAdapter) {
    original := make(map[string]interface{})

    original["old_name"]    = naProperties.Name
    original["mac_address"] = naProperties.MACAddress

    if len(naProperties.DNSClient) > 0 {
        original_dnsClient := make(map[string]interface{})
        original_dnsClient["register_connection_address"] = naProperties.DNSClient[0].RegisterConnectionAddress
        original_dnsClient["register_connection_suffix"]  = naProperties.DNSClient[0].RegisterConnectionSuffix
        original["dns_client"] = []interface{}{ original_dnsClient }
    } else {
        original["dns_client"] = []interface{}{ }
    }

    d.Set("original", []interface{}{ original })
}

//------------------------------------------------------------------------------

func diffNetworkAdapterProperties(d *schema.ResourceData, naProperties *api.NetworkAdapter) bool {
    if v, ok := d.GetOk("new_name"); ok && ( naProperties.Name != v.(string) ) {
        return true
    }

    if v, ok := d.GetOk("mac_address"); ok && ( naProperties.MACAddress != v.(string) ) {
        return true
    }

    if v, ok := d.GetOk("dns_client"); ok && ( len(v.([]interface{})) > 0 ) {
        if len(naProperties.DNSClient) > 0 {
            if v, ok := d.GetOkExists("dns_client.0.register_connection_address"); ok && ( naProperties.DNSClient[0].RegisterConnectionAddress != v.(bool) ) {
                return true
            }
            if v, ok := d.GetOkExists("dns_client.0.register_connection_suffix"); ok &&
               ( ( ( v.(string) != "<empty>" ) && ( naProperties.DNSClient[0].RegisterConnectionSuffix != v.(string) ) )   ||
                 ( ( v.(string) == "<empty>" ) && ( naProperties.DNSClient[0].RegisterConnectionSuffix != ""         ) ) ) {
                return true
            }
        } else {
            return true
        }
    }

    return false
}

func expandNetworkAdapterProperties(naProperties *api.NetworkAdapter, d *schema.ResourceData) {
    naProperties.NewName    = d.Get("new_name").(string)
    naProperties.MACAddress = d.Get("mac_address").(string)

    dnsClientList := tfutil.GetListOfResources(d, "dns_client")
    if len(dnsClientList) > 0 {
        naProperties.DNSClient = make([]api.NetworkAdapterDNSClient, 1, 1)

        if v, ok := d.GetOkExists("dns_client.0.register_connection_address"); ok {
            naProperties.DNSClient[0].RegisterConnectionAddress = v.(bool)
        } else {
            // there is no state yet (the resource is being created), and the attribute is not defined in config
            // since we don't have a state yet, we use the previously read original properties to get the current value (using the zero-value 'false', would overwrite the current value)
            original_dnsClientList := tfutil.GetListOfResources(d, "original.0.dns_client")
            if len(original_dnsClientList) > 0 {
                naProperties.DNSClient[0].RegisterConnectionAddress = original_dnsClientList[0]["register_connection_address"].(bool)
            }
        }

        if v, ok := d.GetOkExists("dns_client.0.register_connection_suffix"); ok {
            if v.(string) == "<empty>" {
                // "" has been replaced with "<empty>" (because "" is a valid value and would get skipped in diff since it is equal to the zero value)
                naProperties.DNSClient[0].RegisterConnectionSuffix = ""
            } else {
                naProperties.DNSClient[0].RegisterConnectionSuffix = v.(string)
            }
        } else {
            // there is no state yet (the resource is being created), and the attribute is not defined in config
            // since we don't have a state yet, we use the previously read original properties to get the current value (using the zero-value "", would overwrite the current value)
            original_dnsClientList := tfutil.GetListOfResources(d, "original.0.dns_client")
            if len(original_dnsClientList) > 0 {
                naProperties.DNSClient[0].RegisterConnectionSuffix = original_dnsClientList[0]["register_connection_suffix"].(string)
            }
        }
    }
}

func expandOriginalNetworkAdapterProperties(naProperties *api.NetworkAdapter, d *schema.ResourceData) {
    original := tfutil.GetResource(d, "original")

    naProperties.NewName    = original["old_name"].(string)
    naProperties.MACAddress = original["mac_address"].(string)

    original_dnsClientList := tfutil.ExpandListOfResources(original, "dns_client")
    if len(original_dnsClientList) > 0 {
        original_dnsClient := original_dnsClientList[0]

        naProperties.DNSClient = make([]api.NetworkAdapterDNSClient, 1, 1)
        naProperties.DNSClient[0].RegisterConnectionAddress = original_dnsClient["register_connection_address"].(bool)
        naProperties.DNSClient[0].RegisterConnectionSuffix  = original_dnsClient["register_connection_suffix"].(string)
    }
}

//------------------------------------------------------------------------------
