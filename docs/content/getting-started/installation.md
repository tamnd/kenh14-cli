---
title: "Installation"
description: "Install kenh14 from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/kenh14-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `kenh14` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/kenh14-cli/cmd/kenh14@latest
```

That puts `kenh14` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/kenh14-cli
cd kenh14-cli
make build        # produces ./bin/kenh14
./bin/kenh14 version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/kenh14:latest --help
```

## Checking the install

```bash
kenh14 version
```

prints the version and exits.
