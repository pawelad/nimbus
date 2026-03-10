### pawelad.me ###
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
    ssl                      = "full"
  }
}

# GitHub domain verification
resource "cloudflare_record" "pawelad_me_github_verification" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "TXT"
  name    = "_github-pages-challenge-pawelad"
  value   = "9e3e75692c0313c903f1a30177555c"
  proxied = false
}

# GitHub Pages
resource "cloudflare_record" "ghp_www" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "CNAME"
  name    = "www"
  value   = "pawelad.github.io"
  proxied = true
}

# https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site
resource "cloudflare_record" "ghp_apex" {
  for_each = toset(["185.199.108.153", "185.199.109.153", "185.199.110.153", "185.199.111.153"])

  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "@"
  value   = each.key
  proxied = true
}

# Zapp
resource "cloudflare_record" "zapp" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "zapp"
  value   = hcloud_server.zapp.ipv4_address
  proxied = false
}

# Wildcard — routes all subdomains to Dokploy / Traefik
resource "cloudflare_record" "zapp_wildcard" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "*"
  value   = hcloud_server.zapp.ipv4_address
  proxied = false
}

moved {
  from = cloudflare_record.dokku_wildcard
  to   = cloudflare_record.zapp_wildcard
}

# Nimbus
resource "cloudflare_record" "nimbus" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "nimbus"
  value   = digitalocean_droplet.nimbus.ipv4_address
  proxied = false
}

# fakester (pre migration)
resource "cloudflare_record" "fakester" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "fakester"
  value   = digitalocean_droplet.nimbus.ipv4_address
  proxied = false
}

### pawelad.dev ###
resource "cloudflare_zone" "pawelad_dev" {
  account_id = var.cloudflare_account_id
  zone       = "pawelad.dev"
  plan       = "free"
}

resource "cloudflare_zone_settings_override" "pawelad_dev" {
  zone_id = cloudflare_zone.pawelad_dev.id

  settings {
    always_use_https         = "on"
    automatic_https_rewrites = "on"
    brotli                   = "on"
    ssl                      = "full"
  }
}

# GitHub domain verification
resource "cloudflare_record" "pawelad_dev_github_verification" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "TXT"
  name    = "_github-pages-challenge-pawelad"
  value   = "038a851ef9fc64d575187ca20e59d3"
  proxied = false
}

# fakester
resource "cloudflare_record" "ghp_fakester" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "fakester"
  value   = "pawelad.github.io"
  proxied = true
}

# monz
resource "cloudflare_record" "rtd_monz" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "monz"
  value   = "readthedocs.io"
  proxied = false
}

# pymonzo
resource "cloudflare_record" "rtd_pymonzo" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "pymonzo"
  value   = "readthedocs.io"
  proxied = false
}
