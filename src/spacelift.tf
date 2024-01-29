resource "spacelift_stack" "nimbus" {
  autodeploy        = false
  branch            = "main"
  name              = "nimbus"
  project_root      = "src"
  repository        = "nimbus"
  terraform_version = "1.5.7"
}

resource "spacelift_environment_variable" "nimbus_tf_var_droplet_username" {
  stack_id   = spacelift_stack.nimbus.id
  name       = "TF_VAR_droplet_username"
  value      = "pawelad"
  write_only = false
}

resource "spacelift_environment_variable" "nimbus_tf_var_dokku_domain" {
  stack_id   = spacelift_stack.nimbus.id
  name       = "TF_VAR_dokku_domain"
  value      = "pawelad.me"
  write_only = false
}
