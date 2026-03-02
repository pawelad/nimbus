output "nimbus_ip_address" {
  value       = digitalocean_droplet.nimbus.ipv4_address
  description = "Nimbus Droplet public IP address."
}

output "pawelad_me_zone_id" {
  value       = cloudflare_zone.pawelad_me.id
  description = "Cloudflare 'pawelad.me' zone ID."
}

output "pawelad_dev_zone_id" {
  value       = cloudflare_zone.pawelad_dev.id
  description = "Cloudflare 'pawelad.dev' zone ID."
}
