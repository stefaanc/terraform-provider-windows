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

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"

    "github.com/stefaanc/terraform-provider-windows/api"
    "github.com/stefaanc/terraform-provider-windows/windows/tfutil"
)

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapter() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,

                StateFunc: tfutil.StateToUpper(),
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: tfutil.ValidateSingleMAC(),
                StateFunc: tfutil.StateToUpper(),
            },

            "ipv4": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Optional: true,
                Computed: true,
                Elem: resourceWindowsNetworkAdapterIPInterface(),
            },

            "ipv6": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Optional: true,
                Computed: true,
                Elem: resourceWindowsNetworkAdapterIPInterface(),
            },

            "dns": &schema.Schema{
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

func resourceWindowsNetworkAdapterIPInterface() *schema.Resource {
    return &schema.Resource{
        Schema:  map[string]*schema.Schema{
            "interface_metric": &schema.Schema{
                Type:     schema.TypeInt,   // uint32
                Optional: true,
                Computed: true,

                ValidateFunc: validation.IntBetween(0, 4294967295),
                DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
                    if d.Get("interface_metric_obtained_automatically").(bool) && ( new == "0" ) {
                        // suppress diff when interface_metric is already obtained automatically
                        return true
                    }
                    return false
                },
            },
            "interface_metric_obtained_automatically": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },

            "dhcp_enabled": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Computed: true,
            },

            "ip_address": &schema.Schema{
                Type:     schema.TypeSet,
                Optional: true,
                Computed: true,
                Elem: resourceWindowsNetworkAdapterIPAddress(),

                DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
                    if d.Get("ip_address_obtained_through_dhcp").(bool) && ( new == "[]" ) {
                        // suppress diff when ip_address is already obtained through dhcp
                        return true
                    }
                    return false
                },
            },
            "ip_address_obtained_through_automatically": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },

            "gateway": &schema.Schema{
                Type:     schema.TypeSet,
                Optional: true,
                Computed: true,
                Elem: resourceWindowsNetworkAdapterGateway(),

                DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
                    if d.Get("gateway_obtained_through_dhcp").(bool) && ( new == "[]" ) {
                        // suppress diff when gateway is already obtained through dhcp
                        return true
                    }
                    return false
                },
            },
            "gateway_obtained_automatically": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },

            "dns_addresses": &schema.Schema{
                Type:     schema.TypeList,
                Optional: true,
                Computed: true,
                Elem:     &schema.Schema{ Type: schema.TypeString, ValidateFunc: validation.SingleIP() },

                DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
                    if d.Get("dns_addresses_obtained_through_dhcp").(bool) && ( new == "[]" ) {
                        // suppress diff when dns_addresses is already obtained through dhcp
                        return true
                    }
                    return false
                },
            },
            "dns_addresses_obtained_automatically": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },

            "connection_status": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "connectivity": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
        },

        CustomizeDiff: resourceWindowsNetworkAdapterIPInterfaceCustomizeDiff,
    }
}

func resourceWindowsNetworkAdapterIPAddress() *schema.Resource {
    return &schema.Resource{
        Schema:  map[string]*schema.Schema{
            "address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: validation.SingleIP(),
            },
            "prefix_length": &schema.Schema{
                Type:     schema.TypeInt,   // uint8
                Optional: true,
                Computed: true,

                ValidateFunc: validation.IntBetween(0, 255),
             },
            "skip_as_source": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Computed: true,
            },
        },
    }
}

func resourceWindowsNetworkAdapterGateway() *schema.Resource {
    return &schema.Resource{
        Schema:  map[string]*schema.Schema{
            "address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ValidateFunc: validation.SingleIP(),
            },
            "route_metric": &schema.Schema{
                Type:     schema.TypeInt,   // uint16
                Optional: true,
                Computed: true,

                ValidateFunc: validation.IntBetween(0, 65535),
                DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
                    if d.Get("route_metric_obtained_automatically").(bool) && ( new == "0" ) {
                        // suppress diff when route_metric is already obtained automatically
                        return true
                    }
                    return false
                },
            },
            "route_metric_obtained_automatically": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
        },

        CustomizeDiff: resourceWindowsNetworkAdapterGatewayCustomizeDiff,
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
            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "ipv4": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "interface_metric": &schema.Schema{
                            Type:     schema.TypeInt,   // uint32
                            Computed: true,
                        },
                        "interface_metric_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "dhcp_enabled": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "ip_address": &schema.Schema{
                            Type:     schema.TypeSet,
                            Computed: true,
                            Elem: &schema.Resource{
                                Schema: map[string]*schema.Schema{
                                    "address": &schema.Schema{
                                        Type:     schema.TypeString,
                                        Computed: true,
                                    },
                                    "prefix_length": &schema.Schema{
                                        Type:     schema.TypeInt,   // uint8
                                        Computed: true,
                                     },
                                    "skip_as_source": &schema.Schema{
                                        Type:     schema.TypeBool,
                                        Computed: true,
                                    },
                                },
                            },
                        },
                        "ip_address_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "gateway": &schema.Schema{
                            Type:     schema.TypeSet,
                            Computed: true,
                            Elem: &schema.Resource{
                                Schema: map[string]*schema.Schema{
                                    "address": &schema.Schema{
                                        Type:     schema.TypeString,
                                        Computed: true,
                                    },
                                    "route_metric": &schema.Schema{
                                        Type:     schema.TypeInt,   // uint16
                                        Computed: true,
                                    },
                                    "route_metric_obtained_automatically": &schema.Schema{
                                        Type:     schema.TypeBool,
                                        Computed: true,
                                    },
                                },
                            },
                        },
                        "gateway_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "dns_addresses": &schema.Schema{
                            Type:     schema.TypeList,
                            Computed: true,
                            Elem:     &schema.Schema{ Type: schema.TypeString },
                        },
                        "dns_addresses_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                    },
                },
            },

            "ipv6": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "interface_metric": &schema.Schema{
                            Type:     schema.TypeInt,   // uint32
                            Computed: true,
                        },
                        "interface_metric_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "dhcp_enabled": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "ip_address": &schema.Schema{
                            Type:     schema.TypeSet,
                            Computed: true,
                            Elem: &schema.Resource{
                                Schema: map[string]*schema.Schema{
                                    "address": &schema.Schema{
                                        Type:     schema.TypeString,
                                        Computed: true,
                                    },
                                    "prefix_length": &schema.Schema{
                                        Type:     schema.TypeInt,   // uint8
                                        Computed: true,
                                     },
                                    "skip_as_source": &schema.Schema{
                                        Type:     schema.TypeBool,
                                        Computed: true,
                                    },
                                },
                            },
                        },
                        "ip_address_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "gateway": &schema.Schema{
                            Type:     schema.TypeSet,
                            Computed: true,
                            Elem: &schema.Resource{
                                Schema: map[string]*schema.Schema{
                                    "address": &schema.Schema{
                                        Type:     schema.TypeString,
                                        Computed: true,
                                    },
                                    "route_metric": &schema.Schema{
                                        Type:     schema.TypeInt,   // uint16
                                        Computed: true,

                                        // cannot DiffSuppress route_metric attribute when route_metric_obtained_automatically is true and route_metric is 0,
                                        // because no way to access route_metric_obtained_automatically in a TypeSet schema (list-index is unknown),
                                        // so need to do this in CustomizeDiff
                                    },
                                    "route_metric_obtained_automatically": &schema.Schema{
                                        Type:     schema.TypeBool,
                                        Computed: true,
                                    },
                                },
                            },
                        },
                        "gateway_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },

                        "dns_addresses": &schema.Schema{
                            Type:     schema.TypeList,
                            Computed: true,
                            Elem:     &schema.Schema{ Type: schema.TypeString },
                        },
                        "dns_addresses_obtained_automatically": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                    },
                },
            },

            "dns": &schema.Schema{
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
    if d.HasChange("mac_address") {
        // this will restart the network adapter
        // hence some obtained and computed attributes will be re-obtained/computed
        if _, ok := d.GetOk("ipv4"); ok &&
           d.Get("ipv4.0.interface_metric_obtained_automatically").(bool) ||
           d.Get("ipv4.0.ip_address_obtained_automatically").(bool) {

            d.SetNewComputed("ipv4")
        }

        if _, ok := d.GetOk("ipv6"); ok &&
           d.Get("ipv6.0.interface_metric_obtained_automatically").(bool) ||
           d.Get("ipv6.0.ip_address_obtained_automatically").(bool) {

            d.SetNewComputed("ipv6")
        }

        d.SetNewComputed("connection_status")
        d.SetNewComputed("connection_speed")
    }

    return nil
}

func resourceWindowsNetworkAdapterIPInterfaceCustomizeDiff(d *schema.ResourceDiff, m interface{}) error {
    if d.HasChange("interface_metric") {
        n := d.Get("interface_metric")
        if n.(int) == 0 {
            // new value will be obtained automatically
            d.SetNewComputed("interface_metric")
            // old value was not obtained automatically or didn't exist, new value will
            // (case where old value was obtained automatically is handled by DiffSuppress)
            d.SetNew("interface_metric_obtained_automatically", true)
        } else {
            // new value is set in config
            if d.Get("interface_metric_obtained_automatically").(bool) {
                // old value was obtained automatically, new value won't
                d.SetNew("interface_metric_obtained_automatically", false)
            }
        }
    }

    if d.HasChange("ip_address") {
        n := d.Get("ip_address")
        if n.(*schema.Set).Len() == 0 {
            // new value will be obtained automatically
            d.SetNewComputed("ip_address")
            // old value was not obtained automatically, new value will
            // (case where old value was automatically dhcp is handled by DiffSuppress)
            d.SetNew("ip_address_obtained_automatically", true)
        } else {
            // new value is set in config
            if d.Get("ip_address_obtained_through_dhcp").(bool) {
                // old value was obtained automatically, new value won't
                d.SetNew("ip_address_obtained_automatically", false)
            }
        }
    }

    if d.HasChange("gateway") {
        n := d.Get("gateway")
        if n.(*schema.Set).Len() == 0 {
            // new value will be obtained automatically
            d.SetNewComputed("gateway")
            // old value was not obtained automatically, new value will
            // (case where old value was obtained automatically is handled by DiffSuppress)
            d.SetNew("gateway_obtained_automatically", true)
        } else {
            // new value is set in config
            if d.Get("gateway_obtained_automatically").(bool) {
                // old value was obtained automatically, new value won't
                d.SetNew("gateway_obtained_automatically", false)
            }
        }
    }

    if d.HasChange("dns_addresses") {
        n := d.Get("dns_addresses")
        if len(n.([]interface{})) == 0 {
            // new value will be obtained through dhcp
            d.SetNewComputed("dns_addresses")
            // old value was not obtained automatically, new value will
            // (case where old value was obtained automatically is handled by DiffSuppress)
            d.SetNew("dns_addresses_obtained_automatically", true)
        } else {
            // new value is set in config
            if d.Get("dns_addresses_obtained_automatically").(bool) {
                // old value was obtained automatically, new value won't
                d.SetNew("dns_addresses_obtained_automatically", false)
            }
        }
    }

    return nil
}

func resourceWindowsNetworkAdapterGatewayCustomizeDiff(d *schema.ResourceDiff, m interface{}) error {
    if d.HasChange("route_metric") {
        n := d.Get("route_metric")
        if n.(int) == 0 {
            // new value will be obtained automatically
            d.SetNewComputed("route_metric")
            // old value was not obtained automatically, new value will
            // (case where old value was obtained automatically is handled by DiffSuppress)
            d.SetNew("route_metric_obtained_automatically", true)
        } else {
            // new value is set in config
            if d.Get("route_metric_obtained_automatically").(bool) {
                // old value was obtained automatically, new value won't
                d.SetNew("route_metric_obtained_automatically", false)
            }
        }
    }

    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterCreate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    name       := d.Get("name").(string)
    macAddress := d.Get("mac_address").(string)
    ipv4       := tfutil.GetResource(d, "ipv4")
    ipv6       := tfutil.GetResource(d, "ipv6")
    dns        := tfutil.GetResource(d, "dns")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id := fmt.Sprintf("//%s/network_adapters/%s", host, d.Get("name").(string))

    log.Printf(`[INFO][terraform-provider-windows] creating windows_network_adapter %q
                    [INFO][terraform-provider-windows]     name:        %#v
                    [INFO][terraform-provider-windows]     mac_address: %#v
                    [INFO][terraform-provider-windows]     ipv4 {
                    [INFO][terraform-provider-windows]         interface_metric: %#v
                    [INFO][terraform-provider-windows]         dhcp_enabled:     %#v
                    [INFO][terraform-provider-windows]         ip_address:       %#v
                    [INFO][terraform-provider-windows]         gateway:          %#v
                    [INFO][terraform-provider-windows]         dns_addresses:    %#v
                    [INFO][terraform-provider-windows]     }
                    [INFO][terraform-provider-windows]     ipv6 {
                    [INFO][terraform-provider-windows]         interface_metric: %#v
                    [INFO][terraform-provider-windows]         dhcp_enabled:     %#v
                    [INFO][terraform-provider-windows]         ip_address:       %#v
                    [INFO][terraform-provider-windows]         gateway:          %#v
                    [INFO][terraform-provider-windows]         dns_addresses:    %#v
                    [INFO][terraform-provider-windows]     }
                    [INFO][terraform-provider-windows]     dns {
                    [INFO][terraform-provider-windows]         register_connection_address: %#v
                    [INFO][terraform-provider-windows]         register_connection_suffix:  %#v
                    [INFO][terraform-provider-windows]     }
`       ,
        id,
        name,
        macAddress,
        ipv4["interface_metric"],
        ipv4["dhcp_enabled"],
        ipv4["ip_address"],
        ipv4["gateway"],
        ipv4["dns_addresses"],
        ipv6["interface_metric"],
        ipv6["dhcp_enabled"],
        ipv6["ip_address"],
        ipv6["gateway"],
        ipv6["dns_addresses"],
        dns["registerConnectionAddress"],
        dns["registerConnectionSuffix"],
    )

    // import network_adapter
    log.Printf("[INFO][terraform-provider-windows] importing windows_network_adapter %q into terraform state\n", id)

    naQuery := new(api.NetworkAdapter)
    naQuery.Name = name

    networkAdapter, err := c.ReadNetworkAdapter(naQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows] cannot import windows_network_adapter %q into terraform state\n", id)
        return err
    }

    // save existing config
    setOriginalNetworkAdapterProperties(d, networkAdapter)

    // update network_adapter
    if networkAdapter.MACAddress != macAddress {                                // ToDo !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

        log.Printf("[INFO][terraform-provider-windows] updating windows_network_adapter %q\n", id)

        naProperties := new(api.NetworkAdapter)
        expandNetworkAdapterProperties(naProperties, d)

        err := c.UpdateNetworkAdapter(networkAdapter, naProperties)
        if err != nil {
            log.Printf("[ERROR][terraform-provider-windows] cannot update windows_network_adapter %q\n", id)
            log.Printf("[ERROR][terraform-provider-windows] cannot import windows_network_adapter %q into terraform state\n", id)
            return err
        }
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] created network_adapter %q\n", id)
    return resourceWindowsNetworkAdapterRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id       := d.Id()
    name     := d.Get("name").(string)
    original := tfutil.GetResource(d, "original")   // make sure new terraform state includes 'original' from the old terraform state when doing a terraform refresh

    log.Printf("[INFO][terraform-provider-windows] reading windows_network_adapter %q\n", id)

    // read network_adapter
    naQuery := new(api.NetworkAdapter)
    naQuery.Name = name

    networkAdapter, err := c.ReadNetworkAdapter(naQuery)
    if err != nil {
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
    d.Set("original", []interface{}{ original })   // make sure new terraform state includes 'original' from the old terraform state when doing a terraform refresh

    log.Printf("[INFO][terraform-provider-windows] read network_adapter %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterUpdate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id          := d.Id()
    name        := d.Get("name").(string)
    macAddress := d.Get("mac_address").(string)
    ipv4       := tfutil.GetResource(d, "ipv4")
    ipv6       := tfutil.GetResource(d, "ipv6")
    dns        := tfutil.GetResource(d, "dns")

    log.Printf(`[INFO][terraform-provider-windows] updating windows_network_adapter %q
                    [INFO][terraform-provider-windows]     name:        %#v
                    [INFO][terraform-provider-windows]     mac_address: %#v
                    [INFO][terraform-provider-windows]     ipv4 {
                    [INFO][terraform-provider-windows]         interface_metric: %#v
                    [INFO][terraform-provider-windows]         dhcp_enabled:     %#v
                    [INFO][terraform-provider-windows]         ip_address:       %#v
                    [INFO][terraform-provider-windows]         gateway:          %#v
                    [INFO][terraform-provider-windows]         dns_addresses:    %#v
                    [INFO][terraform-provider-windows]     }
                    [INFO][terraform-provider-windows]     ipv6 {
                    [INFO][terraform-provider-windows]         interface_metric: %#v
                    [INFO][terraform-provider-windows]         dhcp_enabled:     %#v
                    [INFO][terraform-provider-windows]         ip_address:       %#v
                    [INFO][terraform-provider-windows]         gateway:          %#v
                    [INFO][terraform-provider-windows]         dns_addresses:    %#v
                    [INFO][terraform-provider-windows]     }
                    [INFO][terraform-provider-windows]     dns {
                    [INFO][terraform-provider-windows]         register_connection_address: %#v
                    [INFO][terraform-provider-windows]         register_connection_suffix:  %#v
                    [INFO][terraform-provider-windows]     }
`       ,
        id,
        name,
        macAddress,
        ipv4["interface_metric"],
        ipv4["dhcp_enabled"],
        ipv4["ip_address"],
        ipv4["gateway"],
        ipv4["dns_addresses"],
        ipv6["interface_metric"],
        ipv6["dhcp_enabled"],
        ipv6["ip_address"],
        ipv6["gateway"],
        ipv6["dns_addresses"],
        dns["register_connection_address"],
        dns["register_connection_suffix"],
    )

    // update network_adapter
    na := new(api.NetworkAdapter)
    na.Name = name

    naProperties := new(api.NetworkAdapter)
    expandNetworkAdapterProperties(naProperties, d)

    err := c.UpdateNetworkAdapter(na, naProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows] cannot update windows_network_adapter %q\n", id)
        return err
    }

    log.Printf("[INFO][terraform-provider-windows] updated network_adapter %q\n", id)
    return resourceWindowsNetworkAdapterRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsNetworkAdapterDelete(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id       := d.Id()
    name     := d.Get("name").(string)

    log.Printf("[INFO][terraform-provider-windows] deleting windows_network_adapter %q\n", id)
    log.Printf("[INFO][terraform-provider-windows] restore original config for windows_network_adapter %q\n", id)

    // restore network_adapter
    na := new(api.NetworkAdapter)
    na.Name = name

    naProperties := new(api.NetworkAdapter)
    expandOriginalNetworkAdapterProperties(naProperties, d)

    err := c.UpdateNetworkAdapter(na, naProperties)
    if err != nil {
        log.Printf("[WARNING][terraform-provider-windows] cannot restore original config for windows_network_adapter %q\n", id)
    }

    // set id
    d.SetId("")

    log.Printf("[INFO][terraform-provider-windows] deleted windows_network_adapter %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setNetworkAdapterProperties(d *schema.ResourceData, naProperties *api.NetworkAdapter) {
    d.Set("name", naProperties.Name)
    d.Set("mac_address", naProperties.MACAddress)

    d.Set("ipv4", flattenNetworkAdapterIPInterface(&naProperties.IPv4))

    d.Set("ipv6", flattenNetworkAdapterIPInterface(&naProperties.IPv6))

    dns := make(map[string]interface{})
    dns["register_connection_address"] = naProperties.DNS.RegisterConnectionAddress
    if naProperties.DNS.RegisterConnectionSuffix == "" {
        dns["register_connection_suffix"] = "<empty>"
    } else {
        dns["register_connection_suffix"] = naProperties.DNS.RegisterConnectionSuffix
    }

    d.Set("dns", []interface{}{ dns })

    d.Set("admin_status", naProperties.AdminStatus)
    d.Set("operational_status", naProperties.OperationalStatus)
    d.Set("connection_status", naProperties.ConnectionStatus)
    d.Set("connection_speed", naProperties.ConnectionSpeed)
    d.Set("is_physical", naProperties.IsPhysical)
}

func flattenNetworkAdapterIPInterface(ipProperties *api.NetworkAdapterIPInterface) []interface{} {
    ipInterface := make(map[string]interface{})

    ipInterface["interface_metric"]                        = ipProperties.InterfaceMetric
    ipInterface["interface_metric_obtained_automatically"] = ipProperties.InterfaceMetricObtainedAutomatically
    ipInterface["dhcp_enabled"]                            = ipProperties.DHCPEnabled

    l := len(ipProperties.IPAddresses)
    ipInterface_ipAddresses := make([]interface{}, l, l)
    for i, a := range ipProperties.IPAddresses {
        address := make(map[string]interface{})
        address["address"]        = a.Address
        address["prefix_length"]  = a.PrefixLength
        address["skip_as_source"] = a.SkipAsSource
        ipInterface_ipAddresses[i] = address
    }
    HashIpAddress := schema.HashResource(resourceWindowsNetworkAdapter().Schema["ipv4"].Elem.(*schema.Resource).Schema["ip_address"].Elem.(*schema.Resource))
    ipInterface["ip_address"]                        = schema.NewSet(HashIpAddress, ipInterface_ipAddresses)
    ipInterface["ip_address_obtained_automatically"] = ipProperties.IPAddressObtainedAutomatically

    l = len(ipProperties.Gateways)
    ipInterface_gateways := make([]interface{}, l, l)
    for i, g := range ipProperties.Gateways {
        gateway := make(map[string]interface{})
        gateway["address"]                             = g.Address
        gateway["route_metric"]                        = g.RouteMetric
        gateway["route_metric_obtained_automatically"] = g.RouteMetricObtainedAutomatically
        ipInterface_gateways[i] = gateway
    }
    HashGateway := schema.HashResource(resourceWindowsNetworkAdapter().Schema["ipv4"].Elem.(*schema.Resource).Schema["gateway"].Elem.(*schema.Resource))
    ipInterface["gateway"]                        = schema.NewSet(HashGateway, ipInterface_gateways)
    ipInterface["gateway_obtained_automatically"] = ipProperties.GatewayObtainedAutomatically

    ipInterface["dns_addresses"]                        = ipProperties.DNSAddresses
    ipInterface["dns_addresses_obtained_automatically"] = ipProperties.DNSAddressesObtainedAutomatically

    ipInterface["connection_status"] = ipProperties.ConnectionStatus
    ipInterface["connectivity"]      = ipProperties.Connectivity

    return []interface{}{ ipInterface }
}

func setOriginalNetworkAdapterProperties(d *schema.ResourceData, naProperties *api.NetworkAdapter) {
    original := make(map[string]interface{})

    original["mac_address"] = naProperties.MACAddress

    original["ipv4"] = flattenNetworkAdapterIPInterface(&naProperties.IPv4)

    original["ipv6"] = flattenNetworkAdapterIPInterface(&naProperties.IPv6)

    original_dns := make(map[string]interface{})
    original_dns["register_connection_address"] = naProperties.DNS.RegisterConnectionAddress
    original_dns["register_connection_suffix"]  = naProperties.DNS.RegisterConnectionSuffix

    original["dns"] = []interface{}{ original_dns }

    d.Set("original", []interface{}{ original })
}

func flattenOriginalNetworkAdapterIPInterface(ipProperties *api.NetworkAdapterIPInterface) []interface{} {
    original_ipInterface := make(map[string]interface{})

    original_ipInterface["interface_metric"]                        = ipProperties.InterfaceMetric
    original_ipInterface["interface_metric_obtained_automatically"] = ipProperties.InterfaceMetricObtainedAutomatically
    original_ipInterface["dhcp_enabled"]                            = ipProperties.DHCPEnabled

    l := len(ipProperties.IPAddresses)
    original_ipInterface_ipAddresses := make([]interface{}, l, l)
    for i, a := range ipProperties.IPAddresses {
        address := make(map[string]interface{})
        address["address"]        = a.Address
        address["prefix_length"]  = a.PrefixLength
        address["skip_as_source"] = a.SkipAsSource
        original_ipInterface_ipAddresses[i] = address
    }
    HashIpAddress := schema.HashResource(resourceWindowsNetworkAdapter().Schema["ipv4"].Elem.(*schema.Resource).Schema["ip_address"].Elem.(*schema.Resource))
    original_ipInterface["ip_address"]                        = schema.NewSet(HashIpAddress, original_ipInterface_ipAddresses)
    original_ipInterface["ip_address_obtained_automatically"] = ipProperties.IPAddressObtainedAutomatically

    l = len(ipProperties.Gateways)
    original_ipInterface_gateways := make([]interface{}, l, l)
    for i, g := range ipProperties.Gateways {
        gateway := make(map[string]interface{})
        gateway["address"]                             = g.Address
        gateway["route_metric"]                        = g.RouteMetric
        gateway["route_metric_obtained_automatically"] = g.RouteMetricObtainedAutomatically
        original_ipInterface_gateways[i] = gateway
    }
    HashGateway := schema.HashResource(resourceWindowsNetworkAdapter().Schema["ipv4"].Elem.(*schema.Resource).Schema["gateway"].Elem.(*schema.Resource))
    original_ipInterface["gateway"]                        = schema.NewSet(HashGateway, original_ipInterface_gateways)
    original_ipInterface["gateway_obtained_automatically"] = ipProperties.GatewayObtainedAutomatically

    original_ipInterface["dns_addresses"]                        = ipProperties.DNSAddresses
    original_ipInterface["dns_addresses_obtained_automatically"] = ipProperties.DNSAddressesObtainedAutomatically

    return []interface{}{ original_ipInterface }
}

//------------------------------------------------------------------------------

func expandNetworkAdapterProperties(naProperties *api.NetworkAdapter, d *schema.ResourceData) {
    if v, ok := d.GetOkExists("mac_address"); ok {
        naProperties.MACAddress = v.(string)
    }

    if _, ok := d.GetOk("ipv4"); ok {
        expandNetworkAdapterIPInterfaceProperties(&naProperties.IPv4, d, "ipv4.0")
    }

    if _, ok := d.GetOk("ipv6"); ok {
        expandNetworkAdapterIPInterfaceProperties(&naProperties.IPv6, d, "ipv6.0")
    }

    if _, ok := d.GetOk("dns"); ok {
        if v, ok := d.GetOkExists("register_connection_address"); ok {
            naProperties.DNS.RegisterConnectionAddress = v.(bool)
        }
        if v, ok := d.GetOkExists("register_connection_suffix"); ok {
            if v.(string) == "<empty>" {
                naProperties.DNS.RegisterConnectionSuffix = ""
            } else {
                naProperties.DNS.RegisterConnectionSuffix = v.(string)
            }
        }
    }
}

func expandNetworkAdapterIPInterfaceProperties(ipProperties *api.NetworkAdapterIPInterface, d *schema.ResourceData, prefix string) {
    if v, ok := d.GetOkExists(prefix + ".interface_metric"); ok {
        ipProperties.InterfaceMetric = uint32(v.(int))
        if d.HasChange(prefix + ".interface_metric") {
            // let new value of InterfaceMetric define the new value of InterfaceMetricObtainedAutomatically
            ipProperties.InterfaceMetricObtainedAutomatically = false
        } else {
            // let old value of InterfaceMetricObtainedAutomatically define the new value of InterfaceMetricObtainedAutomatically
            ipProperties.InterfaceMetricObtainedAutomatically = d.Get(prefix + ".interface_metric_obtained_automatically").(bool)
        }
    }

    if v, ok := d.GetOkExists(prefix + ".dhcp_enabled"); ok {
        ipProperties.DHCPEnabled = v.(bool)
    }

    if v, ok := d.GetOk(prefix + ".ip_address"); ok {
        len := v.(*schema.Set).Len()
        ipProperties.IPAddresses = make([]api.NetworkAdapterIPAddress, len, len)
        for i := 0; i < len; i++ {
            ipProperties.IPAddresses[i] = api.NetworkAdapterIPAddress{}
            ii := strconv.Itoa(i)
            if v, ok := d.GetOkExists(prefix + ".ip_address." + ii + ".address"); ok {
                ipProperties.IPAddresses[i].Address = v.(string)
            }
            if v, ok := d.GetOkExists(prefix + ".ip_address." + ii + ".prefix_length"); ok {
                ipProperties.IPAddresses[i].PrefixLength = uint8(v.(int))
            }
            if v, ok := d.GetOkExists(prefix + ".ip_address." + ii + ".skip_as_source"); ok {
                ipProperties.IPAddresses[i].SkipAsSource = v.(bool)
            }
        }
        if d.HasChange(prefix + ".ip_address_obtained_automatically") {
            // let new value of IPAddresses define the new value of IPAddressObtainedAutomatically
            ipProperties.IPAddressObtainedAutomatically = false
        } else {
            // let old value of IPAddressObtainedAutomatically define the new value of IPAddressObtainedAutomatically
            ipProperties.IPAddressObtainedAutomatically = d.Get(prefix + ".ip_address_obtained_automatically").(bool)
        }
    }

    if v, ok := d.GetOk(prefix + ".gateway"); ok {
        len := v.(*schema.Set).Len()
        ipProperties.Gateways = make([]api.NetworkAdapterGateway, len, len)
        for i := 0; i < len; i++ {
            ipProperties.Gateways[i] = api.NetworkAdapterGateway{}
            ii := strconv.Itoa(i)
            if v, ok := d.GetOkExists(prefix + ".gateway." + ii + ".address"); ok {
                ipProperties.Gateways[i].Address = v.(string)
            }
            if v, ok := d.GetOkExists(prefix + ".gateway." + ii + ".route_metric"); ok {
                ipProperties.Gateways[i].RouteMetric = uint16(v.(int))

                if d.HasChange(prefix + ".gateway." + ii + ".interface_metric") {
                    // let new value of InterfaceMetric define the new value of InterfaceMetricObtainedAutomatically
                    ipProperties.InterfaceMetricObtainedAutomatically = false
                } else {
                    // let old value of InterfaceMetricObtainedAutomatically define the new value of InterfaceMetricObtainedAutomatically
                    ipProperties.InterfaceMetricObtainedAutomatically = d.Get(prefix + ".gateway." + ii + ".interface_metric_obtained_automatically").(bool)
                }
            }
        }
        if d.HasChange(prefix + ".gateway_obtained_automatically") {
            // let new value of Gateways define the new value of GatewayObtainedAutomatically
            ipProperties.GatewayObtainedAutomatically = false
        } else {
            // let old value of GatewayObtainedAutomatically define the new value of GatewayObtainedAutomatically
            ipProperties.GatewayObtainedAutomatically = d.Get(prefix + ".gateway_obtained_automatically").(bool)
        }
    }

    if _, ok := d.GetOk(prefix + ".dns_addresses"); ok {
        ipProperties.DNSAddresses = tfutil.GetListOfStrings(d, prefix + ".dns_addresses")
        if d.HasChange(prefix + ".dns_addresses_obtained_automatically") {
            // let new value of DNSAddresses define the new value of DNSAddressesObtainedAutomatically
            ipProperties.DNSAddressesObtainedAutomatically = false
        } else {
            // let old value of DNSAddressesObtainedAutomatically define the new value of DNSAddressesObtainedAutomatically
            ipProperties.DNSAddressesObtainedAutomatically = d.Get(prefix + ".dns_addresses_obtained_automatically").(bool)
        }
    }
}

func expandOriginalNetworkAdapterProperties(naProperties *api.NetworkAdapter, d *schema.ResourceData) {
    original := tfutil.GetResource(d, "original")

    naProperties.MACAddress = original["mac_address"].(string)

    expandOriginalNetworkAdapterIPInterfaceProperties(&naProperties.IPv4, d, "original.0.ipv4.0")

    expandOriginalNetworkAdapterIPInterfaceProperties(&naProperties.IPv6, d, "original.0.ipv6.0")

    original_dns := tfutil.ExpandResource(original, "dns")
    naProperties.DNS.RegisterConnectionAddress = original_dns["register_connection_address"].(bool)
    naProperties.DNS.RegisterConnectionSuffix  = original_dns["register_connection_suffix"].(string)
}

func expandOriginalNetworkAdapterIPInterfaceProperties(ipProperties *api.NetworkAdapterIPInterface, d *schema.ResourceData, prefix string) {
    original_ipInterface := d.Get(prefix).(map[string]interface{})

    ipProperties.InterfaceMetric                      = uint32(original_ipInterface["interface_metric"].(int))
    ipProperties.InterfaceMetricObtainedAutomatically = original_ipInterface["interface_metric_obtained_automatically"].(bool)

    ipProperties.DHCPEnabled = original_ipInterface["dhcp_enabled"].(bool)

    original_ipInterface_ipAddresses := tfutil.ExpandSetOfResources(original_ipInterface, "ip_address")
    l := len(original_ipInterface_ipAddresses)
    ipProperties.IPAddresses = make([]api.NetworkAdapterIPAddress, l, l)
    for i, a := range original_ipInterface_ipAddresses {
        ipProperties.IPAddresses[i] = api.NetworkAdapterIPAddress{
            Address:      a["address"].(string),
            PrefixLength: uint8(a["prefix_length"].(int)),
            SkipAsSource: a["skip_as_source"].(bool),
        }
    }
    ipProperties.IPAddressObtainedAutomatically = original_ipInterface["ip_address_obtained_automatically"].(bool)

    original_ipInterface_gateways := tfutil.ExpandSetOfResources(original_ipInterface, "gateway")
    l = len(original_ipInterface_gateways)
    ipProperties.Gateways = make([]api.NetworkAdapterGateway, l, l)
    for i, g := range original_ipInterface_gateways {
        ipProperties.Gateways[i] = api.NetworkAdapterGateway{
            Address:                          g["address"].(string),
            RouteMetric:                      uint16(g["route_metric"].(int)),
            RouteMetricObtainedAutomatically: g["route_metric_obtained_automatically"].(bool),
        }
    }
    ipProperties.GatewayObtainedAutomatically = original_ipInterface["gateway_obtained_automatically"].(bool)

    ipProperties.DNSAddresses                      = tfutil.ExpandListOfStrings(original_ipInterface, "dns_addresses")
    ipProperties.DNSAddressesObtainedAutomatically = original_ipInterface["dns_addresses_obtained_automatically"].(bool)
}

//------------------------------------------------------------------------------
