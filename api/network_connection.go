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
    GUID               string

    IPv4GatewayAddress string
    IPv6GatewayAddress string

    Name               string
    OldName            string
    NewName            string

    AllowDisconnect    bool

    ConnectionProfile  string

    IPv4Connectivity   string
    IPv6Connectivity   string
}

//------------------------------------------------------------------------------

func (c *WindowsClient) ReadNetworkConnection(ncQuery *NetworkConnection) (ncProperties *NetworkConnection, err error) {
    if ( ncQuery.GUID == "")                &&
       ( ncQuery.IPv4GatewayAddress == "" ) &&
       ( ncQuery.IPv6GatewayAddress == "" ) &&
       ( ncQuery.Name == "" )               &&
       ( ncQuery.OldName == "" )            {
        return nil, fmt.Errorf("[ERROR][terraform-provider-windows/api/ReadNetworkConnection(ncQuery)] empty 'ncQuery'")
    }

    return readNetworkConnection(c, ncQuery)
}

func (c *WindowsClient) UpdateNetworkConnection(ncQuery *NetworkConnection, ncProperties *NetworkConnection) error {
    if ncQuery.GUID == "" {
        return fmt.Errorf("[ERROR][terraform-provider-windows/api/UpdateNetworkConnection(ncQuery)] missing 'ncQuery.GUID'")
    }

    return updateNetworkConnection(c, ncQuery, ncProperties)
}

//------------------------------------------------------------------------------

func readNetworkConnection(c *WindowsClient, ncQuery *NetworkConnection) (ncProperties *NetworkConnection, err error) {
    // find id
    var id interface{}
    if        ncQuery.GUID != ""               { id = ncQuery.GUID
    } else if ncQuery.IPv4GatewayAddress != "" { id = ncQuery.IPv4GatewayAddress
    } else if ncQuery.IPv6GatewayAddress != "" { id = ncQuery.IPv6GatewayAddress
    } else if ncQuery.Name != ""               { id = ncQuery.Name
    } else if ncQuery.OldName != ""            { id = ncQuery.OldName
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
    c.Lock.Lock()
    err = runner.Run(c, readNetworkConnectionScript, readNetworkConnectionArguments{
        NCQueryJSON: string(ncQueryJSON),
    }, &stdout, &stderr)
    c.Lock.Unlock()
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
    $dhcpTimeout = 1000
    $dhcpMaxRetries = 60
    $networkTimeout = 1000
    $networkMaxRetries = 60
    $connectivityTimeout = 5000

    function findGatewayAddress {
        param( $gatewayRoutes, $networkConnectionProfile )

        # given a list of gateway-routes for an address-family and a connection-profile
        # build a list of interfaces for the connection-profile
        # verify that there is only one connection-profile that matches all these interfaces, otherwise it is impossible to determine the gateway
        # build a list of gateway-routes for these interfaces
        # for each gateway in the list of gateway-routes
        # build a reduced list of gateway-routes for this gateway
        # if the reduced list contains a route for every interface then this is a candidate gateway-address for the connection-profile
        # if there is only one candidate then this is the gateway for the connection-profile

        $gatewayAddress = $null

        $interfaceIndexes = $networkConnectionProfile.InterfaceIndex | Sort-Object | Get-Unique
        $profileNames = ( Get-NetConnectionProfile -InterfaceIndex $interfaceIndexes[0] -ErrorAction 'Ignore' ).Name | Sort-Object | Get-Unique
        $interfaceIndexes | foreach {
            $pns = ( Get-NetConnectionProfile -InterfaceIndex $_ -ErrorAction 'Ignore' ).Name | Sort-Object | Get-Unique
            $profileNames = $profileNames | where { $pns -contains $_ }
        }

        if ( $profileNames.Count -eq 1 ) {
            $gatewayRoutes = $gatewayRoutes | where { $interfaceIndexes -contains $_.InterfaceIndex }
            $gatewayRoutes.NextHop | Sort-Object | Get-Unique | foreach {
                $nextHop = $_
                $reduced = $gatewayRoutes | where { $_.NextHop -eq $nextHop }
                if ( $reduced -and                                              # !!! remark that $reduced.Count doesn't work if only one item - don't know why this is !!!!!!!!!!!!!!!!!!!!!!!!
                     ( ( $reduced -is    [array] ) -and ( $reduced.Count -eq $interfaceIndexes.Count ) ) -or
                     ( ( $reduced -isnot [array] ) -and ( $interfaceIndexes.Count -eq 1 ) )
                   ) {

                    if ( $gatewayAddress -eq $null ) {
                        $gatewayAddress = $nextHop
                    }
                    else {
                        $gatewayAddress = ""
                    }
                }
            }
        }

        $gatewayAddress
    }

    function findGatewayAddressWithDisconnections {
        param( $gatewayRoutes, $networkConnectionProfile )

        # given a list of gateway-routes for an address-family and a connection-profile
        # try to delete gateway-routes one by one
        # and verify if connection-profile disappears when building a new list of profiles for the interface of the gateway-route

        $gatewayAddress = $null

        $componentID = if ( $gatewayRoutes[0].AddressFamily -eq 'IPv4' ) { 'ms_tcpip6' } else { 'ms_tcpip' }
        $binding = Get-NetAdapterBinding -Name $gatewayRoutes[0].InterfaceAlias -ComponentID $componentID -ErrorAction 'Ignore' | Select-Object -ExpandProperty 'Enabled'
        if ( $binding ) {
            Set-NetAdapterBinding -Name $gatewayRoutes[0].InterfaceAlias -ComponentID $componentID -Enabled $false -Confirm:$false | Out-Null
        }

        $gatewayRoutes | foreach {
            if ( -not $gatewayAddress ) {
                $profileNames = ( Get-NetConnectionProfile -InterfaceIndex $_.InterfaceIndex -ErrorAction 'Ignore' ).Name | Sort-Object | Get-Unique
                if ( $profileNames.Count -eq 1 ) {
                    $n = $profileNames
                }
                else {
                    $n = findConnectionProfileWithDisconnections $_
                }

                if ( $n -eq $networkConnectionProfile.Name ) {
                    $gatewayAddress = $_.NextHop
                }
            }
        }

        if ( $binding ) {
            Set-NetAdapterBinding -Name $gatewayRoutes[0].InterfaceAlias -ComponentID $componentID -Enabled $binding -Confirm:$false | Out-Null
            $networkRetries = $dhcpMaxRetries
            while ( $true ) {
                Start-Sleep -MilliSeconds $networkTimeout
                if ( $componentID -eq 'ms_tcpip' ) {
                    $interface = Get-NetIPInterface -InterfaceIndex $gatewayRoutes[0].InterfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
                }
                else {
                    $interface = Get-NetIPInterface -InterfaceIndex $gatewayRoutes[0].InterfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
                }
                $restored = ( $interface.ConnectionState.ToString() -eq 'Connected' )
                if ( $restored ) {
                    break
                }

                $networkRetries -= 1
                if ( $networkRetries -eq 0 ) {
                    break
                }
            }
            Start-Sleep -MilliSeconds $connectivityTimeout   # wait a little bit more after 'Connected' to get Connectivity
        }

        $gatewayAddress
    }

    function findConnectionProfileWithDisconnections {
        param( $gatewayRoute )

        # given a gateway-route
        # build a list of profile-names for the interface of the gateway-route
        # delete the gateway-route
        # and find the profile-name that disappears when building a new list of profile-names for the interface of the gateway-route

        $componentID = if ( $gatewayRoute.AddressFamily -eq 'IPv4' ) { 'ms_tcpip6' } else { 'ms_tcpip' }
        $binding = Get-NetAdapterBinding -Name $gatewayRoute.InterfaceAlias -ComponentID $componentID -ErrorAction 'Ignore' | Select-Object -ExpandProperty 'Enabled'
        if ( $binding ) {
            Set-NetAdapterBinding -Name $gatewayRoute.InterfaceAlias -ComponentID $componentID -Enabled $false -Confirm:$false | Out-Null
        }

        $profileNames = ( Get-NetConnectionProfile -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore' ).Name | Sort-Object | Get-Unique
        if ( $profileNames.Count -eq 1 ) {

            $name = $profileNames

        }
        else {

            $arguments = @{
                InterfaceIndex    = $gatewayRoute.InterfaceIndex
                DestinationPrefix = $gatewayRoute.DestinationPrefix
                NextHop           = $gatewayRoute.NextHop
            }
            $manual = ( -not -not ( Get-NetRoute @arguments -PolicyStore PersistentStore -ErrorAction 'Ignore' ) )
            $dhcp   = ( ( Get-NetIPInterface -InterfaceIndex $gatewayRoute.InterfaceIndex -AddressFamily $gatewayRoute.AddressFamily -ErrorAction 'Ignore' | Select-Object -ExpandProperty 'Dhcp' ) -eq 'Enabled' )

            Remove-NetRoute @arguments -Confirm:$false | Out-Null
            Start-Sleep -MilliSeconds $netRouteTimeout

            $reducedProfileNames = ( Get-NetConnectionProfile -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore' ).Name
            $name = $profileNames | where { $reducedProfileNames -notcontains $_ }

            if ( $manual ) {
                New-NetRoute @arguments -Confirm:$false | Out-Null
            }
            elseif ( $dhcp ) {
                $dhcpRetries = $dhcpMaxRetries
                while ( $true ) {
                    if ( ( $dhcpRetries % 5 ) -eq 0 ) {
                        Set-NetIPInterface -InterfaceIndex $gatewayRoute.InterfaceIndex -AddressFamily $gatewayRoute.AddressFamily -Dhcp 'Disabled' -Confirm:$false | Out-Null
                        Set-NetIPInterface -InterfaceIndex $gatewayRoute.InterfaceIndex -AddressFamily $gatewayRoute.AddressFamily -Dhcp 'Enabled' -Confirm:$false | Out-Null
                    }

                    Start-Sleep -MilliSeconds $dhcpTimeout
                    $restored = ( -not -not ( Get-NetRoute @arguments -ErrorAction 'Ignore' ) )
                    if ( $restored ) {
                        break
                    }

                    $dhcpRetries -= 1
                    if ( $dhcpRetries -eq 0 ) {
                        break
                    }
                }
            }

        }

        if ( $binding ) {
            Set-NetAdapterBinding -Name $gatewayRoute.InterfaceAlias -ComponentID $componentID -Enabled $binding -Confirm:$false | Out-Null
            $networkRetries = $dhcpMaxRetries
            while ( $true ) {
                Start-Sleep -MilliSeconds $networkTimeout
                if ( $componentID -eq 'ms_tcpip' ) {
                    $interface = Get-NetIPInterface -InterfaceIndex $gatewayRoute.InterfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
                }
                else {
                    $interface = Get-NetIPInterface -InterfaceIndex $gatewayRoute.InterfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
                }
                $restored = ( $interface.ConnectionState.ToString() -eq 'Connected' )
                if ( $restored ) {
                    break
                }

                $networkRetries -= 1
                if ( $networkRetries -eq 0 ) {
                    break
                }
            }
            Start-Sleep -MilliSeconds $connectivityTimeout   # wait a little bit more after 'Connected' to get Connectivity
        }

        $name
    }

    $ncQuery = ConvertFrom-Json -InputObject '{{.NCQueryJSON}}'
    $guid               = $ncQuery.GUID
    $ipv4GatewayAddress = $ncQuery.IPv4GatewayAddress
    $ipv6GatewayAddress = $ncQuery.IPv6GatewayAddress
    $name               = $ncQuery.Name
    $oldName            = $ncQuery.OldName
    $allowDisconnect    = $ncQuery.AllowDisconnect

    if ( $guid -ne "" ) {
        $id = $guid
        $registryProfile = Get-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles\{$guid}" -ErrorAction Ignore
        if ( $registryProfile ) {
            $n = ( Get-ItemProperty -Path $registryProfile.PSPath -Name 'ProfileName' ).ProfileName
            if ( $n ) {
                $networkConnectionProfile = Get-NetConnectionProfile -Name $n -ErrorAction 'Ignore'
            }
        }
    }
    elseif ( $ipv4GatewayAddress -ne "" ) {
        $id = $ipv4GatewayAddress
        $gatewayRoutes = Get-NetRoute -DestinationPrefix '0.0.0.0/0' -NextHop $ipv4GatewayAddress -ErrorAction 'Ignore'
        if ( $gatewayRoutes ) {
            $gatewayRoute = $gatewayRoutes[0]
            $profileNames = ( Get-NetConnectionProfile -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore' ).Name | Sort-Object | Get-Unique
            if ( $profileNames.Count -eq 1 ) {
                $networkConnectionProfile = Get-NetConnectionProfile -Name $profileNames -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
            }
            elseif ( -not $allowDisconnect ) {
                # cannot determine the exact network-profile
                $networkConnectionProfile = $null
            }
            else {
                $n = findConnectionProfileWithDisconnections $gatewayRoute
                if ( $n ) {
                    $networkConnectionProfile = Get-NetConnectionProfile -Name $n -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
                }
            }
        }
    }
    elseif ( $ipv6GatewayAddress -ne "" ) {
        $id = $ipv6GatewayAddress
        $gatewayRoutes = Get-NetRoute -DestinationPrefix '::/0' -NextHop $ipv6GatewayAddress -ErrorAction 'Ignore'
        if ( $gatewayRoutes ) {
            $gatewayRoute = $gatewayRoutes[0]
            $profileNames = ( Get-NetConnectionProfile -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore' ).Name | Sort-Object | Get-Unique
            if ( $profileNames.Count -eq 1 ) {
                $networkConnectionProfile = Get-NetConnectionProfile -Name $profileNames -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
            }
            elseif ( -not $allowDisconnect ) {
                # cannot determine the exact network-profile
                $networkConnectionProfile = $null
            }
            else {
                $n = findConnectionProfileWithDisconnections $gatewayRoute
                if ( $n ) {
                    $networkConnectionProfile = Get-NetConnectionProfile -Name $n -InterfaceIndex $gatewayRoute.InterfaceIndex -ErrorAction 'Ignore'
                }
            }
        }
    }
    elseif ( $name -ne "" ) {
        $id = $name
        $networkConnectionProfile = Get-NetConnectionProfile -Name $name -ErrorAction 'Ignore'
    }
    elseif ( $oldName -ne "" ) {
        $id = $oldName
        $networkConnectionProfile = Get-NetConnectionProfile -Name $oldName -ErrorAction 'Ignore'
    }
    if ( -not $networkConnectionProfile ) {
        throw "cannot find network_connection '$id'"
    }

    # find guid
    if ( -not $guid -or ( $id -ne $guid ) ) {
        Get-ChildItem -Path 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles' -ErrorAction Ignore | foreach {
            $n = ( Get-ItemProperty -Path $_.PSPath -Name 'ProfileName' ).ProfileName
            if ( $n -eq $networkConnectionProfile[0].Name ) {
                $guid = $_.PSChildName.Trim("{}")
            }
        }
    }

    # find ipv4-gateway-address
    if ( -not $ipv4GatewayAddress -or ( $id -ne $ipv4GatewayAddress ) ) {
        $gatewayRoutes      = Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction 'Ignore'
        $ipv4GatewayAddress = findGatewayAddress $gatewayRoutes $networkConnectionProfile[0]
        if ( -not $ipv4GatewayAddress ) {
            if ( -not $allowDisconnect ) {
                $ipv4GatewayAddress = ""
            }
            else {
                $ipv4GatewayAddress = findGatewayAddressWithDisconnections $gatewayRoutes $networkConnectionProfile[0]
            }
        }
    }

    # find ipv6-gateway-address
    if ( -not $ipv6GatewayAddress -or ( $id -ne $ipv6GatewayAddress ) ) {
        $gatewayRoutes      = Get-NetRoute -DestinationPrefix '::/0' -ErrorAction 'Ignore'
        $ipv6GatewayAddress = findGatewayAddress $gatewayRoutes $networkConnectionProfile[0]
        if ( -not $ipv6GatewayAddress ) {
            if ( -not $allowDisconnect ) {
                $ipv6GatewayAddress = ""
            }
            else {
                $ipv6GatewayAddress = findGatewayAddressWithDisconnections $gatewayRoutes $networkConnectionProfile[0]
            }
        }
    }

    $ncProperties = @{
        GUID               = $guid
        IPv4GatewayAddress = $ipv4GatewayAddress
        IPv6GatewayAddress = $ipv6GatewayAddress
        Name               = $networkConnectionProfile[0].Name
        ConnectionProfile  = $networkConnectionProfile[0].NetworkCategory.ToString()
        IPv4Connectivity   = $networkConnectionProfile[0].IPv4Connectivity.ToString()
        IPv6Connectivity   = $networkConnectionProfile[0].IPv6Connectivity.ToString()
    }

    Write-Output $( ConvertTo-Json -InputObject $ncProperties -Depth 100 )
`)

//------------------------------------------------------------------------------

func updateNetworkConnection(c *WindowsClient, ncQuery *NetworkConnection, ncProperties *NetworkConnection) error {
    // find id
    id := ncQuery.GUID

    // convert query to JSON
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
    c.Lock.Lock()
    err = runner.Run(c, updateNetworkConnectionScript, updateNetworkConnectionArguments{
        NCQueryJSON:      string(ncQueryJSON),
        NCPropertiesJSON: string(ncPropertiesJSON),
    }, &stdout, &stderr)
    c.Lock.Unlock()
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

    $ncQuery = ConvertFrom-Json -InputObject '{{.NCQueryJSON}}'
    $guid = $ncQuery.GUID

    Get-NetConnectionProfile -ErrorAction 'Ignore' | foreach {
        $id = $guid
        $registryProfile = Get-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles\{$guid}" -ErrorAction Ignore
        if ( $registryProfile ) {
            $n = ( Get-ItemProperty -Path $registryProfile.PSPath -Name 'ProfileName' ).ProfileName
            if ( $n ) {
                $networkConnectionProfile = Get-NetConnectionProfile -Name $n -ErrorAction 'Ignore'
            }
        }
    }
    if ( -not $networkConnectionProfile ) {
        throw "cannot find network_connection '$id'"
    }

    $ncProperties = ConvertFrom-Json -InputObject '{{.NCPropertiesJSON}}'

    if ( ( $ncProperties.NewName -ne "" ) -and ( $ncProperties.NewName -ne $networkConnectionProfile.Name ) ) {
        Set-ItemProperty -Path $registryProfile.PSPath -Name 'ProfileName' -Value $ncProperties.NewName
    }

    if ( $ncProperties.ConnectionProfile -ne "" ) {
        Set-NetConnectionProfile -InputObject $networkConnectionProfile -NetworkCategory $ncProperties.ConnectionProfile -Confirm:$false | Out-Default
    }
`)

//------------------------------------------------------------------------------
