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
    Name                       string
    MACAddress                 string

    IPv4                       NetworkAdapterIPInterface
    IPv6                       NetworkAdapterIPInterface
    DNS                        NetworkAdapterDNSClient

    // status
    AdminStatus                string
    OperationalStatus          string
    ConnectionStatus           string
    ConnectionSpeed            string
    IsPhysical                 bool
}

type NetworkAdapterIPInterface struct {
    InterfaceMetric                      uint32
    InterfaceMetricObtainedAutomatically bool

    DHCPEnabled                          bool

    IPAddresses                          []NetworkAdapterIPAddress
    IPAddressObtainedAutomatically       bool
    Gateways                             []NetworkAdapterGateway
    GatewayObtainedAutomatically         bool
    DNSAddresses                         []string
    DNSAddressesObtainedAutomatically    bool

    ConnectionStatus                     string
    Connectivity                         string
}

type NetworkAdapterIPAddress struct {
    Address      string
    PrefixLength uint8
    SkipAsSource bool
}

type NetworkAdapterGateway struct {
    Address                          string
    RouteMetric                      uint16
    RouteMetricObtainedAutomatically bool
}

type NetworkAdapterDNSClient struct {
    RegisterConnectionAddress bool
    RegisterConnectionSuffix  string
}

//------------------------------------------------------------------------------

func (c *WindowsClient) ReadNetworkAdapter(naQuery *NetworkAdapter) (naProperties *NetworkAdapter, err error) {
    if naQuery.Name == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-windows/api/ReadNetworkAdapter(naQuery)] missing 'naQuery.Name'")
    }

    return readNetworkAdapter(c, naQuery)
}

func (c *WindowsClient) UpdateNetworkAdapter(na *NetworkAdapter, naProperties *NetworkAdapter) error {
    if na.Name == "" {
        return fmt.Errorf("[ERROR][terraform-provider-windows/api/UpdateNetworkAdapter(na)] missing 'na.Name'")
    }

    return updateNetworkAdapter(c, na, naProperties)
}

//------------------------------------------------------------------------------

func readNetworkAdapter(c *WindowsClient, naQuery *NetworkAdapter) (naProperties *NetworkAdapter, err error) {
    // find id
    id := naQuery.Name

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
    err = runner.Run(c, readNetworkAdapterScript, readNetworkAdapterArguments{
        NAQueryJSON: string(naQueryJSON),
    }, &stdout, &stderr)
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
    log.Printf("[INFO][terraform-provider-windows/api/readNetworkAdapter()] read network_adapter %#v\n", naQuery.Name, stdout.String())

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

    $name = $naQuery.Name

    $networkAdapter = Get-NetAdapter -Name $name -ErrorAction 'Ignore'
    if ( -not $networkAdapter ) {
        throw "cannot find network_adapter '$name'"
    }

    $interfaceIndex = $networkAdapter.InterfaceIndex

    # prepare result
    $naProperties = @{
        Name              = $networkAdapter.Name
        MACAddress        = $networkAdapter.MacAddress

        IPv4              = @{}
        IPv6              = @{}
        DNS               = @{}

        AdminStatus       = $networkAdapter.AdminStatus.ToString()
        OperationalStatus = $networkAdapter.ifOperStatus.ToString()
        ConnectionStatus  = $networkAdapter.MediaConnectionState.ToString()
        ConnectionSpeed   = $networkAdapter.LinkSpeed.ToString()
        IsPhysical        = $networkAdapter.ConnectorPresent
    }

    $ipv4Binding = Get-NetAdapterBinding -Name $name -ComponentID 'ms_tcpip' -ErrorAction 'Ignore'
    if ( $ipv4Binding -and $ipv4Binding.Enabled ) {

        $naProperties.IPv4 = @{
            IPAddresses = @()
            Gateways    = @()
            DNServers   = @()
        }

        $ipv4Interface = Get-NetIPInterface -InterfaceIndex $interfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
        if ( $ipv4Interface ) {
            $naProperties.IPv4.InterfaceMetric                      = $ipv4Interface.InterfaceMetric
            $naProperties.IPv4.InterfaceMetricObtainedAutomatically = $( $ipv4Interface.AutomaticMetric -eq 'Enabled' )
            $naProperties.IPv4.DHCPEnabled                          = $( $ipv4Interface.Dhcp -eq 'Enabled' )

            $naProperties.IPv4.ConnectionStatus                     = $ipv4Interface.ConnectionState.ToString()
        }

        $ipv4Address = Get-NetIPAddress -InterfaceIndex $interfaceIndex -AddressFamily 'IPv4' -AddressState 'Preferred' -ErrorAction 'Ignore'
        if ( $ipv4Address ) {
            $ipv4Address | foreach {
                $naProperties.IPv4.IPAddresses += @{
                    Address      = $_.IPAddress
                    PrefixLength = $_.PrefixLength
                    SkipAsSource = $_.SkipAsSource
                }
            }
            $naProperties.IPv4.IPAddressObtainedAutomatically = $( $ipv4Address.SuffixOrigin -ne 'Manual' )
        }

        $ipv4Route = Get-NetRoute -InterfaceIndex $interfaceIndex -AddressFamily 'IPv4' -DestinationPrefix '0.0.0.0/0' -ErrorAction 'Ignore'
        if ( $ipv4Route ) {
            $ipv4Route | foreach {
                $naProperties.IPv4.Gateways += @{
                    Address                          = $_.NextHop
                    RouteMetric                      = $_.RouteMetric
                    RouteMetricObtainedAutomatically = $( $_.AutomaticMetric -eq 'Enabled' )
                }
            }
            if ( $naProperties.IPv4.Gateways.Length -gt 0 ) {
                $naProperties.IPv4.GatewayObtainedAutomatically = -not $( Get-NetRoute -InterfaceIndex $interfaceIndex -AddressFamily 'IPv4' -DestinationPrefix '0.0.0.0/0' -NextHop $naProperties.IPv4.Gateways[0].Address -PolicyStore 'PersistentStore' )
            }
        }

        $ipv4DNServerAddresses = Get-DNSClientServerAddress -InterfaceIndex $interfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
        if ( $ipv4DNServerAddresses ) {
            $ipv4DNServerAddresses.ServerAddresses | foreach {
                $naProperties.IPv4.DNServers += $_
            }
            $naProperties.IPv4.DNSObtainedAutomatically = -not -not $( ( netsh interface ipv4 show dns $name | select-string 'DNS servers configured through DHCP:' ) -match 'DNS servers configured through DHCP:' )
        }
    }

    $ipv6Binding = Get-NetAdapterBinding -Name $name -ComponentID 'ms_tcpip6' -ErrorAction 'Ignore'
    if ( $ipv6Binding -and $ipv6Binding.Enabled ) {

        $naProperties.IPv6 = @{
            IPAddresses = @()
            Gateways    = @()
            DNServers   = @()
        }

        $ipv6Interface = Get-NetIPInterface -InterfaceIndex $interfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
        if ( $ipv6Interface ) {
            $naProperties.IPv6.InterfaceMetric                      = $ipv6Interface.InterfaceMetric
            $naProperties.IPv6.InterfaceMetricObtainedAutomatically = $( $ipv6Interface.AutomaticMetric -eq 'Enabled' )
            $naProperties.IPv6.DHCPEnabled                          = $( $ipv6Interface.Dhcp -eq 'Enabled' )
        }

        $ipv6Address = Get-NetIPAddress -InterfaceIndex $interfaceIndex -AddressFamily 'IPv6' -AddressState 'Preferred' -ErrorAction 'Ignore'
        if ( $ipv6Address ) {
            $ipv6Address | foreach {
                $naProperties.IPv6.IPAddresses += @{
                    Address      = $_.IPAddress
                    PrefixLength = $_.PrefixLength
                    SkipAsSource = $_.SkipAsSource
                }
                $naProperties.IPv6.RouteMetricObtainedAutomatically = $( $_.AutomaticMetric -eq 'Enabled' )
            }
            $naProperties.IPv6.IPAddressObtainedAutomatically = $( $ipv6Address.SuffixOrigin -ne 'Manual' )
        }

        $ipv6Route = Get-NetRoute -InterfaceIndex $interfaceIndex -AddressFamily 'IPv6' -DestinationPrefix '::/0' -ErrorAction 'Ignore'
        if ( $ipv6Route ) {
            $ipv6Route | foreach {
                $naProperties.IPv6.Gateways += @{
                    Address                          = $_.NextHop
                    RouteMetric                      = $_.RouteMetric
                    RouteMetricObtainedAutomatically = $( $_.AutomaticMetric -eq 'Enabled' )
                }
            }
            if ( $naProperties.IPv4.Gateways.Length -gt 0 ) {
                $naProperties.IPv6.GatewayObtainedAutomatically = -not $( Get-NetRoute -InterfaceIndex $interfaceIndex -AddressFamily 'IPv6' -DestinationPrefix '::/0' -NextHop $naProperties.IPv6.Gateways[0].Address -PolicyStore 'PersistentStore' )
            }
        }

        $ipv6DNServerAddresses = Get-DNSClientServerAddress -InterfaceIndex $interfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
        if ( $ipv6DNServerAddresses ) {
            $ipv6DNServerAddresses.ServerAddresses | foreach {
                $naProperties.IPv6.DNServers += $_
            }
            $naProperties.IPv6.DNSObtainedAutomatically = -not -not $( ( netsh interface ipv6 show dns $name | select-string 'DNS servers configured through DHCP:' ) -match 'DNS servers configured through DHCP:' )
        }

    }

    if ( ( $ipv4Binding -and $ipv4Binding.Enabled ) -or ( $ipv6Binding -and $ipv6Binding.Enabled ) ) {

        $dnsClient = Get-DNSClient -InterfaceIndex $interfaceIndex -ErrorAction 'Ignore'
        if ( $dnsClient ) {
            $naProperties.DNS.RegisterConnectionAddress = $dnsClient.RegisterThisConnectionsAddress
            if ( $dnsClient.UseSuffixWhenRegistering ) {
                $naProperties.DNS.RegisterConnectionSuffix = $dnsClient.ConnectionSpecificSuffix
            }
            else {
                $naProperties.DNS.RegisterConnectionSuffix = ""
            }
        }

        $connectionProfile = Get-ConnectionProfile -InterfaceIndex $interfaceIndex -ErrorAction 'Ignore'
        if ( $connectionProfile ) {
            if ( $ipv4Binding -and $ipv4Binding.Enabled ) {
                $naProperties.IPv4.Connectivity = $connectionProfile.IPv4Connectivity.ToString()
            }
            if ( $ipv6Binding -and $ipv6Binding.Enabled ) {
                $naProperties.IPv6.Connectivity = $connectionProfile.IPv6Connectivity.ToString()
            }
        }
    }

    Write-Output $( ConvertTo-Json -InputObject $naProperties -Depth 100 )
`)

//------------------------------------------------------------------------------

func updateNetworkAdapter(c *WindowsClient, naQuery *NetworkAdapter, naProperties *NetworkAdapter) error {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // convert query to JSON
    naQueryJSON, err := json.Marshal(naQuery)
    if err != nil {
        return err
    }

    // convert properties to JSON
    naPropertiesJSON, err := json.Marshal(naProperties)
    if err != nil {
        return err
    }

    // run script
    err = runner.Run(c, updateNetworkAdapterScript, updateNetworkAdapterArguments{
        NAQueryJSON:      string(naQueryJSON),
        NAPropertiesJSON: string(naPropertiesJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] cannot update network_adapter\n")
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkAdapter()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/updateNetworkAdapter()] runner: %s", stderr.String())
        }

        return err
    }

    log.Printf("[INFO][terraform-provider-windows/api/updateNetwork()] updated network_adapter\n")
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

    $name = $naQuery.Name

    $networkAdapter = Get-NetAdapter -Name $name -ErrorAction 'Ignore'
    if ( -not $networkAdapter ) {
        throw "cannot find network_adapter '$name'"
    }

    $interfaceIndex = $networkAdapter.InterfaceIndex
    $naProperties = ConvertFrom-Json -InputObject '{{.NAPropertiesJSON}}'

    Set-NetAdapter -Name $name -MacAddress $naProperties.MACAddress -Confirm:$false | Out-Default

    if ( $naProperties.IPv4.InterfaceDisabled ) {
        Disable-NetAdapterBinding -Name $name -ComponentID 'ms_tcpip' -Confirm:$false | Out-Default
    }
    else {
        Enable-NetAdapterBinding -Name $name -ComponentID 'ms_tcpip' -Confirm:$false | Out-Default
    }

    if ( $naProperties.IPv6.InterfaceDisabled ) {
        Disable-NetAdapterBinding -Name $name -ComponentID 'ms_tcpip6' -Confirm:$false | Out-Default
    }
    else {
        Enable-NetAdapterBinding -Name $name -ComponentID 'ms_tcpip6' -Confirm:$false | Out-Default
    }

    if ( -not $naProperties.IPv4.InterfaceDisabled ) {
        if ( $naProperties.IPv4.InterfaceMetricObtainedAutomatically -or ( $naProperties.IPv4.InterfaceMetric -eq 0 ) ) {
            Set-NetIPInterface -InterfaceIndex $interfaceIndex -AddressFamily 'IPv4' -AutomaticMetric Enable -Confirm:$false | Out-Default
        }
        else {
            Set-NetIPInterface -InterfaceIndex $interfaceIndex -AddressFamily 'IPv4' -InterfaceMetric $naProperties.IPv4.InterfaceMetric -Confirm:$false | Out-Default
        }

        if ( -not $naProperties.IPv4.DNSObtainedAutomatically -and ( $naProperties.IPv4.DNServers.Length -gt 0 ) ) {
            Set-DnsClientServerAddress -InterfaceIndex $interfaceIndex -ServerAddresses $naProperties.IPv4.DNServers -Confirm:$false | Out-Default
        }
    }

    if ( -not $naProperties.IPv6.InterfaceDisabled ) {
        if ( $naProperties.IPv6.InterfaceMetricObtainedAutomatically -or ( $naProperties.IPv6.InterfaceMetric -eq 0 ) ) {
            Set-NetIPInterface -InterfaceIndex $interfaceIndex -AddressFamily 'IPv6' -AutomaticMetric Enable -Confirm:$false | Out-Default
        }
        else {
            Set-NetIPInterface -InterfaceIndex $interfaceIndex -AddressFamily 'IPv6' -InterfaceMetric $naProperties.IPv6.InterfaceMetric -Confirm:$false | Out-Default
        }

        if ( -not $naProperties.IPv6.DNSObtainedAutomatically -and ( $naProperties.IPv6.DNServers.Length -gt 0 ) ) {
            Set-DnsClientServerAddress -InterfaceIndex $interfaceIndex -ServerAddresses $naProperties.IPv6.DNServers -Confirm:$false | Out-Default
        }
    }

    if ( -not $naProperties.IPv4.InterfaceDisabled -or -not $naProperties.IPv6.InterfaceDisabled ) {
        Set-DNSClient -InterfaceIndex $interfaceIndex -RegisterThisConnectionsAddress $naProperties.RegisterConnectionAddress
        if ( $naProperties.RegisterConnectionSuffix -eq "" ) {
            Set-DNSClient -InterfaceIndex $interfaceIndex -ConnectionSpecificSuffix "" -UseSuffixWhenRegistering $false -Confirm:$false | Out-Default
        }
        else {
            Set-DNSClient -InterfaceIndex $interfaceIndex -ConnectionSpecificSuffix $naProperties.RegisterConnectionSuffix -UseSuffixWhenRegistering $true -Confirm:$false | Out-Default
        }
    }
`)

//------------------------------------------------------------------------------
