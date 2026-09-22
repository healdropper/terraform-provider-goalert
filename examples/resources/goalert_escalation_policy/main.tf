terraform {
  required_version = ">= 1.10"
  required_providers {
    goalert = {
      source = "healdropper/goalert"
    }
  }
}

provider "goalert" {}

resource "goalert_escalation_policy" "production" {
  name        = "Production High Priority"
  description = "Escalates critical alerts to on-call webhooks"
  repeat      = 3

  step {
    delay_minutes = 5
    webhook_action {
      url = "https://events.example.com/alerts/primary"
    }
  }

  step {
    delay_minutes = 10
    webhook_action {
      url = "https://events.example.com/alerts/secondary"
    }
  }
}
