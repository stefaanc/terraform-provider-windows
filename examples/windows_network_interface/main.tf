provider "windows" {
    version = "~> 0.0"

    type = "local"
}



data "windows_network_interface" "my_network_interface_A" {
    guid = "00A92352-A69E-4068-90FD-E0EA11BD5296"
}
output "my_network_interface_A_index" {
    value = data.windows_network_interface.my_network_interface_A.index
}
output "my_network_interface_A_network_connections" {
    value = data.windows_network_interface.my_network_interface_A.network_connection_names
}
output "my_network_interface_A_computer" {
    value = data.windows_network_interface.my_network_interface_A.computer_name
}

data "windows_network_interface" "my_network_interface_B" {
    index = 2
}
output "my_network_interface_B_index" {
    value = data.windows_network_interface.my_network_interface_B.index
}

data "windows_network_interface" "my_network_interface_C" {
    alias = "Ethernet"
}
output "my_network_interface_C_index" {
    value = data.windows_network_interface.my_network_interface_C.index
}

data "windows_network_interface" "my_network_interface_D" {
    description = "Intel(R) 82579LM Gigabit Network Connection"
}
output "my_network_interface_D_index" {
    value = data.windows_network_interface.my_network_interface_D.index
}

data "windows_network_interface" "my_network_interface_E" {
    mac_address = "00-15-5D-C5-83-07"
}
output "my_network_interface_E_index" {
    value = data.windows_network_interface.my_network_interface_E.index
}

data "windows_network_interface" "my_network_interface_F" {
    network_adapter_name = "Ethernet"
}
output "my_network_interface_F_index" {
    value = data.windows_network_interface.my_network_interface_F.index
}

data "windows_network_interface" "my_network_interface_G" {
    vnetwork_adapter_name = "External Switch"
}
output "my_network_interface_G_index" {
    value = data.windows_network_interface.my_network_interface_G.index
}
output "my_network_interface_G_vswitch" {
    value = data.windows_network_interface.my_network_interface_G.vswitch_name
}

data "windows_network_interface" "my_network_interface_H" {
    alias = "vEthernet (Default Switch) 2"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_interface_H_exists" {
    value = data.windows_network_interface.my_network_interface_H.x_lifecycle[0].exists
}
