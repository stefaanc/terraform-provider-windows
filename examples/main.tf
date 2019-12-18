provider "windows" {
    version = "~> 0.0"

    type = "local"
}



data "windows_link_ip_interface" "my_ip_interface_1" {
    index = 16
}
output "my_ip_interface_1_index" {
    value = data.windows_link_ip_interface.my_ip_interface_1.index
}
output "my_ip_interface_1_network_connections" {
    value = data.windows_link_ip_interface.my_ip_interface_1.network_connection_names
}
output "my_ip_interface_1_computer" {
    value = data.windows_link_ip_interface.my_ip_interface_1.computer_name
}

data "windows_link_ip_interface" "my_ip_interface_2" {
    alias = "Ethernet"
}
output "my_ip_interface_2_index" {
    value = data.windows_link_ip_interface.my_ip_interface_2.index
}

data "windows_link_ip_interface" "my_ip_interface_3" {
    description = "Intel(R) 82579LM Gigabit Network Connection"
}
output "my_ip_interface_3_index" {
    value = data.windows_link_ip_interface.my_ip_interface_3.index
}

data "windows_link_ip_interface" "my_ip_interface_4" {
    guid = "E34DC156-C49F-42DB-A6F6-D5609648D274"
}
output "my_ip_interface_4_index" {
    value = data.windows_link_ip_interface.my_ip_interface_4.index
}

data "windows_link_ip_interface" "my_ip_interface_5" {
    mac_address = "90-B1-1C-63-D3-82"
}
output "my_ip_interface_5_index" {
    value = data.windows_link_ip_interface.my_ip_interface_5.index
}

data "windows_link_ip_interface" "my_ip_interface_6" {
    network_adapter_name = "Ethernet"
}
output "my_ip_interface_6_index" {
    value = data.windows_link_ip_interface.my_ip_interface_6.index
}

data "windows_link_ip_interface" "my_ip_interface_7" {
    vnetwork_adapter_name = "External Switch"
}
output "my_ip_interface_7_index" {
    value = data.windows_link_ip_interface.my_ip_interface_7.index
}
output "my_ip_interface_7_vswitch" {
    value = data.windows_link_ip_interface.my_ip_interface_7.vswitch_name
}

data "windows_link_ip_interface" "my_ip_interface_8" {
    vnetwork_adapter_name = "Default Switch"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_ip_interface_8_index" {
    value = data.windows_link_ip_interface.my_ip_interface_8.index
}
output "my_ip_interface_8_exists" {
    value = data.windows_link_ip_interface.my_ip_interface_8.x_lifecycle[0].exists
}



resource "windows_computer" "my_computer" {
    new_name = "MY-COMPUTER"
    #new_name = "HOLMECROFT-H2"

    dns_client {
        suffix_search_list = [ "local" ]
        enable_devolution  = true
        devolution_level   = 0
    }
}
output "my_computer_name" {
    value = windows_computer.my_computer.name
}
output "my_computer_reboot_pending" {
    value = windows_computer.my_computer.reboot_pending
}
output "my_computer_rename_pending" {
    value = windows_computer.my_computer.reboot_pending_details[0].computer_rename_pending
}



# resource "null" {
#     provisioner "remote-exec" {
#         inline = [ "reboot" ]
#         on_failure = "continue"
#         connection { host = self.ipv4_address }
#     }
# }

# resource "hyperv_network" "updated" {
#     depends_on = [
#         hyperv_network_adapter.updated,
#     ]

#     name               = "network"
#     connection_profile = "public"
# }

# resource "hyperv_network_adapter" "updated" {
#     name = "vEthernet (External Switch)"
#     mac_address = "00-10-18-57-1b-0e"
#     ipv4 {
# #         address_family = "ipv4"
# #        interface_disabled = false
# #        interface_metric = 8
# #        dns = [ "8.8.8.8", "8.8.4.4" ]
#     }
#     ipv6 {
#         interface_disabled = true
#         interface_metric = 8
#         dns = [ "2001:4860:4860::8888", "2001:4860:4860::8844" ]
#     }
# #    register_connection_address = false
#     register_connection_suffix = ""
# }
