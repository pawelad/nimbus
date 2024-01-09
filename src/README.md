# nimbus
Simple [Terraform] stack that consists of:
- [DigitalOcean]
  - `nimbus` droplet with [dokku] installed
- [Cloudflare]
  - `pawelad.me` zone
  - `pawelad.dev` zone
  - DNS records for [GitHub Pages]
  - DNS records for [dokku]


[cloudflare]: https://www.cloudflare.com/
[digitalocean]: https://www.digitalocean.com/
[dokku]: https://dokku.com/
[github pages]: https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site
[terraform]: https://www.terraform.io/
