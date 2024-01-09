# nimbus
My personal [Terraform] stack, deployed with Terraform Cloud.

It consists of:
- [DigitalOcean]
  - `nimbus` droplet with [dokku] installed
- [Cloudflare]
  - `pawelad.me` zone
  - `pawelad.dev` zone
  - DNS records for [GitHub Pages]
  - DNS records for [dokku]

## Makefile
Available `make` commands:

```console
$ make help  
plan                                      Generate a (speculative) Terraform plan
apply                                     Generate, confirm and apply a Terraform plan
upgrade                                   Upgrade Terraform providers
destroy                                   Destroy infrastructure managed by Terraform
format                                    Format Terraform files
help                                      Show help message
```

## Authors
Developed and maintained by [Paweł Adamczak][pawelad].

Source code is available at [GitHub][github nimbus].

Released under [Mozilla Public License 2.0][license].


[cloudflare]: https://www.cloudflare.com/
[digitalocean]: https://www.digitalocean.com/
[dokku]: https://dokku.com/
[github nimbus]: https://github.com/pawelad/nimbus
[github pages]: https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site
[license]: ./LICENSE
[pawelad]: https://pawelad.me/
[terraform]: https://www.terraform.io/
