//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package api

import (
    "bytes"
    "errors"
    "fmt"
    "encoding/json"
    "log"
    "strings"

    "github.com/stefaanc/golang-exec/runner"
    "github.com/stefaanc/golang-exec/script"
)

//------------------------------------------------------------------------------

type LinkIPInterface struct {
    Index                  uint32
    Alias                  string
    Description            string
    GUID                   string

    MACAddress             string
    NetworkAdapterName     string
    VNetworkAdapterName    string

    NetworkConnectionNames []string
    VSwitchName            string
    ComputerName           string
}

//------------------------------------------------------------------------------

func (c *WindowsClient) ReadLinkIPInterface(iQuery *LinkIPInterface) (iProperties *LinkIPInterface, err error) {
    if iQuery.Index == 0 &&
       iQuery.Alias == "" &&
       iQuery.Description == "" &&
       iQuery.GUID == "" &&
       iQuery.MACAddress == "" &&
       iQuery.NetworkAdapterName == "" &&
       iQuery.VNetworkAdapterName == "" {

        return nil, fmt.Errorf("[ERROR][terraform-provider-windows/api/ReadLinkIPInterface(iQuery)] empty 'iQuery'")
    }

    return readLinkIPInterface(c, iQuery)
}

//------------------------------------------------------------------------------

func readLinkIPInterface(c *WindowsClient, iQuery *LinkIPInterface) (iProperties *LinkIPInterface, err error) {
    // find id
    var id interface{}
    if        iQuery.Index != 0                { id = iQuery.Index
    } else if iQuery.Alias != ""               { id = iQuery.Alias
    } else if iQuery.Description != ""         { id = iQuery.Description
    } else if iQuery.GUID != ""                { id = iQuery.GUID
    } else if iQuery.MACAddress != ""          { id = iQuery.MACAddress
    } else if iQuery.NetworkAdapterName != ""  { id = iQuery.NetworkAdapterName
    } else if iQuery.VNetworkAdapterName != "" { id = iQuery.VNetworkAdapterName
    }

    // convert query to JSON
    iQueryJSON, err := json.Marshal(iQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readLinkIPInterface()] cannot convert 'iQuery' to json for link_ip_interface %#v\n", id)
        return nil, err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readLinkIPInterfaceScript, readLinkIPInterfaceArguments{
        IQueryJSON: string(iQueryJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/readLinkIPInterface()] cannot read link_ip_interface %#v\n", id)
        log.Printf("[ERROR][terraform-provider-windows/api/readLinkIPInterface()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/readLinkIPInterface()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/readLinkIPInterface()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/readLinkIPInterface()] runner: %s", stderr.String())
        }

        return nil, err
    }
    log.Printf("[INFO][terraform-provider-windows/api/readLinkIPInterface()] read link_ip_interface %#v \n%s", id, stdout.String())

    // convert stdout-JSON to properties
    iProperties = new(LinkIPInterface)
    err = json.Unmarshal(stdout.Bytes(), iProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readLinkIPInterface()] cannot convert json to 'iProperties' for link_ip_interface %#v\n", id)
        return nil, err
    }

    return iProperties, nil
}

type readLinkIPInterfaceArguments struct{
    IQueryJSON string
}

var readLinkIPInterfaceScript = script.New("readLinkIPInterface", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    # convert vsProperties to JSON
    $iQuery = ConvertFrom-Json -InputObject '{{.IQueryJSON}}'

    $index               = $iQuery.Index
    $alias               = $iQuery.Alias
    $description         = $iQuery.Description
    $guid                = $iQuery.GUID
    $macAddress          = $iQuery.MACAddress
    $networkAdapterName  = $iQuery.NetworkAdapterName
    $vnetworkAdapterName = $iQuery.VNetworkAdapterName

    # find network-adapter
    if ( $index -ne 0 ) {
        $id = $index
        $networkAdapter = Get-NetAdapter -InterfaceIndex $index -ErrorAction 'Ignore'
    }
    elseif ( $alias -ne "" ) {
        $id = $alias
        Get-NetAdapter -IncludeHidden | foreach {
            if ( $alias -eq $_.InterfaceAlias ) {
                $networkAdapter = $_
            }
        }
    }
    elseif ( $description -ne "" ) {
        $id = $description
        $networkAdapter = Get-NetAdapter -InterfaceDescription $description -ErrorAction 'Ignore'
    }
    elseif ( $guid -ne "" ) {
        $id = $guid
        Get-NetAdapter -IncludeHidden | foreach {
            if ( $guid -eq $_.InterfaceGUID.Trim("{}") ) {
                $networkAdapter = $_
            }
        }
    }
    elseif ( $macAddress -ne "" ) {
        $id = $macAddress
        Get-NetAdapter -IncludeHidden | foreach {
            if ( $macAddress -eq $_.MacAddress ) {
                $networkAdapter = $_
            }
        }
    }
    elseif ( $networkAdapterName -ne "" ) {
        $id = $networkAdapterName
        $networkAdapter = Get-NetAdapter -Name $networkAdapterName -ErrorAction 'Ignore'
    }
    elseif ( $vnetworkAdapterName -ne "" ) {
        $id = $vnetworkAdapterName
        $vnetworkAdapter = Get-VMNetworkAdapter -ManagementOS -Name $vnetworkAdapterName -ErrorAction 'Ignore'
        if ( $vnetworkAdapter ) {
            Get-NetAdapter -IncludeHidden | foreach {
                if ( $vnetworkAdapter.DeviceID -eq $_.DeviceID ) {
                    $networkAdapter = $_
                }
            }
        }
    }
    if ( -not $networkAdapter ) {
        throw "cannot find link_ip_interface '$id'"
    }

    # find vnetwork-adapter
    if ( ( -not $vnetworkAdapter ) -and ( $networkAdapter.DriverDescription -eq "Hyper-V Virtual Ethernet Adapter" ) ) {
        Get-VMNetworkAdapter  -ManagementOS | foreach {
            if ( $networkAdapter.DeviceID -eq $_.DeviceID ) {
                $vnetworkAdapter = $_
            }
        }
    }

    $iProperties = @{
        Index                  = $networkAdapter.InterfaceIndex
        Alias                  = $networkAdapter.InterfaceAlias
        Description            = $networkAdapter.InterfaceDescription
        GUID                   = $networkAdapter.InterfaceGUID
        MACAddress             = $networkAdapter.MacAddress
        NetworkAdapterName     = $networkAdapter.Name

        NetworkConnectionNames = @()
        ComputerName           = $networkAdapter.SystemName
    }

    if ( $vnetworkAdapter ) {
        $iProperties.VNetworkAdapterName = $vnetworkAdapter.Name
        $iProperties.VSwitchName         = $vnetworkAdapter.SwitchName
    }

    Get-NetConnectionProfile -InterfaceIndex $networkAdapter.InterfaceIndex -ErrorAction 'Ignore' | foreach {
        $iProperties.NetworkConnectionNames += $_.Name
    }

    Write-Output $( ConvertTo-Json -InputObject $iProperties -Depth 100 )
`)

//------------------------------------------------------------------------------
