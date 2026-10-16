# AutoK3s

## What is AutoK3s

English / [简体中文](docs/i18n/zh_cn/README.md)

[K3s](https://k3s.io/) is a Lightweight Kubernetes which is great for Edge, IoT, CI, Development, ARM and so on.

AutoK3s is a lightweight tool for simplifying the cluster management of K3s, it can help you **Run K3s Everywhere**.

<!-- toc -->

- [Key Features](#key-features)
- [Providers](#providers)
- [Quick Start (tl;dr)](#quick-start-tldr)
- [Demo Video](#demo-video)
- [Developers' Guide](#developers-guide)
- [Source Code](#source-code)

<!-- /toc -->

## Key Features

- Flexible installation options, like K3s cluster HA and datastore(embedded etcd, RDS, SQLite, etc.).
- Simplify operations by UI dashboard.
- [Airgap](docs/i18n/en_us/airgap/README.md) K3s package management and installation.
- [Add-on](docs/i18n/en_us/add-on/README.md) streamlines the management and deployment of plugins, seamlessly integrating applications into K3s clusters.

## Providers

Now supports the following providers, we encourage submitting PR contribution for more providers:

- [aws](docs/i18n/en_us/aws/README.md) - Bootstrap K3s onto Amazon EC2
- [google](docs/i18n/en_us/google/README.md) - Bootstrap K3s onto Google Compute Engine
- [alibaba](docs/i18n/en_us/alibaba/README.md) - Bootstrap K3s onto Alibaba ECS
- [tencent](docs/i18n/en_us/tencent/README.md) - Bootstrap K3s onto Tencent CVM
- [k3d](docs/i18n/en_us/k3d/README.md) - Bootstrap K3d onto Local Machine
- [native](docs/i18n/en_us/native/README.md) - Bootstrap K3s onto any VM

## Quick Start (tl;dr)

Scenario 1 - Run with docker:

```bash
# The commands will start autok3s daemon with an interactionable UI.
# Standard Linux (Recommended)
$ docker run -itd --restart=unless-stopped --net=host -v /var/run/docker.sock:/var/run/docker.sock orstar/autok3s:v0.10.0

# MacOS
$ docker run -itd --restart=unless-stopped -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_HOST="" orstar/autok3s:v0.10.0
```

Scenario 2 - Run with docker-compose

```bash
# AutoK3s is available at: http://autok3s.vcap.me (vcap.me resolves to 127.0.0.1).
# The cli and container uses the same `AUTOK3S_CONFIG` directory.
docker-compose up -d
```

Scenario 3 - Run with cli:

Download the executable for your platform from the [Releases](https://github.com/orangedeng/autok3s/releases) page, then:

```bash
chmod +x ./autok3s
sudo mv ./autok3s /usr/local/bin

# The commands will start autok3s daemon and popup default browser with an interactionable UI.
autok3s -d serve
```

## Uninstall

To uninstall autok3s cli, please execute commands below:

```bash
sudo rm -f /usr/local/bin/autok3s
rm -rf ~/.autok3s
```

## Demo Video

> **Versioning**: starting from `v0.10.0`, this project is maintained
> as an independent upstream. Earlier releases were forked from
> [cnrancher/autok3s](https://github.com/cnrancher/autok3s) (last
> upstream release: v0.9.3).

In the demo, we install K3s by k3d provider in around 1 minute, and
we can manage K3s clusters with the bundled web ui.

Watch the demo:

![](docs/assets/autok3s-v0.7.0-demo.gif)

## Use Cases

- [Quickly create K3s cluster and add K3s nodes with GUI Tool using AutoK3s (Installation guide in WSL)](https://dev.to/msh2050/quickly-create-k3s-cluster-and-add-k3s-nodes-with-gui-tool-using-autok3s-installation-guide-in-wsl-3e5d)

## Developers' Guide

Use `make` to manage project compilation, testing and packaging.

- compilation: `make autok3s build`
- testing: `make autok3s unit`
- packing: `make autok3s package only`

## Source Code

AutoK3s is 100% open-source software.
