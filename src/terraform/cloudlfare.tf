### pawelad.me ###
resource "cloudflare_zone" "pawelad_me" {
  account = { id = var.cloudflare_account_id }
  name    = "pawelad.me"
}

# GitHub domain verification
resource "cloudflare_dns_record" "pawelad_me_github_verification" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "TXT"
  name    = "_github-pages-challenge-pawelad"
  content = "9e3e75692c0313c903f1a30177555c"
  ttl     = 1
  proxied = false
}

# GitHub Pages
resource "cloudflare_dns_record" "ghp_www" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "CNAME"
  name    = "www"
  content = "pawelad.github.io"
  ttl     = 1
  proxied = true
}

# https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site
resource "cloudflare_dns_record" "ghp_apex" {
  for_each = toset(["185.199.108.153", "185.199.109.153", "185.199.110.153", "185.199.111.153"])

  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "@"
  content = each.key
  ttl     = 1
  proxied = true
}

# Zapp
resource "cloudflare_dns_record" "zapp" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "zapp"
  content = hcloud_server.zapp.ipv4_address
  ttl     = 1
  proxied = false
}

# Wildcard — routes all subdomains to Dokploy / Traefik
resource "cloudflare_dns_record" "zapp_wildcard" {
  zone_id = cloudflare_zone.pawelad_me.id
  type    = "A"
  name    = "*"
  content = hcloud_server.zapp.ipv4_address
  ttl     = 1
  proxied = false
}

### pawelad.dev ###
resource "cloudflare_zone" "pawelad_dev" {
  account = { id = var.cloudflare_account_id }
  name    = "pawelad.dev"
}

# GitHub domain verification
resource "cloudflare_dns_record" "pawelad_dev_github_verification" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "TXT"
  name    = "_github-pages-challenge-pawelad"
  content = "038a851ef9fc64d575187ca20e59d3"
  ttl     = 1
  proxied = false
}

# fakester
resource "cloudflare_dns_record" "ghp_fakester" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "fakester"
  content = "pawelad.github.io"
  ttl     = 1
  proxied = true
}

# monz
resource "cloudflare_dns_record" "rtd_monz" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "monz"
  content = "readthedocs.io"
  ttl     = 1
  proxied = false
}

# pymonzo
resource "cloudflare_dns_record" "rtd_pymonzo" {
  zone_id = cloudflare_zone.pawelad_dev.id
  type    = "CNAME"
  name    = "pymonzo"
  content = "readthedocs.io"
  ttl     = 1
  proxied = false
}

### pipusznicy.cloud ###
resource "cloudflare_zone" "pipusznicy_cloud" {
  account = { id = var.cloudflare_account_id }
  name    = "pipusznicy.cloud"
}

# Zapp
resource "cloudflare_dns_record" "zapp_pipusznicy" {
  zone_id = cloudflare_zone.pipusznicy_cloud.id
  type    = "A"
  name    = "zapp"
  content = hcloud_server.zapp.ipv4_address
  ttl     = 1
  proxied = false
}
