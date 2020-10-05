# Tinkerbell Terraform Provider

![](https://img.shields.io/badge/Stability-Experimental-red.svg)

This repository is [Experimental](https://github.com/packethost/standards/blob/master/experimental-statement.md) meaning that it's based on untested ideas or techniques and not yet established or finalized or involves a radically new and innovative style! This means that support is best effort (at best!) and we strongly encourage you to NOT use this in production.

The Tinkerbell provider allows to create [Tinkerbell](https://tinkerbell.org/) hardware entried, templates and workflows in a declarative way.

## Table of contents
* [User documentation](#user-documentation)
* [Building and testing](#building-and-testing)
* [Releasing](#releasing)
* [Authors](#authors)

## User documentation

For user documentation, see [Terraform Registry](https://registry.terraform.io/providers/tinkerbell/tinkerbell/latest/docs).

## Building and testing

For local builds, run `make` which will build the binary, run unit tests and linter.

## Releasing

This project use `goreleaser` for releasing. To release new version, follow the following steps:

* Add a changelog for new release to CHANGELOG.md file.

* Tag new release on desired git, using example command:

  ```sh
  git tag -a v0.4.7 -s -m "Release v0.4.7"
  ```

* Push the tag to GitHub
  ```sh
  git push origin v0.4.7
  ```

* Run `make release` to create a GitHub Release:
  ```sh
  GITHUB_TOKEN=githubtoken GPG_FINGERPRINT=gpgfingerprint make release
  ```

* Go to newly create [GitHub release](https://github.com/tinkerbell/terraform-provider-tinkerbell/releases/tag/v0.4.7), verify that the changelog and artefacts looks correct and publish it.

## Authors

* **Mateusz Gozdek** - *Initial work* - [invidian](https://github.com/invidian)
