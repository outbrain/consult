# Consult
Query Consul from the command line and SSH into a server from Consul catalog.

[![Build Status](https://travis-ci.org/outbrain/consult.svg?branch=master)](https://travis-ci.org/outbrain/consult)

## Building

```
go get -u github.com/outbrain/consult
```

## Using

`consult` needs to know where Consul's API is, this can be configured using the `CONSUL_URL` environment variable or the `--server` command line flag.
Run `consult -h` or `consult help` for a complete list of flags.

### Misc flags

* `--json`, `-j` - JSON output
* `--detailed`, `-d` - Print detailed results with header line

### Query subcommand

To list the node for service and tag:

```
consult query --service=my_awesome_service --tag=some_tag
```

Multiple tags - you can choose to match _any_ or _all_ tags. By default, all tags must be matched.

```
consult query -s my_awesome_service -t tag1 -t tag2 -m any
```

Multiple datacenter support
```
consult query -s service1 --dc dc1 --dc dc2
```

`query` flags also work on `http` and `ssh` subcommands.

### ssh subcommand

```
consult ssh --service=my_awesome_service --tag=some_tag --username=joe
```
`--username` is optional of course


### http subcommand

Perform an HTTP request on endpoints returned from Consul query:

```
consult http -s service1 -t tag1 --all-endpoints --uri="/ping"
```

For other options check out `consul help http`

### List subcommands

```
consult list dc
consult list service --dc dc1 --dc dc2
consult list node -r '^ap.*a' # list node matching regex
```

## License

Apache V2