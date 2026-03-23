data "hcloud_ssh_key" "zapp" {
  name = var.hcloud_ssh_key_name
}

resource "hcloud_server" "zapp" {
  name        = "zapp"
  image       = "ubuntu-24.04"
  server_type = "cx23" # 2 vCPU, 4GB RAM (Intel x86)
  location    = "nbg1" # Nuremberg
  ssh_keys    = [data.hcloud_ssh_key.zapp.id]

  labels = {
    provisioner = "terraform"
  }

  user_data = templatefile("${path.module}/templates/cloud-config.yaml", {
    username            = "pawel"
    user_ssh_public_key = chomp(data.hcloud_ssh_key.zapp.public_key)
  })

  lifecycle {
    ignore_changes = [
      ssh_keys,  # Ignore if we manually add/remove keys later
      user_data, # Though empty, good practice if we ever add it
    ]
  }
}
