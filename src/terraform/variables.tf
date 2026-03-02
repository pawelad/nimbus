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

# DigitalOcean
variable "digitalocean_api_token" {
  type        = string
  description = "DigitalOcean API token."
  sensitive   = true
}

variable "droplet_username" {
  type        = string
  description = "Local Droplet username."
}

# Hetzner
variable "hcloud_token" {
  type        = string
  description = "Hetzner Cloud API token."
  sensitive   = true
}

# Dokku
variable "dokku_domain" {
  type        = string
  description = "Global Dokku domain."
}
