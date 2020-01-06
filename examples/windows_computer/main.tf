provider "windows" {
    version = "~> 0.0"

    type = "local"
}



data "windows_computer" "my_computer_A" {
}
output "my_computer_A_name" {
    value = data.windows_computer.my_computer_A.name
}
output "my_computer_A_reboot_pending" {
    value = data.windows_computer.my_computer_A.reboot_pending
}
output "my_computer_A_rename_pending" {
    value = data.windows_computer.my_computer_A.reboot_pending_details[0].computer_rename_pending
}



resource "windows_computer" "my_computer_1" {
    dns_client {
        suffix_search_list = [ "local" ]
        enable_devolution  = true
        devolution_level   = 0
    }
}
output "my_computer_1_name" {
    value = windows_computer.my_computer_1.name
}

# resource "windows_computer" "my_computer_2" {
#     new_name = "MY-COMPUTER"
# }
# output "my_computer_2_name" {
#     value = windows_computer.my_computer_2.name
# }
# output "my_computer_2_reboot_pending" {
#     value = windows_computer.my_computer_2.reboot_pending
# }
# output "my_computer_2_rename_pending" {
#     value = windows_computer.my_computer_2.reboot_pending_details[0].computer_rename_pending
# }
