# Bzzz + Antennae: Distributed P2P Task Coordination

Bzzz is a P2P task coordination system with the Antennae meta-discussion layer for collaborative AI reasoning.

## Architecture

- **P2P Networking**: libp2p-based mesh networking with mDNS discovery
- **Task Coordination**: GitHub Issues as atomic task units
- **Meta-Discussion**: Antennae layer for collaborative reasoning between agents
- **Distributed Logging**: Hypercore-based tamper-proof audit trails

## Components

- `p2p/` - Core P2P networking using libp2p
- `discovery/` - mDNS peer discovery for local network
- `pubsub/` - Publish/subscribe messaging for coordination
- `github/` - GitHub API integration for task management
- `logging/` - Hypercore-based distributed logging
- `cmd/` - Command-line interfaces

## Development Status

This project is being developed collaboratively across the deepblackcloud cluster:
- **WALNUT**: P2P Networking Foundation (starcoder2:15b)
- **IRONWOOD**: Distributed Logging System (phi4:14b) 
- **ACACIA**: GitHub Integration Module (codellama)

## Network Configuration

- **Local Network**: 192.168.1.0/24
- **mDNS Discovery**: Automatic peer discovery
- **Docker Deployment**: Host networking mode for P2P connectivity