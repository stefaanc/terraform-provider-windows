//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package windows

import (
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/terraform"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

//------------------------------------------------------------------------------

func Provider() terraform.ResourceProvider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema {
            "type": &schema.Schema{
                Description: "The type of connection to the windows-computer: \"local\" or \"ssh\"",
                Type:     schema.TypeString,
                Optional: true,
                Default: "local",

                ValidateFunc: validation.StringInSlice([]string{ "local", "ssh" }, true),
            },

            // ssh
            "host": &schema.Schema{                                // config ignored when type is not "ssh"
                Description: "The windows-computer",
                Type:     schema.TypeString,
                Optional: true,
                Default: "localhost",
            },
            "port": &schema.Schema{                                // config ignored when type is not "ssh"
                Description: "The port for communication with the windows-computer",
                Type:     schema.TypeInt,
                Optional: true,
                Default:  22,

                ValidateFunc: validation.IntBetween(0, 65535),
            },
            "user": &schema.Schema{                                // config ignored when type is not "ssh"
                Description: "The user name for communication with the windows-computer",
                Type:     schema.TypeString,
                Optional: true,
                Default: "",
            },
            "password": &schema.Schema{                            // config ignored when type is not "ssh"
                Description: "The user password for communication with the windows-computer",
                Type:      schema.TypeString,
                Optional:  true,
                Default:   "",
                Sensitive: true,
            },
            "insecure": &schema.Schema{                            // config ignored when type is not "ssh"
                Description: "Allow insecure communication - disable checking of the certificate of the windows-computer",
                Type:     schema.TypeBool,
                Optional: true,
                Default: false,
            },
        },

        DataSourcesMap: map[string]*schema.Resource {
            "windows_computer": dataSourceWindowsComputer(),
            "windows_network_adapter": dataSourceWindowsNetworkAdapter(),
            "windows_network_connection": dataSourceWindowsNetworkConnection(),
            "windows_network_interface": dataSourceWindowsNetworkInterface(),
        },

        ResourcesMap: map[string]*schema.Resource{
            "windows_computer": resourceWindowsComputer(),
            "windows_network_adapter": resourceWindowsNetworkAdapter(),
            "windows_network_connection": resourceWindowsNetworkConnection(),
        },

        ConfigureFunc: providerConfigure,
    }
}

//------------------------------------------------------------------------------

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
    config := Config{
        Type:     strings.ToLower(d.Get("type").(string)),

        // ssh
        Host:     d.Get("host").(string),
        Port:     uint16(d.Get("port").(int)),
        User:     d.Get("user").(string),
        Password: d.Get("password").(string),
        Insecure: d.Get("insecure").(bool),
    }

    return config.Client()
}

//------------------------------------------------------------------------------
