# Gorch

**Note**: This module is still very early in development. It should be considered pre-pre-alpha.

**Warning**: By definition portions of this module will be used for remote code execution. Ensure you understand the security implications of this before using this module.

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
  --cert-path /path/to/pem/certs \
  --log /some/path/to/gorch_log.txt # optional
```

### Running a node

```bash
./gorch node \
  --cert-path /path/to/pem/certs \
  --data /some/path/to/data_dir \
  --actions /some/path/to/actions.yaml \
  --name cool_node_1 \
  --orchestrator "127.0.0.1:443" \
  --port 8776 \ # optional
  --arbitrary-actions # Optional; Danger: allows arbitrary code execution

```

#### Setting up an actions file for a node

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

### Running user operations

Get info about the orchestrator

```bash
./gorch user info \
  --orchestrator "127.0.0.1:443"
```

Get all the data from a node

```bash
./gorch user data \
  --orchestrator "127.0.0.1:443" \
  --node cool_node_1 \
  --json # optional

```

Get a specific json file from a node

```bash
./gorch user data \
  --orchestrator "127.0.0.1:443" \
  --node cool_node_1 \
  --path asdf \
  --json # optional

```

Run an action on a node

```bash
./gorch user action \
  --orchestrator "127.0.0.1:443" \
  --node cool_node_1 \
  --action hello \
  --data message=hello \
  --data other=world
```

Run an action on a node and stream output.

```bash
./gorch user action \
  --orchestrator "127.0.0.1:443" \
  --node cool_node_1 \
  --action sleep \
  --data time=5 \
  --stream-port 8323
```

Specify a data file to use as the body of the request

```bash
./gorch user action \
  --orchestrator "127.0.0.1:443" \
  --node cool_node_1 \
  --action sleep \
  --data-file params.json \
  --stream-port 8323
```

Run arbitrary commands on a node
(Note: The node must be running with the `--arbitrary-actions` flag set)

```bash
.gorch user action \
  --node brad \
  --data-file adhoc.json \
  --data message="hello" \  # data can be specified in the data-file, or as a flag
  --stream-port 8323
```

Where `adhoc.json` is:

```json
{
  "action": {
    "name": "adhoc-list",
    "description": "List the contents of a directory",
    "params": ["dir", "message"],
    "commands": ["ls {{.dir}}", "echo {{.message}}"]
  },
  "dir": "/path/to/list"
}
```

## TODO

- [ ] Add a way to specify a configuration file for a node
- [ ] Add a way to run periodic actions on a node (should be an optional configuration option for a node) Figure out what to do with the output of the action.
- [ ] Setup web hooks for data changes or events related to actions
- [ ] Setup centralized logging for nodes so logs will be accessible through the orchestrator even if the node is offline
- [ ] Add a user command to stream logs from either the orchestrator or a specific node
- [ ] Gracefully handle errors in the actions
- [ ] Hook listeners should have IDs for actions that are tracked on the node side

- [ ] webhook for action completion
- [ ] Generate TLS certs on the fly
- [ ] TLS for action streaming
- [ ] some basic form of auth even if it's just a shared secret
- [ ] a flag to chose if its able to run on a real network
- [ ] Ability to list currently running actions (with info about them; params, age, etc)
- [ ] Ability to kill a running action
- [ ] a front end for the orchestrator and nodes

- [ ] a broadcast command to run action a set of nodes
- [ ] number of actions running to be specified on a group of actions rather than the entire node (i.e. should be able to run status action if there is a long running worker action)
