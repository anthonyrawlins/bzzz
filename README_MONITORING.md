# Bzzz Antennae Monitoring Dashboard

A real-time console monitoring dashboard for the Bzzz P2P coordination system, similar to btop/nvtop for system monitoring.

## Features

🔍 **Real-time P2P Status**
- Connected peer count with history graph
- Node ID and network status
- Hive API connectivity status

🤖 **Agent Activity Monitoring**
- Live agent availability updates
- Agent status distribution (ready/working/busy)
- Recent activity tracking

🎯 **Coordination Activity**
- Task announcements and completions
- Coordination session tracking
- Message flow statistics

📊 **Visual Elements**
- ASCII graphs for historical data
- Color-coded status indicators
- Live activity log with timestamps

## Usage

### Basic Usage
```bash
# Run with default 1-second refresh rate
python3 cmd/bzzz-monitor.py

# Custom refresh rate (2 seconds)
python3 cmd/bzzz-monitor.py --refresh-rate 2.0

# Disable colors for logging/screenshots
python3 cmd/bzzz-monitor.py --no-color
```

### Installation as System Command
```bash
# Copy to system bin
sudo cp cmd/bzzz-monitor.py /usr/local/bin/bzzz-monitor
sudo chmod +x /usr/local/bin/bzzz-monitor

# Now run from anywhere
bzzz-monitor
```

## Dashboard Layout

```
┌─ Bzzz P2P Coordination Monitor ─┐
│ Uptime: 0:02:15 │ Node: 12*SEE3To... │
└───────────────────────────────────┘

P2P Network Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Connected Peers: 2
Hive API Status: Offline (Overlay Network Issues)

Peer History (last 20 samples):
███▇▆▆▇████▇▆▇███▇▆▇ (1-3 peers)

Agent Activity  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Recent Updates (1m): 8
  Ready: 6
  Working: 2

Coordination Activity
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total Messages: 45
Total Tasks: 12
Active Sessions: 1
Recent Tasks (5m): 8

Recent Activity
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
11:10:35 [AVAIL] Agent acacia-node... status: ready
11:10:33 [TASK]  Task announcement: hive#15 - WebSocket support
11:10:30 [COORD] Meta-coordination session started
11:10:28 [AVAIL] Agent ironwood-node... status: working
11:10:25 [ERROR] Failed to get active repositories: API 404

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Press Ctrl+C to exit | Refresh rate: 1.0s
```

## Monitoring Data Sources

The dashboard pulls data from:

1. **Systemd Service Logs**: `journalctl -u bzzz.service`
2. **P2P Network Status**: Extracted from bzzz log messages
3. **Agent Availability**: Parsed from availability_broadcast messages
4. **Task Activity**: Detected from task/repository-related log entries
5. **Error Tracking**: Monitors for failures and connection issues

## Color Coding

- 🟢 **Green**: Good status, active connections, ready agents
- 🟡 **Yellow**: Working status, moderate activity 
- 🔴 **Red**: Errors, failed connections, busy agents
- 🔵 **Blue**: Information, neutral data
- 🟣 **Magenta**: Coordination-specific activity
- 🔷 **Cyan**: Network and P2P data

## Real-time Updates

The dashboard updates every 1-2 seconds by default and tracks:

- **P2P Connections**: Shows immediate peer join/leave events
- **Agent Status**: Real-time availability broadcasts from all nodes
- **Task Flow**: Live task announcements and coordination activity
- **System Health**: Continuous monitoring of service status and errors

## Performance

- **Low Resource Usage**: Python-based with minimal CPU/memory impact
- **Efficient Parsing**: Only processes recent logs (last 30-50 lines)
- **Responsive UI**: Fast refresh rates without overwhelming the terminal
- **Historical Data**: Maintains rolling buffers for trend analysis

## Troubleshooting

### No Data Appearing
```bash
# Check if bzzz service is running
systemctl status bzzz.service

# Verify log access permissions
journalctl -u bzzz.service --since "1 minute ago"
```

### High CPU Usage
```bash
# Reduce refresh rate
bzzz-monitor --refresh-rate 5.0
```

### Color Issues
```bash
# Disable colors
bzzz-monitor --no-color

# Check terminal color support
echo $TERM
```

## Integration

The monitor works alongside:
- **Live Bzzz System**: Monitors real P2P mesh (WALNUT/ACACIA/IRONWOOD)
- **Test Suite**: Can monitor test coordination scenarios  
- **Development**: Perfect for debugging antennae coordination logic

## Future Enhancements

- 📈 Export metrics to CSV/JSON
- 🔔 Alert system for critical events
- 📊 Web-based dashboard version
- 🎯 Coordination session drill-down
- 📱 Mobile-friendly output