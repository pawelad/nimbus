# Cloudflare
variable "cloudflare_account_id" {
  type        = string
  description = "Cloudflare account ID."
  sensitive   = true
}

variable "cloudflare_api_token" {
  type        = string
  description = "Cloudflare API token."
  sensitive   = true
}

# Hetzner
variable "hcloud_token" {
  type        = string
  description = "Hetzner Cloud API token."
  sensitive   = true
}

variable "hcloud_ssh_key_name" {
  type        = string
  description = "Name of the SSH key on Hetzner Cloud."
}

variable "hcloud_ssh_public_key" {
  type        = string
  description = "SSH public key for Hetzner servers."
}
