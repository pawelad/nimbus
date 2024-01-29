resource "spacelift_stack" "nimbus" {
  autodeploy        = false
  branch            = "main"
  name              = "nimbus"
  project_root      = "src"
  repository        = "nimbus"
  terraform_version = "1.5.7"

  # 8< --------------------------------------------------------------
  # Delete the following line after the stack has been created
  import_state_file = "/mnt/workspace/state-import/nimbus.tfstate"
  # -------------------------------------------------------------- 8<
}
