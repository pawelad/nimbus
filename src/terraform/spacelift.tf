terraform {
  backend "remote" {
    hostname     = "spacelift.io"
    organization = "pawelad"

    workspaces {
      name = "nimbus"
    }
  }
}