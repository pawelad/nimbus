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

variable "digitalocean_ssh_key_name" {
  type        = string
  description = "DigitalOcean SSH key name."
}

variable "droplet_username" {
  type        = string
  description = "Local Droplet username."
}

# Dokku
variable "dokku_domain" {
  type        = string
  description = "Global Dokku domain."
}
