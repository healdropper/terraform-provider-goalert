terraform {
  required_version = ">= 1.10"
  required_providers {
    goalert = {
      source = "healdropper/goalert"
    }
  }
}

provider "goalert" {}

variable "escalation_policy_id" {
  type        = string
  description = "UUID of an existing escalation policy on the selected GoAlert installation."
}

resource "goalert_service" "api" {
  name                 = "Example API"
  description          = "Managed by Terraform"
  escalation_policy_id = var.escalation_policy_id
}
