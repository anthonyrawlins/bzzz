# Bzzz P2P Service Deployment Guide

This document provides detailed instructions for deploying Bzzz as a production systemd service across multiple nodes.

## Overview

Bzzz has been successfully deployed as a systemd service across the deepblackcloud cluster, providing:
- Automatic startup on boot
- Automatic restart on failure
- Centralized logging via systemd journal
- Security sandboxing and resource limits
- Full mesh P2P network connectivity

## Installation Steps

### 1. Build Binary

```bash
cd /home/tony/AI/projects/Bzzz
go build -o bzzz
```

### 2. Install Service

```bash
# Install as systemd service (requires sudo)
sudo ./install-service.sh
```

The installation script:
- Makes the binary executable
- Copies service file to `/etc/systemd/system/bzzz.service`
- Reloads systemd daemon
- Enables auto-start on boot
- Starts the service immediately

### 3. Verify Installation

```bash
# Check service status
sudo systemctl status bzzz

# View recent logs
sudo journalctl -u bzzz -n 20

# Follow live logs
sudo journalctl -u bzzz -f
```

## Current Deployment Status

### Cluster Overview

| Node | IP Address | Service Status | Node ID | Connected Peers |
|------|------------|----------------|---------|-----------------|
| **WALNUT** | 192.168.1.27 | ✅ Active | `12D3KooWEeVXdHkXtUp2ewzdqD56gDJCCuMGNAqoJrJ7CKaXHoUh` | 3 peers |
| **IRONWOOD** | 192.168.1.113 | ✅ Active | `12D3KooWFBSR...8QbiTa` | 3 peers |
| **ACACIA** | 192.168.1.xxx | ✅ Active | `12D3KooWE6c...Q9YSYt` | 3 peers |

### Network Connectivity

Full mesh P2P network established:

```
    WALNUT (aXHoUh)
       ↕    ↕
      ↙      ↘
IRONWOOD ←→ ACACIA
(8QbiTa)   (Q9YSYt)
```

- All nodes automatically discovered via mDNS
- Bidirectional connections established
- Capability broadcasts exchanged every 30 seconds
- Ready for distributed task coordination

## Service Management

### Basic Commands

```bash
# Start service
sudo systemctl start bzzz

# Stop service
sudo systemctl stop bzzz

# Restart service
sudo systemctl restart bzzz

# Check status
sudo systemctl status bzzz

# Enable auto-start (already enabled)
sudo systemctl enable bzzz

# Disable auto-start
sudo systemctl disable bzzz
```

### Logging

```bash
# View recent logs
sudo journalctl -u bzzz -n 50

# Follow live logs
sudo journalctl -u bzzz -f

# View logs from specific time
sudo journalctl -u bzzz --since "2025-07-12 19:00:00"

# View logs with specific priority
sudo journalctl -u bzzz -p info
```

### Troubleshooting

```bash
# Check if service is running
sudo systemctl is-active bzzz

# Check if service is enabled
sudo systemctl is-enabled bzzz

# View service configuration
sudo systemctl cat bzzz

# Reload service configuration (after editing service file)
sudo systemctl daemon-reload
sudo systemctl restart bzzz
```

## Service Configuration

### Service File Location

`/etc/systemd/system/bzzz.service`

### Key Configuration Settings

- **Type**: `simple` - Standard foreground service
- **User/Group**: `tony:tony` - Runs as non-root user
- **Working Directory**: `/home/tony/AI/projects/Bzzz`
- **Restart Policy**: `always` with 10-second delay
- **Timeout**: 30-second graceful stop timeout

### Security Settings

- **NoNewPrivileges**: Prevents privilege escalation
- **PrivateTmp**: Isolated temporary directory
- **ProtectSystem**: Read-only system directories
- **ProtectHome**: Limited home directory access

### Resource Limits

- **File Descriptors**: 65,536 (for P2P connections)
- **Processes**: 4,096 (for Go runtime)

## Network Configuration

### Port Usage

Bzzz automatically selects available ports for P2P communication:
- TCP ports in ephemeral range (32768-65535)
- IPv4 and IPv6 support
- Automatic port discovery and sharing via mDNS

### Firewall Considerations

For production deployments:
- Allow inbound TCP connections on used ports
- Allow UDP port 5353 for mDNS discovery
- Consider restricting to local network (192.168.1.0/24)

### mDNS Discovery

- Service Tag: `bzzz-peer-discovery`
- Network Scope: `192.168.1.0/24`
- Discovery Interval: Continuous background scanning

## Monitoring and Maintenance

### Health Checks

```bash
# Check P2P connectivity
sudo journalctl -u bzzz | grep "Connected to"

# Monitor capability broadcasts
sudo journalctl -u bzzz | grep "capability_broadcast"

# Check for errors
sudo journalctl -u bzzz -p err
```

### Performance Monitoring

```bash
# Resource usage
sudo systemctl status bzzz

# Memory usage
ps aux | grep bzzz

# Network connections
sudo netstat -tulpn | grep bzzz
```

### Maintenance Tasks

1. **Log Rotation**: Systemd handles log rotation automatically
2. **Service Updates**: Stop service, replace binary, restart
3. **Configuration Changes**: Edit service file, reload systemd, restart

## Uninstalling

To remove the service:

```bash
sudo ./uninstall-service.sh
```

This will:
- Stop the service if running
- Disable auto-start
- Remove service file
- Reload systemd daemon
- Reset any failed states

Note: Binary and project files remain intact.

## Deployment Timeline

- **2025-07-12 19:46**: WALNUT service installed and started
- **2025-07-12 19:49**: IRONWOOD service installed and started  
- **2025-07-12 19:49**: ACACIA service installed and started
- **2025-07-12 19:50**: Full mesh network established (3 nodes)

## Next Steps

1. **Integration**: Connect with Hive task coordination system
2. **Monitoring**: Set up centralized monitoring dashboard
3. **Scaling**: Add additional nodes to expand P2P mesh
4. **Task Execution**: Implement actual task processing workflows