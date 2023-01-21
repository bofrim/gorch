# Gorch

__Note__: This module is still very early in development. It should be considered pre-pre-alpha.

__Warning__: By definition portions of this module will be used for remote code execution. Ensure you understand the security implications of this before using this module.

## About
Gorch is a tool that can be used to interface with and manage multiple remote nodes.
Drop json files into your node's data directory and gorch will serve them for you.

(Future) Gorch will also be able to run remote actions on your nodes. Specify a configuration file when starting your node and gorch will provide an interface for executing those actions.

## Installation

### From source

```bash
git clone https://github.com/bofrim/gorch
cd gorch
go run gorch.go
```

## TODO

* [ ] Add a central that can gateway requests to multiple nodes
* [ ] Figure out how to register and run remote actions (registration can happen at node startup)
