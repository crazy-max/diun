# regclient

[![Go Workflow Status](https://img.shields.io/github/actions/workflow/status/regclient/regclient/go.yml?branch=main&label=Go%20build)](https://github.com/regclient/regclient/actions/workflows/go.yml)
[![Docker Workflow Status](https://img.shields.io/github/actions/workflow/status/regclient/regclient/docker.yml?branch=main&label=Docker%20build)](https://github.com/regclient/regclient/actions/workflows/docker.yml)
[![Dependency Workflow Status](https://img.shields.io/github/actions/workflow/status/regclient/regclient/version-check.yml?branch=main&label=Dependency%20check)](https://github.com/regclient/regclient/actions/workflows/version-check.yml)
[![Vulnerability Workflow Status](https://img.shields.io/github/actions/workflow/status/regclient/regclient/vulnscans.yml?branch=main&label=Vulnerability%20check)](https://github.com/regclient/regclient/actions/workflows/vulnscans.yml)

[![Go Reference](https://pkg.go.dev/badge/github.com/regclient/regclient.svg)](https://pkg.go.dev/github.com/regclient/regclient)
![License](https://img.shields.io/github/license/regclient/regclient)
[![Go Report Card](https://goreportcard.com/badge/github.com/regclient/regclient)](https://goreportcard.com/report/github.com/regclient/regclient)
[![GitHub Downloads](https://img.shields.io/github/downloads/regclient/regclient/total?label=GitHub%20downloads)](https://github.com/regclient/regclient/releases)

regclient is a client interface to OCI conformant registries and content shipped with the OCI Image Layout.
It includes a Go library and several CLI commands.

## regclient Go Library Features

- Runs without a container runtime and without privileged access to the local host.
- Querying for a tag listing, repository listing, and remotely inspecting the contents of images.
- Efficiently copying and retagging images, only pulling layers when required, and without changing the image digest.
- Support for multi-platform images.
- Support for querying, creating, and copying OCI Artifacts, allowing arbitrary data to be stored in an OCI registry.
- Support for packaging OCI Artifacts with an Index of multiple artifacts, which can be used for platform specific artifacts.
- Support for querying OCI referrers, copying referrers, and pushing content with an OCI subject field, associating artifacts with other content on the registry.
- Support for the “digest tags” used by projects like sigstore/cosign, allowing the content to be included when copying images.
- Efficiently query for an image digest.
- Efficiently query for pull rate limits used by Docker Hub.
- Import and export content into OCI Layouts and Docker formatted tar files.
- Support OCI Layouts in all commands as a local disk equivalent of a repository.
- Support for deleting tags, manifests, and blobs.
- Ability to mutate existing images, including:
  - Settings annotations or labels
  - Deleting content from layers
  - Changing timestamps for reproducibility
  - Converting between Docker and OCI media types
  - Replacing the base image layers
  - Add or remove volumes and exposed ports
  - Change digest algorithms
- Support for registry warning headers, which may be used to notify users of issues with the server or content they are using.
- Automatically import logins from the docker CLI, and registry certificates from the docker engine.
- Automatic retry, and fallback to a chunked blob push, when network issues are encountered.

The full Go references is available on [pkg.go.dev](https://pkg.go.dev/github.com/regclient/regclient).

## regctl Features

`regctl` is a CLI interface to the `regclient` library.
In addition to the features listed for `regclient`, `regctl` adds the following abilities:

- Generating multi-platform manifests from multiple images that may have been separately built.
- Repackage a multi-platform image with only the requested platforms.
- Push and pull arbitrary OCI artifacts.
- Recursively list all content associated with an image.
- Extract files from a layer or image.
- Compare images, showing the differences between manifests, the config, and layers.
- Formatted output using Go templates.

The project website includes [usage instructions](https://regclient.org/usage/regctl/) and a [CLI reference](https://regclient.org/cli/regctl/).

## regsync features

`regsync` is an image mirroring tool.
It will copy images between two locations with the following additional features:

- Ability to run on a cron schedule, one time synchronization, or only report stale images.
- Uses a yaml configuration.
- Each source may be an entire registry (not recommended), a repository, or a single image, with the ability to filter repositories and tags.
- Support for multi-platform images, OCI referrers, “digest tags”, and copying to or from an OCI Layout (for maintaining a mirror over an air-gap).
- Ability to mirror multiple images concurrently.
- Support for copying a single platform from multi-platform images.
- Ability to backup an existing image before overwriting the tag.
- Ability to postpone mirror step when rate limit (used by Docker Hub) is below a threshold.
- Can use user’s docker configuration for user credentials and registry certificates.

The project website includes [usage instructions](https://regclient.org/usage/regsync/) and a [CLI reference](https://regclient.org/cli/regsync/).

## regbot features

`regbot` is a scripting tool on top of the `regclient` API with the following features:

- Ability to run on a cron schedule, one time execution, or test with a dry-run mode.
- Uses a yaml configuration.
- Scripts are written in Lua and executed directly in Go.
- Built-in functions include:
  - Repository list
  - Tag list
  - Image manifest (either head or get, and optional resolving multi-platform reference)
  - Image config (this includes the creation time, labels, and other details shown in a docker image inspect)
  - Image rate limit and a wait function to delay the script when rate limit remaining is below a threshold
  - Image copy
  - Manifest delete
  - Tag delete

The project website includes [usage instructions](https://regclient.org/usage/regbot/) and a [CLI reference](https://regclient.org/cli/regbot/).

## Development Status

This project is using v0 version numbers due to Go's backwards compatibility requirements of a v1 release.
The library and commands are stable for external use.
Minor version updates may contain breaking changes, however effort is made to first deprecate and provide warnings to give users time to move off of older APIs and commands.

## Installing

See the [installation instructions](https://regclient.org/install/) on the project website for the various ways to download or build CLI binaries.

## Usage

See the [project documentation](https://regclient.org/usage/).

## Contributors

<a href="https://github.com/regclient/regclient/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=regclient/regclient" alt="contributor list"/>
</a>

<!-- markdownlint-disable-file MD033 -->
