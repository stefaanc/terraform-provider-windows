provider "windows" {
    version = "~> 0.0"

    type = "local"
}



data "windows_network_connection" "my_network_connection_A" {
    guid = "C61471E6-63BF-4944-B625-94EEFA1833B2"
}
output "my_network_connection_A_name" {
    value = data.windows_network_connection.my_network_connection_A.name
}

data "windows_network_connection" "my_network_connection_B" {
    ipv4_gateway_address = "192.168.2.1"
    allow_disconnect     = true
}
output "my_network_connection_B_name" {
    value = data.windows_network_connection.my_network_connection_B.name
}
output "my_network_connection_B_connectivity" {
    value = data.windows_network_connection.my_network_connection_B.ipv4_connectivity
}

data "windows_network_connection" "my_network_connection_C" {
    ipv6_gateway_address = "fe80::69c2:41ab:c6a1:1"
    allow_disconnect     = true
}
output "my_network_connection_C_name" {
    value = data.windows_network_connection.my_network_connection_C.name
}
output "my_network_connection_C_connectivity" {
    value = data.windows_network_connection.my_network_connection_C.ipv6_connectivity
}

data "windows_network_connection" "my_network_connection_D" {
    name             = "Network"
    allow_disconnect = true
}
output "my_network_connection_D_guid" {
    value = data.windows_network_connection.my_network_connection_D.guid
}
output "my_network_connection_D_ipv4_gateway" {
    value = data.windows_network_connection.my_network_connection_D.ipv4_gateway_address
}
output "my_network_connection_D_ipv6_gateway" {
    value = data.windows_network_connection.my_network_connection_D.ipv6_gateway_address
}

data "windows_network_connection" "my_network_connection_E" {
    name = "Network 2"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_connection_E_exists" {
    value = data.windows_network_connection.my_network_connection_E.x_lifecycle[0].exists
}



# resource "windows_network_connection" "my_network_connection_1" {
#     guid = "8598CC0D-18B8-4364-80D6-2F75C37ABCC0"

#     connection_profile = "private"
# }
# output "my_network_connection_1_name" {
#     value = windows_network_connection.my_network_connection_1.name
# }

# resource "windows_network_connection" "my_network_connection_2" {
#     ipv4_gateway_address = "192.168.2.1"
#     allow_disconnect     = true
# }
# output "my_network_connection_2_name" {
#     value = windows_network_connection.my_network_connection_2.name
# }
# output "my_network_connection_2_connectivity" {
#     value = windows_network_connection.my_network_connection_2.ipv4_connectivity
# }

# resource "windows_network_connection" "my_network_connection_3" {
#     ipv6_gateway_address = "fe80::69c2:41ab:c6a1:1"
#     allow_disconnect     = true
# }
# output "my_network_connection_3_name" {
#     value = windows_network_connection.my_network_connection_3.name
# }
# output "my_network_connection_3_connectivity" {
#     value = windows_network_connection.my_network_connection_3.ipv6_connectivity
# }

resource "windows_network_connection" "my_network_connection_4" {
    name             = "Network"
    allow_disconnect = true
}
output "my_network_connection_4_guid" {
    value = windows_network_connection.my_network_connection_4.guid
}
output "my_network_connection_4_ipv4_gateway" {
    value = windows_network_connection.my_network_connection_4.ipv4_gateway_address
}
output "my_network_connection_4_ipv6_gateway" {
    value = windows_network_connection.my_network_connection_4.ipv6_gateway_address
}

# resource "windows_network_connection" "my_network_connection_5" {
#     old_name = "Network"
#     new_name = "My-Network"
# }
# output "my_network_connection_5_name" {
#     value = windows_network_connection.my_network_connection_5.name
# }

# resource "windows_network_connection" "my_network_connection_6" {
#     old_name = "Network 2"
#     new_name = "My-Backup-Network"

#     x_lifecycle {
#         ignore_error_if_not_exists = true
#     }
# }
# output "my_network_connection_6_exists" {
#     value = windows_network_connection.my_network_connection_6.x_lifecycle[0].exists
# }
