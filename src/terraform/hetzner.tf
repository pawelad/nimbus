resource "hcloud_ssh_key" "zapp" {
  name       = "pawelad@zapp"
  public_key = file("~/.ssh/hetzner_ed25519.pub")
}

resource "hcloud_server" "zapp" {
  name        = "zapp"
  image       = "ubuntu-24.04"
  server_type = "cx22" # 2 vCPU, 4GB RAM (Intel x86)
  location    = "fsn1" # Falkenstein
  ssh_keys    = [hcloud_ssh_key.zapp.id]

  labels = {
    provisioner = "terraform"
  }

  lifecycle {
    ignore_changes = [
      ssh_keys,  # Ignore if we manually add/remove keys later
      user_data, # Though empty, good practice if we ever add it
    ]
  }
}
