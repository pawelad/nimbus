resource "spacelift_stack" "nimbus" {
  autodeploy        = false
  branch            = "main"
  name              = "nimbus"
  project_root      = "src"
  repository        = "nimbus"
  terraform_version = "1.5.7"
}
