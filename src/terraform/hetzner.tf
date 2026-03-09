resource "hcloud_ssh_key" "zapp" {
  name       = "pawelad@zapp"
  public_key = var.hcloud_ssh_public_key
}

resource "hcloud_server" "zapp" {
  name        = "zapp"
  image       = "ubuntu-24.04"
  server_type = "cx23" # 2 vCPU, 4GB RAM (Intel x86)
  location    = "nbg1" # Nuremberg
  ssh_keys    = [hcloud_ssh_key.zapp.id]

  labels = {
    provisioner = "terraform"
  }

  user_data = templatefile("${path.module}/templates/cloud-config.yaml", {
    username            = "pawel"
    user_ssh_public_key = chomp(var.hcloud_ssh_public_key)
  })

  lifecycle {
    ignore_changes = [
      ssh_keys,  # Ignore if we manually add/remove keys later
      user_data, # Though empty, good practice if we ever add it
    ]
  }
}
