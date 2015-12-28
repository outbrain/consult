# consul-ssh
SSH into a server from Consul catalog.

## Building

```
go get github.com/avishai-ish-shalom/consul-ssh
```

## Using

`consul-ssh` needs to know where Consul's API is, this can be configured using the `CONSUL_URL` environment variable or the `--url` command line flag.
Run `consul-ssh --help` for a complete list of flags.

```
consul-ssh --service=my_awesome_service --tag=some_tag --username=joe
```
