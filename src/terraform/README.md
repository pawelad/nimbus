# Terraform Infrastructure

This directory contains the Terraform configuration for the Nimbus homelab public infrastructure.

## Scope
It provisions and manages:
- **[Cloudflare]**: DNS records for `pawelad.me`, `pawelad.dev`, and `pipusznicy.cloud`.
- **[Hetzner Cloud]**: The `zapp` VPS instance.
- **[Spacelift]**: Configuration for the CI/CD pipeline that applies these changes.

## State Management (Spacelift)
State is **not** managed locally. We do not use a standard `backend.tf` with S3 or Terraform Cloud. 
Instead, Nimbus uses **[Spacelift]** as its infrastructure-as-code CI/CD platform.

Spacelift manages the remote state, provides state locking, and automatically applies changes when commits are merged to the `main` branch.

### Local Workflow
1. Make your changes to the `.tf` files.
2. Run `make tf-plan` (from the root directory) or `make plan` (from this directory) to format, validate, and generate a speculative execution plan locally.
   > [!NOTE]
   > Local speculative plans are run with `-lock=false` because Spacelift's remote state backend restricts CLI state locking to avoid conflicts with active CI/CD runs.
3. Commit and push your changes. Spacelift will automatically detect the changes, run a plan, and apply it upon merge.

### Emergency Local Apply
If Spacelift is completely down and you need to perform an emergency local apply, the remote backend (`spacelift.io`) will be unreachable. Because the backend is configured using Spacelift's remote API, standard `terraform` commands will fail.

Follow these steps to temporarily bypass the remote backend and apply changes locally:

1. **Disable the remote backend**:
   Rename or move the `spacelift_override.tf` file so Terraform defaults to using the `local` backend:
   ```bash
   mv spacelift_override.tf spacelift_override.tf.bak
   ```
2. **Obtain the state file**:
   Locate the latest state file (from a backup or previous run cache) and save it locally as `terraform.tfstate` in this directory.
3. **Reinitialize Terraform**:
   Initialize Terraform using the local backend:
   ```bash
   terraform init
   ```
4. **Apply changes**:
   Apply your infrastructure changes locally:
   ```bash
   terraform apply
   ```
5. **Restore Spacelift integration**:
   Once Spacelift is back online, you must migrate the updated local state back to the Spacelift backend:
   ```bash
   mv spacelift_override.tf.bak spacelift_override.tf
   terraform init -migrate-state
   ```

[cloudflare]: https://www.cloudflare.com/
[hetzner cloud]: https://www.hetzner.com/cloud/
[spacelift]: https://spacelift.io/
