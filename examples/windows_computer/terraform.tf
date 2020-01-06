terraform {
    required_version = ">= 0.12.17"

    backend "local" {
        path = "./_terraform.tfstate"
    }
}