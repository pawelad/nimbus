output "zapp_ip_address" {
  value       = hcloud_server.zapp.ipv4_address
  description = "New Zapp server public IP address."
}

output "pawelad_me_zone_id" {
  value       = cloudflare_zone.pawelad_me.id
  description = "Cloudflare 'pawelad.me' zone ID."
}

output "pawelad_dev_zone_id" {
  value       = cloudflare_zone.pawelad_dev.id
  description = "Cloudflare 'pawelad.dev' zone ID."
}

output "pipusznicy_cloud_zone_id" {
  value       = cloudflare_zone.pipusznicy_cloud.id
  description = "Cloudflare 'pipusznicy.cloud' zone ID."
}
