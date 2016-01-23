# Consult
Query Consul from the command line and SSH into a server from Consul catalog.

[![Build Status](https://travis-ci.org/outbrain/consult.svg?branch=master)](https://travis-ci.org/outbrain/consult)

## Building

```
go get -u github.com/outbrain/consult
```

## Using

`consult` needs to know where Consul's API is, this can be configured using the `CONSUL_URL` environment variable or the `--server` command line flag.
Run `consult -h` for a complete list of flags.

```
consult ssh --service=my_awesome_service --tag=some_tag --username=joe
```
`--username` is optional of course

To list the node for service and tag:

```
consult query --service=my_awesome_service --tag=some_tag
```

Multiple tags - you can choose to match _any_ or _all_ tags. By default, all tags must be matched.

```
consult query -s my_awesome_service -t tag1 -t tag2 -m any
```

## License

Apache V2