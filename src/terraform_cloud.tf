terraform {
  cloud {
    organization = "pawelad"

    workspaces {
      name = "nimbus"
    }
  }
}
