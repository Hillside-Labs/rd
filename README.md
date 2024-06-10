# Remote Docker

Remote Docker (rd) takes inspiration from [Kamal](https://kamal-deploy.org/) with the primary difference being that instead of managing raw docker containers, you use a docker compose file to manage an application. The result is that you get a little more consistency b/w your local development and production, and you don't have a custom config format to think about.

## How to use it?

First create a `deploy` directory in your repo. This is not necessary, but just a pattern to use. In the deploy directory you create an `infra` directory where you provision your infra with Terraform. Rd uses the tfstate to find the hosts to run commands on.

With a vm provisioned, you can run `rd bootstrap`. This will install and start the docker daemon on the machine.

From there, you can `rd sync` to sync your docker compose to the remote hosts, `rd pull` to grab the containers in your compose file, and `rd run` to start things up.

## Filtering Nodes

Rd will first try to find the availble nodes to connect to by looking in the `infra` directory and running `terraform show -json`. You can filter what hosts you want the commands to run on by passing some flags.

```sh
❯ rd run --help
NAME:
   rd run

USAGE:
   rd run [command options] [arguments...]

OPTIONS:
   --name value, -n value     filter hosts by name
   --ip value, -i value       filter hosts by public ip
   --private value, -p value  filter hosts by public ip
   --help, -h                 show help
```

The `--name` flag also works as a prefix filter. Naming nodes then with prefixes will help selectively run operations.

The `rd` tool has some helpful subcommands.

 - `rdo sync` copies a local file or directory to the remote machine. You can use `rdo sync -r` to copy directories and `rdo sync` to copy a file. This does not define a remote directory or target and will simply put the file in the default directory as defined by the ssh server. Usually this will be the home directory of the user.

 - `rdo run` runs a command on the remote machine. The stdout and stderr will be printed along with a prefix of the host that the command is running on.

There are also commands to run typical docker compose commands like `rd logs`.

## Infra and Terraform

In the app deploy directory, there should be an `infra` folder. That folder should have the terraform code necessary to spin up, scale up, and generally run the application. We're using DigitalOcean VMs, but
this could be any infrastructure on any cloud.
