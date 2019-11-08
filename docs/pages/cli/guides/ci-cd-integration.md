---
title: How To Use DevSpace In CI/CD Pipelines
sidebar_label: CI/CD Integration
---

DevSpace is designed to work in non-interactive environments the same way it does in interactive environments. To use DevSpace inside a CI/CD pipeline, simply install it as part of your pipeline or start the pipeline with a VM or container image that already contains DevSpace and run any of the DevSpace commands.

> If you want to run commands such as `devspace enter`, make sure you specify a label selector using `-l`, a container using `-c` etc. to make sure DevSpace can find the right pod/container without having to ask you to select one (which would lead to your pipeline getting stuck).

If you want to use DevSpace Cloud in your CI/CD pipeline as well, you need to generate an ["Access Key"](https://app.devspace.cloud/settings/access-keys) and use it to login to DevSpace Cloud using the following non-interactive command:
```bash
devspace login --key=[YOUR_ACCESS_KEY]
```
