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

type NetworkInterface struct {
    GUID                   string
    Index                  uint32
    Alias                  string
    Description            string

    MACAddress             string
    NetworkAdapterName     string
    VNetworkAdapterName    string

    NetworkConnectionNames []string
    VSwitchName            string
    ComputerName           string
}

//------------------------------------------------------------------------------

func (c *WindowsClient) ReadNetworkInterface(niQuery *NetworkInterface) (niProperties *NetworkInterface, err error) {
    if niQuery.GUID == "" &&
       niQuery.Index == 0 &&
       niQuery.Alias == "" &&
       niQuery.Description == "" &&
       niQuery.MACAddress == "" &&
       niQuery.NetworkAdapterName == "" &&
       niQuery.VNetworkAdapterName == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-windows/api/ReadNetworkInterface(niQuery)] empty 'niQuery'")
    }

    return readNetworkInterface(c, niQuery)
}

//------------------------------------------------------------------------------

func readNetworkInterface(c *WindowsClient, niQuery *NetworkInterface) (niProperties *NetworkInterface, err error) {
    // find id
    var id interface{}
    if niQuery.GUID                != "" { id = niQuery.GUID                } else
    if niQuery.Index               != 0  { id = niQuery.Index               } else
    if niQuery.Alias               != "" { id = niQuery.Alias               } else
    if niQuery.Description         != "" { id = niQuery.Description         } else
    if niQuery.MACAddress          != "" { id = niQuery.MACAddress          } else
    if niQuery.NetworkAdapterName  != "" { id = niQuery.NetworkAdapterName  } else
    if niQuery.VNetworkAdapterName != "" { id = niQuery.VNetworkAdapterName }

    // convert query to JSON
    niQueryJSON, err := json.Marshal(niQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkInterface()] cannot convert 'niQuery' to json for network_interface %#v\n", id)
        return nil, err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    c.Lock.Lock()
    err = runner.Run(c, readNetworkInterfaceScript, readNetworkInterfaceArguments{
        NIQueryJSON: string(niQueryJSON),
    }, &stdout, &stderr)
    c.Lock.Unlock()
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkInterface()] cannot read network_interface %#v\n", id)
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkInterface()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkInterface()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkInterface()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/readNetworkInterface()] runner: %s", stderr.String())
        }

        return nil, err
    }
    log.Printf("[INFO][terraform-provider-windows/api/readNetworkInterface()] read network_interface %#v \n%s", id, stdout.String())

    // convert stdout-JSON to properties
    niProperties = new(NetworkInterface)
    err = json.Unmarshal(stdout.Bytes(), niProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkInterface()] cannot convert json to 'niProperties' for network_interface %#v\n", id)
        return nil, err
    }

    return niProperties, nil
}

type readNetworkInterfaceArguments struct{
    NIQueryJSON string
}

var readNetworkInterfaceScript = script.New("readNetworkInterface", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $niQuery = ConvertFrom-Json -InputObject '{{.NIQueryJSON}}'
    $guid                = $niQuery.GUID
    $index               = $niQuery.Index
    $alias               = $niQuery.Alias
    $description         = $niQuery.Description
    $macAddress          = $niQuery.MACAddress
    $networkAdapterName  = $niQuery.NetworkAdapterName
    $vnetworkAdapterName = $niQuery.VNetworkAdapterName

    if ( $guid -ne "" ) {
        $id = $guid
        $networkAdapter = Get-NetAdapter -IncludeHidden | where { $_.InterfaceGUID -eq "{$guid}" }
    }
    elseif ( $index -ne 0 ) {
        $id = $index
        $networkAdapter = Get-NetAdapter -InterfaceIndex $index -ErrorAction 'Ignore'
    }
    elseif ( $alias -ne "" ) {
        $id = $alias
        $networkAdapter = Get-NetAdapter -IncludeHidden | where { $_.InterfaceAlias -eq $alias }
    }
    elseif ( $description -ne "" ) {
        $id = $description
        $networkAdapter = Get-NetAdapter -InterfaceDescription $description -ErrorAction 'Ignore'
    }
    elseif ( $macAddress -ne "" ) {
        $id = $macAddress
        $networkAdapter = Get-NetAdapter -IncludeHidden | where { $_.MacAddress -eq $macAddress }
        if ( $networkAdapter.Length -gt 1 ) {
            $networkAdapter = $null
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
            $networkAdapter = Get-NetAdapter -IncludeHidden | where { $_.DeviceID -eq $vnetworkAdapter.DeviceID }
        }
    }
    if ( -not $networkAdapter ) {
        throw "cannot find network_interface '$id'"
    }

    # find vnetwork-adapter
    if ( ( -not $vnetworkAdapter ) -and ( $networkAdapter.DriverDescription -eq "Hyper-V Virtual Ethernet Adapter" ) ) {
        $vnetworkAdapter = Get-VMNetworkAdapter -ManagementOS | where { $_.DeviceID -eq $networkAdapter.DeviceID }
    }

    $niProperties = @{
        GUID                   = $networkAdapter.InterfaceGUID.Trim("{}")
        Index                  = $networkAdapter.InterfaceIndex
        Alias                  = $networkAdapter.InterfaceAlias
        Description            = $networkAdapter.InterfaceDescription
        MACAddress             = $networkAdapter.MacAddress
        NetworkAdapterName     = $networkAdapter.Name

        NetworkConnectionNames = @()
        ComputerName           = $networkAdapter.SystemName
    }

    if ( $vnetworkAdapter ) {
        $niProperties.VNetworkAdapterName = $vnetworkAdapter.Name
        $niProperties.VSwitchName         = $vnetworkAdapter.SwitchName
    }

    Get-NetConnectionProfile -InterfaceIndex $networkAdapter.InterfaceIndex -ErrorAction 'Ignore' | foreach {
        $niProperties.NetworkConnectionNames += $_.Name
    }

    Write-Output $( ConvertTo-Json -InputObject $niProperties -Depth 100 )
`)

//------------------------------------------------------------------------------
