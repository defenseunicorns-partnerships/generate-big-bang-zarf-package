# generate-big-bang-zarf-package

This tool automatically generates a zarf.yaml alongside the requisite manifests to deploy Big Bang so that `zarf package create` is ready to be run.

## Installation
This tool can be installed with the following command
```bash
go install github.com/defenseunicorns-partnerships/generate-big-bang-zarf-package@latest
```

## Usage

Example: 
```bash
generate-big-bang-zarf-package 2.34.0 --values-file-manifests=values-files/kyverno.yaml,values-files/loki.yaml,values-files/neuvector.yaml
```

See the structure this command generated in the [example](example) folder.

Run `generate-big-bang-zarf-package -h` to see the full usage options

## Setup

An account on `https://registry1.dso.mil` to retrieve Big Bang images. You can register for an account [here](https://login.dso.mil/auth/realms/baby-yoda/protocol/openid-connect/registrations?client_id=account&response_type=code)

By default, Big Bang uses images from [Iron Bank](https://p1.dso.mil/products/iron-bank) which will require you to set your login credentials for [Registry One](https://registry1.dso.mil) (see [pre-requisites](#prerequisites) for information on account setup).

```bash
# Authenticate to https://registry1.dso.mil/, then retrieve your CLI secret from your User Profile and run the following:
set +o history
export REGISTRY1_USERNAME=<REPLACE_ME>
export REGISTRY1_CLI_SECRET=<REPLACE_ME>
echo $REGISTRY1_CLI_SECRET | zarf tools registry login registry1.dso.mil --username $REGISTRY1_USERNAME --password-stdin
set -o history
```

## Troubleshooting

See the Troubleshooting section of the Big Bang Quick Start for help troubleshooting the Big Bang deployment: https://repo1.dso.mil/big-bang/bigbang/-/blob/master/docs/guides/deployment-scenarios/quickstart.md#troubleshooting