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

## Building

```bash
git clone https://github.com/bofrim/gorch
cd gorch
go build -o gorch gorch.go
```

## Usage

### Running an orchestrator

```bash
./gorch orchestrator \
  --port 8322 \
  --log /some/path/to/gorch_log.txt # optional
```

### Running a node

```bash
./gorch node \        
  --data /some/path/to/data_dir \
  --actions /some/path/to/actions.yaml \
  --name cool_node_1 \
  --orchestrator "127.0.0.1:8322"
```

### Running user operations

```bash
# Get info about the orchestrator
./gorch user info \
  --orchestrator "127.0.0.1:8322"

# Get all the data from a node
./gorch user data \
  --orchestrator "127.0.0.1:8322" \
  --node cool_node_1 \
  --json # optional

# Get a specific json file from a node
./gorch user data \
  --orchestrator "127.0.0.1:8322" \
  --node cool_node_1 \
  --path asdf \
  --json # optional

# Run an action on a node
./gorch user action \
  --orchestrator "127.0.0.1:8322" \
  --node cool_node_1 \
  --action hello \
  --data message=hello \
  --data other=world

# Run an action on a node and stream output
./gorch user action \
  --orchestrator "127.0.0.1:8322" \
  --node cool_node_1 \
  --action sleep \
  --data time=5 \
  --stream-port 8323
```

## Setting up an actions file

```yaml
# actions.yaml

"list":
  description: "List the contents of a directory"
  params: []
  commands:
    - "ls"

"hello":
  description: "Print a message"
  params: ["message", "other"]
  commands:
    - "echo {{.message}}"
    - "echo {{.other}}"

"sleep":
  description: "Send a message, sleep, then send another message"
  params: ["time"]
  commands:
    - "date"
    - "sleep {{.time}}"
    - "date"
```

## TODO

* [ ] Add a way to specify a configuration file for a node
* [ ] Add a way to run periodic actions on a node (should be an optional configuration option for a node) Figure out what to do with the output of the action.
* [ ] Setup web hooks for data changes or events related to actions
* [ ] Setup centralized logging for nodes so logs will be accessible through the orchestrator even if the node is offline
* [ ] Add a user command to stream logs from either the orchestrator or a specific node
