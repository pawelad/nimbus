resource "cloudflare_zone" "pawelad_me" {
  account_id = var.cloudflare_account_id
  zone       = "pawelad.me"
  plan       = "free"
}

resource "cloudflare_zone_settings_override" "pawelad_me" {
  zone_id = cloudflare_zone.pawelad_me.id

  settings {
    always_use_https         = "on"
    automatic_https_rewrites = "on"
    brotli                   = "on"
  }
}

# GitHub Pages
resource "cloudflare_record" "ghp_www" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "CNAME"
  name    = "www"
  value   = "pawelad.github.io"
  proxied = false
}

# https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site
resource "cloudflare_record" "ghp_apex" {
  for_each = toset(["185.199.108.153", "185.199.109.153", "185.199.110.153", "185.199.111.153"])

  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "@"
  value   = each.key
  proxied = false
}

# Nimbus
resource "cloudflare_record" "nimbus" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "nimbus"
  value   = digitalocean_droplet.nimbus.ipv4_address
  proxied = false
}

# Dokku
resource "cloudflare_record" "dokku_wildcart" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "*"
  value   = digitalocean_droplet.nimbus.ipv4_address
  proxied = false
}
