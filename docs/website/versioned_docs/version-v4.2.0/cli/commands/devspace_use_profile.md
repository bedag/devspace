---
title: Command - devspace use profile
sidebar_label: profile
id: version-v4.2.0-devspace_use_profile
original_id: devspace_use_profile
---


Use a specific DevSpace profile

## Synopsis


```
devspace use profile [flags]
```

```
#######################################################
################ devspace use profile #################
#######################################################
Use a specific DevSpace profile

Example:
devspace use profile production
devspace use profile staging
devspace use profile --reset
#######################################################
```
## Options

```
  -h, --help    help for profile
      --reset   Don't use a profile anymore
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

* [devspace use](../../cli/commands/devspace_use)	 - Use specific config
