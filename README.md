# Gorch


__Note__: This module is still very early in development. It should be considered pre-pre-alpha.

__Warning__: By definition portions of this module will be used for remote code execution. Ensure you understand the security implications of this before using this module.

<p align="center">
  <img src="https://cdn.discordapp.com/attachments/1055542894221602816/1067475495580610590/image.png" alt="gorch mascot" width="200"/>
</p>

## About
Gorch (pronounced gork) is a tool that can be used to interface with and manage multiple remote nodes.
Drop json files into your node's data directory and gorch will serve them for you.

Gorch is also able to run remote actions on your nodes. Specify a configuration file when starting your node and gorch will provide an interface for executing those actions.

## Installation

### From source

```bash
git clone https://github.com/bofrim/gorch
cd gorch
go run gorch.go
```

## TODO

* [ ] Add a way to specify a configuration file for a node
* [ ] Add a way to run periodic actions on a node
* [ ] Provide an interface for users to get data from nodes through the orchestrator
* [ ] Add streaming output for actions
* [ ] Setup web hooks for data changes or events related to actions
* [ ] Setup centralized logging for nodes so logs will be accessible through the orchestrator even if the node is offline
