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
  content = "9e3e75692c0313c903f1a30177555c"
  proxied = false
}

# GitHub Pages
resource "cloudflare_record" "ghp_www" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "CNAME"
  name    = "www"
  content = "pawelad.github.io"
  proxied = true
}

# https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site
resource "cloudflare_record" "ghp_apex" {
  for_each = toset(["185.199.108.153", "185.199.109.153", "185.199.110.153", "185.199.111.153"])

  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "@"
  content = each.key
  proxied = true
}

# Zapp
resource "cloudflare_record" "zapp" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "zapp"
  content = hcloud_server.zapp.ipv4_address
  proxied = false
}

# Wildcard — routes all subdomains to Dokploy / Traefik
resource "cloudflare_record" "zapp_wildcard" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "*"
  content = hcloud_server.zapp.ipv4_address
  proxied = false
}

moved {
  from = cloudflare_record.dokku_wildcard
  to   = cloudflare_record.zapp_wildcard
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
  content = "038a851ef9fc64d575187ca20e59d3"
  proxied = false
}

# fakester
resource "cloudflare_record" "ghp_fakester" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "fakester"
  content = "pawelad.github.io"
  proxied = true
}

# monz
resource "cloudflare_record" "rtd_monz" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "monz"
  content = "readthedocs.io"
  proxied = false
}

# pymonzo
resource "cloudflare_record" "rtd_pymonzo" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "pymonzo"
  content = "readthedocs.io"
  proxied = false
}

### pipusznicy.cloud ###
resource "cloudflare_zone" "pipusznicy_cloud" {
  account_id = var.cloudflare_account_id
  zone       = "pipusznicy.cloud"
  plan       = "free"
}

resource "cloudflare_zone_settings_override" "pipusznicy_cloud" {
  zone_id = cloudflare_zone.pipusznicy_cloud.id

  settings {
    always_use_https         = "on"
    automatic_https_rewrites = "on"
    brotli                   = "on"
    ssl                      = "full"
  }
}

# Zapp
resource "cloudflare_record" "zapp_pipusznicy" {
  zone_id = cloudflare_zone.pipusznicy_cloud.id
  type    = "A"
  name    = "zapp"
  content = hcloud_server.zapp.ipv4_address
  proxied = false
}
