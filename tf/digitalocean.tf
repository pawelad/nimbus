data "digitalocean_ssh_key" "key" {
  name = var.digitalocean_ssh_key_name
}

resource "digitalocean_droplet" "nimbus" {
  image         = "ubuntu-22-04-x64"
  name          = "nimbus"
  region        = "fra1"
  size          = "s-1vcpu-2gb"
  monitoring    = true
  droplet_agent = true
  ssh_keys      = [data.digitalocean_ssh_key.key.id]
  tags          = ["terraform"]

  user_data = templatefile("${path.module}/templates/cloud-config.yaml", {
    username       = var.droplet_username
    dokku_domain   = var.dokku_domain
    ssh_public_key = data.digitalocean_ssh_key.key.public_key
  })
}
