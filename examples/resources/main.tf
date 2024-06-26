terraform {
  required_providers {
    sealedsecret = {
      version = ">=1.2.0"
      source  = "jifwin/terraform-provider-sealedsecret"
    }
  }
}

provider "sealedsecret" {
  controller_name      = "sealed-secret-controller-sealed-secrets"
  controller_namespace = "kube-system"

  kubernetes {
    host                   = var.k8s_host
    client_certificate     = base64decode(var.k8s_client_certificate)
    client_key             = base64decode(var.k8s_client_key)
    cluster_ca_certificate = base64decode(var.k8s_cluster_ca_certificate)
  }
}

resource "sealedsecret" "example" {
  name      = "example-secret"
  namespace = "default"
  data      = {
    "key" : "value"
  }
}

resource "local_file" "example" {
  filename = "sealedsecret.yaml"
  content  = sealedsecret.example.yaml_content
}

variable "k8s_client_certificate" {
  type = string
}

variable "k8s_client_key" {
  type = string
}

variable "k8s_cluster_ca_certificate" {
  type = string
}
variable "k8s_host" {
  type = string
}