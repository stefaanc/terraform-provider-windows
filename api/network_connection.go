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

type NetworkConnection struct {
    IPv4GatewayAddress  string
    IPv6GatewayAddress  string
    AllowDisconnect     bool

    Name                string

    ConnectionProfile   string
}

//------------------------------------------------------------------------------

func (c *WindowsClient) ReadNetworkConnection(ncQuery *NetworkConnection) (ncProperties *NetworkConnection, err error) {
    if ncQuery.IPv4GatewayAddress == "" &&
       ncQuery.IPv6GatewayAddress == "" &&
       ncQuery.Name == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-windows/api/ReadNetworkConnection(ncQuery)] empty 'ncQuery'")
    }

    return readNetworkConnection(c, ncQuery)
}

func (c *WindowsClient) UpdateNetworkConnection(nc *NetworkConnection, ncProperties *NetworkConnection) error {
    if nc.IPv4GatewayAddress == "" &&
       nc.IPv6GatewayAddress == "" {
        return fmt.Errorf("[ERROR][terraform-provider-windows/api/UpdateNetworkConnection(n)] missing 'nc.IPv4GatewayAddress' and 'nc.IPv6GatewayAddress'")
    }

    return updateNetworkConnection(c, nc, ncProperties)
}

//------------------------------------------------------------------------------

func readNetworkConnection(c *WindowsClient, ncQuery *NetworkConnection) (ncProperties *NetworkConnection, err error) {
    // find id
    var id interface{}
    if        ncQuery.IPv4GatewayAddress != "" { id = ncQuery.IPv4GatewayAddress
    } else if ncQuery.IPv6GatewayAddress != "" { id = ncQuery.IPv6GatewayAddress
    } else if ncQuery.Name != ""               { id = ncQuery.Name
    }

    // convert query to JSON
    ncQueryJSON, err := json.Marshal(ncQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkConnection(ncQuery)] cannot cannot convert 'ncQuery' to json for network_connection %#v\n", id)
        return nil, err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readNetworkConnectionScript, readNetworkConnectionArguments{
        NCQueryJSON: string(ncQueryJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkConnection()] cannot read network_connection %#v\n", id)
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkConnection()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkConnection()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkConnection()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/readNetworkConnection()] runner: %s", stderr.String())
        }

        return nil, err
    }
    log.Printf("[INFO][terraform-provider-windows/api/readNetworkConnection()] read network_connection %#v \n%s", id, stdout.String())

    // convert stdout-JSON to ncProperties
    ncProperties = new(NetworkConnection)
    err = json.Unmarshal(stdout.Bytes(), ncProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readNetworkConnection()] cannot convert json to 'nProperties' for network_connection %#v\n", id)
        return nil, err
    }

    return ncProperties, nil
}

type readNetworkConnectionArguments struct{
    NCQueryJSON string
}

var readNetworkConnectionScript = script.New("readNetworkConnection", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $netRouteTimeout = 250

    function findGatewayAddress {
        param( $gatewayRoutes, $interfaceIndexes )

        if ( $gatewayRoutes ) {
            $gatewayRoutes.NextHop | Sort-Object | Get-Unique | foreach {
                $nextHop = $_
                $reduced = $gatewayRoutes | where { $_.NextHop -eq $nextHop }
                if (
                    $reduced -and
                    (
                        ( ( $reduced -is    [array] ) -and ( $reduced.Length -eq $interfaceIndexes.Length ) ) -or
                        ( ( $reduced -isnot [array] ) -and ( $interfaceIndexes.Length -eq 1 ) )
                    )
                ) {
                    if ( $gatewayAddress -eq $null ) {
                        $gatewayAddress = $nextHop
                    }
                    else {
                        if ( $gatewayAddress -isnot [array] ) {
                            $gatewayAddress = @( $gatewayAddress )
                        }
                        $gatewayAddress += $nextHop
                    }
                }
            }
        }

        $gatewayAddress
    }

    function findGatewayAddressWithDisconnections {
        param( $gatewayRoutes, $interfaceIndexes, $networkConnectionProfileName )

        if ( $gatewayRoutes ) {
            # for an interface-index
            # try to delete gateway-routes one by one
            # and verify if network-profile disappears or if the interface-index disappears from the network-profile's interface-indexes
            $index = $interfaceIndexes[0]
            $gatewayRoutes | where { $_.InterfaceIndex -eq $index } | foreach {
                if ( $gatewayAddress -eq $null ) {
                    $arguments = @{
                        InterfaceIndex    = $_.InterfaceIndex
                        DestinationPrefix = $_.DestinationPrefix
                        NextHop           = $_.NextHop
                    }
                    $manual = ( -not -not ( Get-NetRoute @arguments -PolicyStore PersistentStore -ErrorAction 'Ignore' ) )
                    $dhcp   = ( ( Get-NetIPInterface -InterfaceIndex $index -AddressFamily $_.AddressFamily -ErrorAction 'Ignore' | Select-Object -ExpandProperty 'Dhcp' ) -eq 'Enabled' )

                    Remove-NetRoute @arguments -Confirm:$false | Out-Null
                    Start-Sleep -MilliSeconds $netRouteTimeout

                    $profile = Get-NetConnectionProfile -Name $networkConnectionProfileName -InterfaceIndex $index -ErrorAction 'Ignore'
                    if ( -not $profile ) {
                        $gatewayAddress = $_.NextHop
                    }

                    if ( $manual ) {
                        New-NetRoute @arguments -Confirm:$false | Out-Null
                    }
                    elseif ( $dhcp ) {
                        Set-NetIPInterface -InterfaceIndex $index -AddressFamily $_.AddressFamily -Dhcp 'Enabled'  -Confirm:$false | Out-Null
                    }
                }
            }
            Start-Sleep -MilliSeconds $netRouteTimeout
        }

        $gatewayAddress
    }

    $ncQuery = ConvertFrom-Json -InputObject '{{.NCQueryJSON}}'

    $ipv4GatewayAddress = $ncQuery.IPv4GatewayAddress
    $ipv6GatewayAddress = $ncQuery.IPv6GatewayAddress
    $allowDisconnection = $ncQuery.AllowDisconnection
    $name               = $ncQuery.Name

    if ( $ipv4GatewayAddress -ne "" ) {
        $id = $ipv4GatewayAddress
        $ipv4GatewayRoute = Get-NetRoute -DestinationPrefix '0.0.0.0/0' -NextHop $ipv4GatewayAddress -ErrorAction 'Ignore'
        if ( $ipv4GatewayRoute ) {
            $networkConnectionProfile = Get-NetConnectionProfile -InterfaceIndex $ipv4GatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
            if ( $networkConnectionProfile -is [array] ) {
                # cannot determine the exact network-profile
                $networkConnectionProfile = $null
            }
        }
    }
    if ( -not $networkConnectionProfile -and ( $ipv6GatewayAddress -ne "" ) ) {
        $id = $ipv6GatewayAddress
        $ipv6GatewayRoute = Get-NetRoute -DestinationPrefix '::/0' -NextHop $ipv6GatewayAddress -ErrorAction 'Ignore'
        if ( $ipv6GatewayRoute ) {
            $networkConnectionProfile = Get-NetConnectionProfile -InterfaceIndex $ipv6GatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
            if ( $networkConnectionProfile -is [array] ) {
                # cannot determine the exact network-profile
                $networkConnectionProfile = $null
            }
        }
    }
    if ( -not $networkConnectionProfile -and ( $name -ne "" ) ) {
        $id = $name
        $networkConnectionProfile = Get-NetConnectionProfile -Name $name -ErrorAction 'Ignore'
    }
    if ( -not $networkConnectionProfile ) {
        throw "cannot find network_connection '$id'"
    }

    # find ipv4-gateway-address
    if ( -not $ipv4GatewayRoute ) {
        $interfaceIndexes = $networkConnectionProfile.InterfaceIndex | Sort-Object | Get-Unique
        $gatewayRoutes    = Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction 'Ignore' | where { $interfaceIndexes -contains $_.InterfaceIndex }
        $ipv4GatewayAddress = findGatewayAddress $gatewayRoutes $interfaceIndexes
        if ( $ipv4GatewayAddress -is [array] ) {
            if ( -not $allowDisconnection ) {
                $ipv4GatewayAddress = ""
            }
            else {
                $gatewayRoutes = $gatewayRoutes | where { $ipv4GatewayAddress -contains $_.NextHop }
                $ipv4GatewayAddress = findGatewayAddressWithDisconnections $gatewayRoutes $interfaceIndexes $networkConnectionProfile.Name
            }
        }
    }

    # find ipv6-gateway-address
    if ( -not $ipv6GatewayRoute ) {
        $interfaceIndexes = $networkConnectionProfile.InterfaceIndex | Sort-Object | Get-Unique
        $gatewayRoutes    = Get-NetRoute -DestinationPrefix '::/0' -ErrorAction 'Ignore' | where { $interfaceIndexes -contains $_.InterfaceIndex }
        $ipv6GatewayAddress = findGatewayAddress $gatewayRoutes $interfaceIndexes
        if ( $ipv6GatewayAddress -is [array] ) {
            if ( -not $allowDisconnection ) {
                $ipv6GatewayAddress = ""
            }
            else {
                $gatewayRoutes = $gatewayRoutes | where { $ipv6GatewayAddress -contains $_.NextHop }
                $ipv6GatewayAddress = findGatewayAddressWithDisconnections $gatewayRoutes $interfaceIndexes $networkConnectionProfile.Name
            }
        }
    }

    $ncProperties = @{
        Name               = $networkConnectionProfile.Name
        IPv4GatewayAddress = $ipv4GatewayAddress
        IPv6GatewayAddress = $ipv6GatewayAddress
        ConnectionProfile  = $networkConnectionProfile.NetworkCategory.ToString()
    }

    Write-Output $( ConvertTo-Json -InputObject $ncProperties -Depth 100 )
`)

//------------------------------------------------------------------------------

func updateNetworkConnection(c *WindowsClient, ncQuery *NetworkConnection, ncProperties *NetworkConnection) error {
    // find id
    var id interface{}
    if        ncQuery.IPv4GatewayAddress != "" { id = ncQuery.IPv4GatewayAddress
    } else if ncQuery.IPv6GatewayAddress != "" { id = ncQuery.IPv6GatewayAddress
    } else if ncQuery.Name != ""               { id = ncQuery.Name
    }

    // convert nquery to JSON
    ncQueryJSON, err := json.Marshal(ncQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkConnection(ncQuery, ncProperties)] cannot cannot convert 'ncQuery' to json for network_connection %#v\n", id)
        return err
    }

    // convert properties to JSON
    ncPropertiesJSON, err := json.Marshal(ncProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkConnection(ncQuery, ncProperties)] cannot cannot convert 'ncProperties' to json for network_connection %#v\n", id)
        return err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, updateNetworkConnectionScript, updateNetworkConnectionArguments{
        NCQueryJSON:      string(ncQueryJSON),
        NCPropertiesJSON: string(ncPropertiesJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkConnection()] cannot update network_connection %#v\n", id)
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkConnection()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkConnection()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/updateNetworkConnection()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/updateNetworkConnection()] runner: %s", stderr.String())
        }

        return err
    }

    log.Printf("[INFO][terraform-provider-windows/api/updateNetworkConnection()] updated network_connection %#v\n", id)
    return nil
}

type updateNetworkConnectionArguments struct{
    NCQueryJSON      string
    NCPropertiesJSON string
}

var updateNetworkConnectionScript = script.New("updateNetworkConnection", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $netRouteTimeout = 250

    $ncQuery = ConvertFrom-Json -InputObject '{{.NCQueryJSON}}'

    $ipv4GatewayAddress = $ncQuery.IPv4GatewayAddress
    $ipv6GatewayAddress = $ncQuery.IPv6GatewayAddress
    $allowDisconnection = $ncQuery.AllowDisconnection
    $name               = $ncQuery.Name

    if ( $ipv4GatewayAddress -ne "" ) {
        $id = $ipv4GatewayAddress
        $ipv4GatewayRoute = Get-NetRoute -DestinationPrefix '0.0.0.0/0' -NextHop $ipv4GatewayAddress -ErrorAction 'Ignore'
        if ( $ipv4GatewayRoute ) {
            $networkConnectionProfile = Get-NetConnectionProfile -InterfaceIndex $ipv4GatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
            if ( $networkConnectionProfile -is [array] ) {
                # cannot determine the exact network-profile
                $networkConnectionProfile = $null
            }
        }
    }
    if ( -not $networkConnectionProfile -and ( $ipv6GatewayAddress -ne "" ) ) {
        $id = $ipv6GatewayAddress
        $ipv6GatewayRoute = Get-NetRoute -DestinationPrefix '::/0' -NextHop $ipv6GatewayAddress -ErrorAction 'Ignore'
        if ( $ipv6GatewayRoute ) {
            $networkConnectionProfile = Get-NetConnectionProfile -InterfaceIndex $ipv6GatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
            if ( $networkConnectionProfile -is [array] ) {
                # cannot determine the exact network-profile
                $networkConnectionProfile = $null
            }
        }
    }
    if ( -not $networkConnectionProfile -and ( $name -ne "" ) ) {
        $id = $name
        $networkConnectionProfile = Get-NetConnectionProfile -Name $name -ErrorAction 'Ignore'
    }
    if ( -not $networkConnectionProfile ) {
        throw "cannot find network_connection '$id'"
    }

    $ncProperties = ConvertFrom-Json -InputObject '{{.NCPropertiesJSON}}'

    if ( $networkConnectionProfile.Name -ne $ncProperties.Name ) {
        # update $networkConnectionProfile.Name in registry
        Get-ChildItem -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles\" | foreach {
            $itemProperty = Get-ItemProperty -Path $_.PSPath
            if ( $itemProperty.ProfileName -eq $networkConnectionProfile.Name ) {
                Set-ItemProperty -Path $_.PSPath -Name 'ProfileName' -Value $ncProperties.Name
            }
        }
    }

    Set-NetConnectionProfile -InputObject $networkConnectionProfile -NetworkCategory $ncProperties.ConnectionProfile -Confirm:$false | Out-Default
`)

//------------------------------------------------------------------------------
