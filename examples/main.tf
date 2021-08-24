
terraform {
  required_providers {
    sealedsecret = {
      version = "0.0.1"
      source  = "terraform.example.com/local/sealedsecret"
    }
  }
}

provider "sealedsecret" {
  controller_name      = "sealed-secrets-controller"
  controller_namespace = "kube-system"

  kubernetes {
    host                   = ""
    client_certificate     = base64decode("")
    client_key             = base64decode("")
    cluster_ca_certificate = base64decode("")
  }

  git {
    url      = ""
    username = ""
    token    = ""
  }
}

resource "sealedsecret_in_git" "example" {
  name      = "example-secret"
  namespace = "default"
  secrets = {
    "key" : "value"
  }
  filepath = "sealed-file.yaml"
}

