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

    "github.com/stefaanc/terraform-provider-windows/api"
)

//------------------------------------------------------------------------------

func dataSourceWindowsComputer() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "new_name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "dns_client": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: dataSourceWindowsComputerDNSClient(),
            },

            "reboot_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "reboot_pending_details": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: dataSourceWindowsComputerRebootPendingDetails(),
            },

            "network_adapter_names": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem:     &schema.Schema{ Type: schema.TypeString },
            },
            "network_connection_names": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem:     &schema.Schema{ Type: schema.TypeString },
            },
        },

        Read:   dataSourceWindowsComputerRead,
    }
}

func dataSourceWindowsComputerDNSClient() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "suffix_search_list": &schema.Schema{
                Type:     schema.TypeList,
                Elem:     &schema.Schema{ Type: schema.TypeString },
                Computed: true,
            },
            "enable_devolution": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "devolution_level": &schema.Schema{
                Type:     schema.TypeInt,   // uint32
                Computed: true,
            },
        },
    }
}

func dataSourceWindowsComputerRebootPendingDetails() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "reboot_required": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "post_reboot_reporting": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "dvd_reboot_signal": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "reboot_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "reboot_in_progress": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "packages_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "services_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "update_exe_volatile": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "computer_rename_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "file_rename_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "netlogon_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "current_reboot_attemps": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceWindowsComputerRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id := fmt.Sprintf("//%s/computer", host)

    log.Printf("[INFO][terraform-provider-windows] reading windows_computer %q\n", id)

    // read
    computer, err := c.ReadComputer()
    if err != nil {
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot read windows_computer %q\n", id)
        return err
    }

    // set properties
    setDataComputerProperties(d, computer)

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] read windows_computer %q\n", id)
    return nil
}

func setDataComputerProperties(d *schema.ResourceData, cProperties *api.Computer) {
    d.Set("name", cProperties.Name)
    d.Set("new_name", cProperties.NewName)

    dnsClient := make(map[string]interface{})
    dnsClient["suffix_search_list"] = cProperties.DNSClient.SuffixSearchList
    dnsClient["enable_devolution"]  = cProperties.DNSClient.EnableDevolution
    dnsClient["devolution_level"]   = cProperties.DNSClient.DevolutionLevel
    d.Set("dns_client", []interface{}{ dnsClient })

    d.Set("reboot_pending", cProperties.RebootPending)

    rebootPendingDetails := make(map[string]interface{})
    rebootPendingDetails["reboot_required"]         = cProperties.RebootPendingDetails.RebootRequired
    rebootPendingDetails["post_reboot_reporting"]   = cProperties.RebootPendingDetails.PostRebootReporting
    rebootPendingDetails["dvd_reboot_signal"]       = cProperties.RebootPendingDetails.DVDRebootSignal
    rebootPendingDetails["reboot_pending"]          = cProperties.RebootPendingDetails.RebootPending
    rebootPendingDetails["reboot_in_progress"]      = cProperties.RebootPendingDetails.RebootInProgress
    rebootPendingDetails["packages_pending"]        = cProperties.RebootPendingDetails.PackagesPending
    rebootPendingDetails["services_pending"]        = cProperties.RebootPendingDetails.ServicesPending
    rebootPendingDetails["update_exe_volatile"]     = cProperties.RebootPendingDetails.UpdateExeVolatile
    rebootPendingDetails["computer_rename_pending"] = cProperties.RebootPendingDetails.ComputerRenamePending
    rebootPendingDetails["file_rename_pending"]     = cProperties.RebootPendingDetails.FileRenamePending
    rebootPendingDetails["netlogon_pending"]        = cProperties.RebootPendingDetails.NetlogonPending
    rebootPendingDetails["current_reboot_attemps"]  = cProperties.RebootPendingDetails.CurrentRebootAttemps
    d.Set("reboot_pending_details", []interface{}{ rebootPendingDetails })

    d.Set("network_adapter_names", cProperties.NetworkAdapterNames)
    d.Set("network_connection_names", cProperties.NetworkConnectionNames)
}

//------------------------------------------------------------------------------
