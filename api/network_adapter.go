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

type NetworkAdapter struct {
    GUID                string

    Name                string
    OldName             string
    NewName             string

    MACAddress          string
    PermanentMACAddress string

    DNSClient           []NetworkAdapterDNSClient

    // status
    AdminStatus         string
    OperationalStatus   string
    ConnectionStatus    string
    ConnectionSpeed     string
    IsPhysical          bool
}

type NetworkAdapterDNSClient struct {
    RegisterConnectionAddress bool
    RegisterConnectionSuffix  string
}

//------------------------------------------------------------------------------

func (c *WindowsClient) ReadNetworkAdapter(naQuery *NetworkAdapter) (naProperties *NetworkAdapter, err error) {
    if ( naQuery.GUID    == "" ) &&
       ( naQuery.Name    == "" ) &&
       ( naQuery.OldName == "" ) {
        return nil, fmt.Errorf("[ERROR][terraform-provider-windows/api/ReadNetworkAdapter(naQuery)] empty 'naQuery'")
    }

    return readNetworkAdapter(c, naQuery)
}

func (c *WindowsClient) UpdateNetworkAdapter(naQuery *NetworkAdapter, naProperties *NetworkAdapter) error {
    if naQuery.GUID == "" {
        return fmt.Errorf("[ERROR][terraform-provider-windows/api/UpdateNetworkAdapter(naQuery)] missing 'naQuery.GUID'")
    }

    return updateNetworkAdapter(c, naQuery, naProperties)
}

//------------------------------------------------------------------------------

func readNetworkAdapter(c *WindowsClient, naQuery *NetworkAdapter) (naProperties *NetworkAdapter, err error) {
    // find id
    var id interface{}
    if naQuery.GUID               != "" { id = naQuery.GUID               } else
    if naQuery.Name               != "" { id = naQuery.Name               } else
    if naQuery.OldName            != "" { id = naQuery.OldName            }

    // convert query to JSON
    naQueryJSON, err := json.Marshal(naQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkAdapter()] cannot cannot convert 'naQuery' to json for network_adapter %#v\n", id)
        return nil, err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    c.Lock.Lock()
    err = runner.Run(c, readNetworkAdapterScript, readNetworkAdapterArguments{
        NAQueryJSON: string(naQueryJSON),
    }, &stdout, &stderr)
    c.Lock.Unlock()
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkAdapter()] cannot read network_adapter %#v\n", id)
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkAdapter()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkAdapter()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkAdapter()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/readNetworkAdapter()] runner: %s", stderr.String())
        }

        return nil, err
    }
    log.Printf("[INFO][terraform-provider-windows/api/readNetworkAdapter()] read network_adapter %#v \n%s", id, stdout.String())

    // convert stdout-JSON to naProperties
    naProperties = new(NetworkAdapter)
    err = json.Unmarshal(stdout.Bytes(), naProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkAdapter()] cannot convert json to 'naProperties' for network_adapter %#v\n", id)
        return nil, err
    }

    return naProperties, nil
}

type readNetworkAdapterArguments struct{
    NAQueryJSON string
}

var readNetworkAdapterScript = script.New("readNetworkAdapter", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $naQuery = ConvertFrom-Json -InputObject '{{.NAQueryJSON}}'
    $guid    = $naQuery.GUID
    $name    = $naQuery.Name
    $oldName = $naQuery.OldName

    if ( $guid -ne "" ) {
        $id = $guid
        $networkAdapter = Get-NetAdapter -IncludeHidden -ErrorAction 'Ignore' | where { $_.InstanceID -eq "{$guid}" }
    }
    elseif ( $name -ne "" ) {
        $id = $name
        $networkAdapter = Get-NetAdapter -Name $name -ErrorAction 'Ignore'
    }
    elseif ( $oldName -ne "" ) {
        $id = $oldName
        $networkAdapter = Get-NetAdapter -Name $oldName -ErrorAction 'Ignore'
    }
    if ( -not $networkAdapter ) {
        throw "cannot find network_adapter '$id'"
    }

    # prepare result
    $naProperties = @{
        GUID                = $networkAdapter.InstanceID.Trim("{}")
        Name                = $networkAdapter.Name
        MACAddress          = $networkAdapter.MacAddress
        PermanentMACAddress = $networkAdapter.PermanentAddress -replace '..(?!$)', '$&-'
        DNSClient           = @()
        AdminStatus         = $networkAdapter.AdminStatus.ToString()
        OperationalStatus   = $networkAdapter.ifOperStatus.ToString()
        ConnectionStatus    = $networkAdapter.MediaConnectionState.ToString()
        ConnectionSpeed     = $networkAdapter.LinkSpeed.ToString()
        IsPhysical          = $networkAdapter.ConnectorPresent
    }

    $dnsClient = Get-DNSClient -InterfaceAlias $networkAdapter.Name -ErrorAction 'Ignore'
    if ( $dnsClient ) {
        $naProperties.DNSClient += @{
            RegisterConnectionAddress = $dnsClient.RegisterThisConnectionsAddress
            RegisterConnectionSuffix  = ""
        }

        if ( $dnsClient.UseSuffixWhenRegistering ) {
            $naProperties.DNSClient[0].RegisterConnectionSuffix = $dnsClient.ConnectionSpecificSuffix
        }
    }

    Write-Output $( ConvertTo-Json -InputObject $naProperties -Depth 100 )
`)

//------------------------------------------------------------------------------

func updateNetworkAdapter(c *WindowsClient, naQuery *NetworkAdapter, naProperties *NetworkAdapter) error {
    // find id
    id := naQuery.GUID

    // convert query to JSON
    naQueryJSON, err := json.Marshal(naQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter(naQuery, naProperties)] cannot cannot convert 'naQuery' to json for network_adapter %#v\n", id)
        return err
    }

    // convert properties to JSON
    naPropertiesJSON, err := json.Marshal(naProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter(naQuery, naProperties)] cannot cannot convert 'naProperties' to json for network_adapter %#v\n", id)
        return err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    c.Lock.Lock()
    err = runner.Run(c, updateNetworkAdapterScript, updateNetworkAdapterArguments{
        NAQueryJSON:      string(naQueryJSON),
        NAPropertiesJSON: string(naPropertiesJSON),
    }, &stdout, &stderr)
    c.Lock.Unlock()
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] cannot update network_adapter %#v\n", id)
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/updateNetworkAdapter()] runner: %s", stderr.String())
        }

        return err
    }
    log.Printf("[INFO][terraform-provider-windows/api/updateNetworkAdapter()] updated network_adapter %#v\n", id)

    return nil
}

type updateNetworkAdapterArguments struct{
    NAQueryJSON      string
    NAPropertiesJSON string
}

var updateNetworkAdapterScript = script.New("updateNetworkAdapter", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $naQuery = ConvertFrom-Json -InputObject '{{.NAQueryJSON}}'
    $guid = $naQuery.GUID

    $networkAdapter = Get-NetAdapter -IncludeHidden -ErrorAction 'Ignore' | where { $_.InstanceID -eq "{$guid}" }
    if ( -not $networkAdapter ) {
        throw "cannot find network_adapter '$guid'"
    }

    $naProperties = ConvertFrom-Json -InputObject '{{.NAPropertiesJSON}}'

    if ( $naProperties.DNSClient.Count -ne 0 ) {
        $dnsClient = Get-DNSClient -InterfaceIndex $networkAdapter.InterfaceIndex -ErrorAction 'Ignore'
        if ( !$dnsClient ) {
            throw "cannot set 'dns_client'-properties for network_adapter '$guid', network_adapter is not a dns_client"
        }
    }

    if ( ( $naProperties.NewName -ne "" ) -and ( $naProperties.NewName -ne $networkAdapter.Name ) ) {
        Rename-NetAdapter -InputObject $networkAdapter -NewName $naProperties.NewName -Confirm:$false | Out-Default
    }

    if ( $naProperties.MACAddress -ne "" ) {
        if ( $naProperties.MACAddress -eq ( $networkAdapter.PermanentAddress -replace '..(?!$)', '$&-' ) ) {
            Set-NetAdapter -InputObject $networkAdapter -MacAddress "" -Confirm:$false | Out-Default
        }
        else {
            Set-NetAdapter -InputObject $networkAdapter -MacAddress $naProperties.MACAddress -Confirm:$false | Out-Default
        }
    }

    if ( $naProperties.DNSClient.Count -ne 0 ) {
        $arguments = @{
            RegisterThisConnectionsAddress = $naProperties.DNSClient[0].RegisterConnectionAddress
            UseSuffixWhenRegistering       = $false
            ConnectionSpecificSuffix       = ""
        }
        if ( $naProperties.DNSClient[0].RegisterConnectionSuffix -ne "" ) {
            $arguments.UseSuffixWhenRegistering = $true
            $arguments.ConnectionSpecificSuffix = $naProperties.DNSClient[0].RegisterConnectionSuffix
        }

        Set-DNSClient -InterfaceAlias $networkAdapter.Name @arguments -Confirm:$false | Out-Default
    }

`)

//------------------------------------------------------------------------------
