# Bzzz + Antennae: Distributed P2P Task Coordination

Bzzz is a P2P task coordination system with the Antennae meta-discussion layer for collaborative AI reasoning. The system enables distributed AI agents to automatically discover each other, coordinate task execution, and engage in structured meta-discussions for improved collaboration.

## Architecture

- **P2P Networking**: libp2p-based mesh networking with mDNS discovery
- **Task Coordination**: GitHub Issues as atomic task units
- **Meta-Discussion**: Antennae layer for collaborative reasoning between agents
- **Distributed Logging**: Hypercore-based tamper-proof audit trails
- **Service Deployment**: SystemD service for production deployment

## Components

- `p2p/` - Core P2P networking using libp2p
- `discovery/` - mDNS peer discovery for local network
- `pubsub/` - Publish/subscribe messaging for coordination
- `github/` - GitHub API integration for task management
- `logging/` - Hypercore-based distributed logging
- `cmd/` - Command-line interfaces

## Quick Start

### Building from Source

```bash
go build -o bzzz
```

### Running as Service

Install Bzzz as a systemd service for production deployment:

```bash
# Install service (requires sudo)
sudo ./install-service.sh

# Check service status
sudo systemctl status bzzz

# View live logs
sudo journalctl -u bzzz -f

# Stop service
sudo systemctl stop bzzz

# Uninstall service
sudo ./uninstall-service.sh
```

### Running Manually

```bash
./bzzz
```

## Production Deployment

### Service Management

Bzzz is deployed as a systemd service across the cluster:

- **Auto-start**: Service starts automatically on boot
- **Auto-restart**: Service restarts on failure with 10-second delay
- **Logging**: All output captured in systemd journal
- **Security**: Runs with limited privileges and filesystem access
- **Resource Limits**: Configured file descriptor and process limits

### Cluster Status

Currently deployed on:

| Node | Service Status | Node ID | Connected Peers |
|------|----------------|---------|-----------------|
| **WALNUT** | ✅ Active | `12D3Koo...aXHoUh` | 3 peers |
| **IRONWOOD** | ✅ Active | `12D3Koo...8QbiTa` | 3 peers |
| **ACACIA** | ✅ Active | `12D3Koo...Q9YSYt` | 3 peers |

### Network Topology

Full mesh P2P network established:
- Automatic peer discovery via mDNS on `192.168.1.0/24`
- All nodes connected to all other nodes
- Capability broadcasts exchanged every 30 seconds
- Ready for distributed task coordination

## Service Configuration

The systemd service (`bzzz.service`) includes:

- **Working Directory**: `/home/tony/AI/projects/Bzzz`
- **User/Group**: `tony:tony`
- **Restart Policy**: `always` with 10-second delay
- **Security**: NoNewPrivileges, PrivateTmp, ProtectSystem
- **Logging**: Output to systemd journal with `bzzz` identifier
- **Resource Limits**: 65536 file descriptors, 4096 processes

## Development Status

This project is being developed collaboratively across the deepblackcloud cluster:
- **WALNUT**: P2P Networking Foundation (starcoder2:15b)
- **IRONWOOD**: Distributed Logging System (phi4:14b) 
- **ACACIA**: GitHub Integration Module (codellama)

## Network Configuration

- **Local Network**: 192.168.1.0/24
- **mDNS Discovery**: Automatic peer discovery with service tag `bzzz-peer-discovery`
- **PubSub Topics**: 
  - `bzzz/coordination/v1` - Task coordination messages
  - `antennae/meta-discussion/v1` - Collaborative reasoning
- **Security**: Message signing and signature verification enabled

## Related Projects

- **[Hive](https://github.com/anthonyrawlins/hive)** - Multi-Agent Task Coordination System
- **[Antennae](https://github.com/anthonyrawlins/antennae)** - AI Collaborative Reasoning Protocol