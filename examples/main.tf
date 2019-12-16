provider "windows" {
    version = "~> 0.0"

    type = "local"
}

# data "hyperv_link_ip_interface" "test" {
#     # index = 7
#     # name = "ethernet_32776"
#     # alias = "vEthernet (External Switch)"
#     # description = "Hyper-V Virtual Ethernet Adapter #2"
#     # mac_address = "00-10-18-57-1b-0e"
#     # network_adapter_name = "vEthernet (External Switch)"
#     vnetwork_adapter_name = "External Switch"
# }

resource "windows_computer" "my_computer" {
    new_name = "MY-COMPUTER"

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
    value = windows_computer.my_computer.reboot_pending_details #[*].computer_rename_pending
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
