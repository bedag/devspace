---
title: Command - devspace add sync
sidebar_label: sync
id: version-v4.2.0-devspace_add_sync
original_id: devspace_add_sync
---


Add a sync path

## Synopsis


```
devspace add sync [flags]
```

```
#######################################################
################# devspace add sync ###################
#######################################################
Add a sync path to this project's devspace.yaml

Example:
devspace add sync --local=app --container=/app
#######################################################
```
## Options

```
      --container string        Absolute container path
      --exclude string          Comma separated list of paths to exclude (e.g. node_modules/,bin,*.exe)
  -h, --help                    help for sync
      --label-selector string   Comma separated key=value selector list (e.g. release=test)
      --local string            Relative local path
```

### Options inherited from parent commands

```
      --debug                 Prints the stack trace if an error occurs
      --kube-context string   The kubernetes context to use
  -n, --namespace string      The kubernetes namespace to use
      --no-warn               If true does not show any warning when deploying into a different namespace or kube-context than before
  -p, --profile string        The devspace profile to use (if there is any)
      --silent                Run in silent mode and prevents any devspace log output except panics & fatals
  -s, --switch-context        Switches and uses the last kube context and namespace that was used to deploy the DevSpace project
      --var strings           Variables to override during execution (e.g. --var=MYVAR=MYVALUE)
```

## See Also

* [devspace add](../../cli/commands/devspace_add)	 - Convenience command: adds something to devspace.yaml
