data "digitalocean_ssh_key" "root" {
  name = "pawelad@digitalocean"
}

data "digitalocean_ssh_key" "dokku" {
  name = "dokku@digitalocean"
}

resource "digitalocean_droplet" "nimbus" {
  image         = "ubuntu-22-04-x64"
  name          = "nimbus"
  region        = "fra1"
  size          = "s-1vcpu-2gb"
  monitoring    = true
  droplet_agent = true
  ssh_keys      = [data.digitalocean_ssh_key.root.id]
  tags          = ["terraform"]

  lifecycle {
    ignore_changes = [
      user_data,
    ]
  }
}
