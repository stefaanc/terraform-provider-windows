provider "windows" {
    version = "~> 0.0"

    type = "local"
}



data "windows_network_adapter" "my_network_adapter_A" {
    guid = "C42B1E6D-0856-4932-B06C-3085DA1B1978"
}
output "my_network_adapter_A_name" {
    value = data.windows_network_adapter.my_network_adapter_A.name
}
output "my_network_adapter_A_mac_address" {
    value = data.windows_network_adapter.my_network_adapter_A.mac_address
}
output "my_network_adapter_A_admin_status" {
    value = data.windows_network_adapter.my_network_adapter_A.admin_status
}
output "my_network_adapter_A_operational_status" {
    value = data.windows_network_adapter.my_network_adapter_A.operational_status
}
output "my_network_adapter_A_connection_status" {
    value = data.windows_network_adapter.my_network_adapter_A.connection_status
}
output "my_network_adapter_A_connection_speed" {
    value = data.windows_network_adapter.my_network_adapter_A.connection_speed
}
output "my_network_adapter_A_is_physical" {
    value = data.windows_network_adapter.my_network_adapter_A.is_physical
}

data "windows_network_adapter" "my_network_adapter_B" {
    name = "Ethernet"
}
output "my_network_adapter_B_guid" {
    value = data.windows_network_adapter.my_network_adapter_B.guid
}

data "windows_network_adapter" "my_network_adapter_C" {
    name = "Ethernet"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_adapter_C_exists" {
    value = data.windows_network_adapter.my_network_adapter_C.x_lifecycle[0].exists
}



resource "windows_network_adapter" "my_network_adapter_1" {
    guid = "00A92352-A69E-4068-90FD-E0EA11BD5296"

    mac_address = "00-10-18-57-1B-0E"

    dns_client {
        register_connection_address = true
        register_connection_suffix  = "staging.local"
    }
}
output "my_network_adapter_1_name" {
    value = windows_network_adapter.my_network_adapter_1.name
}
output "my_network_adapter_1_admin_status" {
    value = windows_network_adapter.my_network_adapter_1.admin_status
}
output "my_network_adapter_1_operational_status" {
    value = windows_network_adapter.my_network_adapter_1.operational_status
}
output "my_network_adapter_1_connection_status" {
    value = windows_network_adapter.my_network_adapter_1.connection_status
}
output "my_network_adapter_1_connection_speed" {
    value = windows_network_adapter.my_network_adapter_1.connection_speed
}
output "my_network_adapter_1_is_physical" {
    value = windows_network_adapter.my_network_adapter_1.is_physical
}

# resource "windows_network_adapter" "my_network_adapter_2" {
#     name = "Ethernet"
# }
# output "my_network_adapter_2_guid" {
#     value = windows_network_adapter.my_network_adapter_2.guid
# }

# resource "windows_network_adapter" "my_network_adapter_3" {
#     old_name = "Ethernet"
#     new_name = "Staging"
# }
# output "my_network_adapter_3_name" {
#     value = windows_network_adapter.my_network_adapter_3.name
# }

# resource "windows_network_adapter" "my_network_adapter_4" {
#     name = "Staging"

#     x_lifecycle {
#         ignore_error_if_not_exists = true
#     }
# }
# output "my_network_adapter_4_exists" {
#     value = windows_network_adapter.my_network_adapter_4.x_lifecycle[0].exists
# }
