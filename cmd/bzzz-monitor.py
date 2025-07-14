#!/usr/bin/env python3
"""
Bzzz Antennae Real-time Monitoring Dashboard
Similar to btop/nvtop for system monitoring, but for P2P coordination activity

Usage: python3 bzzz-monitor.py [--refresh-rate 1.0]
"""

import argparse
import json
import os
import subprocess
import sys
import time
from collections import defaultdict, deque
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple

# Color codes for terminal output
class Colors:
    RESET = '\033[0m'
    BOLD = '\033[1m'
    DIM = '\033[2m'
    
    # Standard colors
    RED = '\033[31m'
    GREEN = '\033[32m'
    YELLOW = '\033[33m'
    BLUE = '\033[34m'
    MAGENTA = '\033[35m'
    CYAN = '\033[36m'
    WHITE = '\033[37m'
    
    # Bright colors
    BRIGHT_RED = '\033[91m'
    BRIGHT_GREEN = '\033[92m'
    BRIGHT_YELLOW = '\033[93m'
    BRIGHT_BLUE = '\033[94m'
    BRIGHT_MAGENTA = '\033[95m'
    BRIGHT_CYAN = '\033[96m'
    BRIGHT_WHITE = '\033[97m'
    
    # Background colors
    BG_RED = '\033[41m'
    BG_GREEN = '\033[42m'
    BG_YELLOW = '\033[43m'
    BG_BLUE = '\033[44m'
    BG_MAGENTA = '\033[45m'
    BG_CYAN = '\033[46m'

class BzzzMonitor:
    def __init__(self, refresh_rate: float = 1.0):
        self.refresh_rate = refresh_rate
        self.start_time = datetime.now()
        
        # Data storage
        self.p2p_history = deque(maxlen=60)  # Last 60 data points
        self.availability_history = deque(maxlen=100)
        self.task_history = deque(maxlen=50)
        self.error_history = deque(maxlen=20)
        self.coordination_sessions = {}
        self.agent_stats = defaultdict(lambda: {'messages': 0, 'tasks': 0, 'last_seen': None})
        
        # Current stats
        self.current_peers = 0
        self.current_node_id = "Unknown"
        self.total_messages = 0
        self.total_tasks = 0
        self.total_errors = 0
        self.api_status = "Unknown"
        
        # Terminal size
        self.update_terminal_size()
    
    def update_terminal_size(self):
        """Get current terminal size"""
        try:
            result = subprocess.run(['stty', 'size'], capture_output=True, text=True)
            lines, cols = map(int, result.stdout.strip().split())
            self.terminal_height = lines
            self.terminal_width = cols
        except:
            self.terminal_height = 24
            self.terminal_width = 80
    
    def clear_screen(self):
        """Clear the terminal screen"""
        print('\033[2J\033[H', end='')
    
    def get_bzzz_status(self) -> Dict:
        """Get current bzzz service status"""
        try:
            # Get systemd status
            result = subprocess.run(['systemctl', 'is-active', 'bzzz.service'], 
                                  capture_output=True, text=True)
            service_status = result.stdout.strip()
            
            # Get recent logs for analysis
            result = subprocess.run([
                'journalctl', '-u', 'bzzz.service', '--since', '30 seconds ago', '-n', '50'
            ], capture_output=True, text=True)
            
            recent_logs = result.stdout
            return {
                'service_status': service_status,
                'logs': recent_logs,
                'timestamp': datetime.now()
            }
        except Exception as e:
            return {
                'service_status': 'error',
                'logs': f"Error getting status: {e}",
                'timestamp': datetime.now()
            }
    
    def parse_logs(self, logs: str):
        """Parse bzzz logs and extract coordination data"""
        lines = logs.split('\n')
        
        for line in lines:
            timestamp = datetime.now()
            
            # Extract node ID
            if 'Node Status - ID:' in line:
                try:
                    node_part = line.split('Node Status - ID: ')[1].split(',')[0]
                    self.current_node_id = node_part.strip()
                except:
                    pass
            
            # Extract peer count
            if 'Connected Peers:' in line:
                try:
                    peer_count = int(line.split('Connected Peers: ')[1].split()[0])
                    self.current_peers = peer_count
                    self.p2p_history.append({
                        'timestamp': timestamp,
                        'peers': peer_count
                    })
                except:
                    pass
            
            # Track availability broadcasts (agent activity)
            if 'availability_broadcast' in line:
                try:
                    # Extract agent info from availability broadcast
                    if 'node_id:' in line and 'status:' in line:
                        agent_id = "unknown"
                        status = "unknown"
                        
                        # Parse the log line for agent details
                        parts = line.split('node_id:')
                        if len(parts) > 1:
                            agent_part = parts[1].split()[0].strip('<>')
                            agent_id = agent_part
                        
                        if 'status:' in line:
                            status_part = line.split('status:')[1].split()[0]
                            status = status_part
                        
                        self.availability_history.append({
                            'timestamp': timestamp,
                            'agent_id': agent_id,
                            'status': status
                        })
                        
                        self.agent_stats[agent_id]['last_seen'] = timestamp
                        self.agent_stats[agent_id]['messages'] += 1
                        
                except:
                    pass
            
            # Track task activity
            if any(keyword in line.lower() for keyword in ['task', 'repository', 'github']):
                self.task_history.append({
                    'timestamp': timestamp,
                    'activity': line.strip()
                })
                self.total_tasks += 1
            
            # Track errors
            if any(keyword in line.lower() for keyword in ['error', 'failed', 'cannot']):
                self.error_history.append({
                    'timestamp': timestamp,
                    'error': line.strip()
                })
                self.total_errors += 1
                
                # Check API status
                if 'Failed to get active repositories' in line:
                    self.api_status = "Offline (Overlay Network Issues)"
                elif 'API request failed' in line:
                    self.api_status = "Error"
            
            # Track coordination activity (when antennae system is active)
            if any(keyword in line.lower() for keyword in ['coordination', 'antennae', 'meta']):
                # This would track actual coordination sessions
                pass
    
    def draw_header(self):
        """Draw the header section"""
        uptime = datetime.now() - self.start_time
        uptime_str = str(uptime).split('.')[0]  # Remove microseconds
        
        header = f"{Colors.BOLD}{Colors.BRIGHT_CYAN}┌─ Bzzz P2P Coordination Monitor ─┐{Colors.RESET}"
        status_line = f"{Colors.CYAN}│{Colors.RESET} Uptime: {uptime_str} {Colors.CYAN}│{Colors.RESET} Node: {self.current_node_id[:12]}... {Colors.CYAN}│{Colors.RESET}"
        separator = f"{Colors.CYAN}└───────────────────────────────────┘{Colors.RESET}"
        
        print(header)
        print(status_line)
        print(separator)
        print()
    
    def draw_p2p_status(self):
        """Draw P2P network status section"""
        print(f"{Colors.BOLD}{Colors.BRIGHT_GREEN}P2P Network Status{Colors.RESET}")
        print("━" * 30)
        
        # Current peers
        peer_color = Colors.BRIGHT_GREEN if self.current_peers > 0 else Colors.BRIGHT_RED
        print(f"Connected Peers: {peer_color}{self.current_peers}{Colors.RESET}")
        
        # API Status
        api_color = Colors.BRIGHT_RED if "Error" in self.api_status or "Offline" in self.api_status else Colors.BRIGHT_GREEN
        print(f"Hive API Status: {api_color}{self.api_status}{Colors.RESET}")
        
        # Peer connection history (mini graph)
        if len(self.p2p_history) > 1:
            print(f"\nPeer History (last {len(self.p2p_history)} samples):")
            self.draw_mini_graph([p['peers'] for p in self.p2p_history], "peers")
        
        print()
    
    def draw_agent_activity(self):
        """Draw agent activity section"""
        print(f"{Colors.BOLD}{Colors.BRIGHT_YELLOW}Agent Activity{Colors.RESET}")
        print("━" * 30)
        
        if not self.availability_history:
            print(f"{Colors.DIM}No recent agent activity{Colors.RESET}")
        else:
            # Recent availability updates
            recent_count = len([a for a in self.availability_history if 
                              datetime.now() - a['timestamp'] < timedelta(minutes=1)])
            print(f"Recent Updates (1m): {Colors.BRIGHT_YELLOW}{recent_count}{Colors.RESET}")
            
            # Agent status summary
            agent_counts = defaultdict(int)
            for activity in list(self.availability_history)[-10:]:  # Last 10 activities
                if activity['status'] in ['ready', 'working', 'busy']:
                    agent_counts[activity['status']] += 1
            
            for status, count in agent_counts.items():
                status_color = {
                    'ready': Colors.BRIGHT_GREEN,
                    'working': Colors.BRIGHT_YELLOW,
                    'busy': Colors.BRIGHT_RED
                }.get(status, Colors.WHITE)
                print(f"  {status.title()}: {status_color}{count}{Colors.RESET}")
        
        print()
    
    def draw_coordination_status(self):
        """Draw coordination activity section"""
        print(f"{Colors.BOLD}{Colors.BRIGHT_MAGENTA}Coordination Activity{Colors.RESET}")
        print("━" * 30)
        
        # Total coordination stats
        print(f"Total Messages: {Colors.BRIGHT_CYAN}{self.total_messages}{Colors.RESET}")
        print(f"Total Tasks: {Colors.BRIGHT_CYAN}{self.total_tasks}{Colors.RESET}")
        print(f"Active Sessions: {Colors.BRIGHT_GREEN}{len(self.coordination_sessions)}{Colors.RESET}")
        
        # Recent task activity
        if self.task_history:
            recent_tasks = len([t for t in self.task_history if 
                              datetime.now() - t['timestamp'] < timedelta(minutes=5)])
            print(f"Recent Tasks (5m): {Colors.BRIGHT_YELLOW}{recent_tasks}{Colors.RESET}")
        
        print()
    
    def draw_recent_activity(self):
        """Draw recent activity log"""
        print(f"{Colors.BOLD}{Colors.BRIGHT_WHITE}Recent Activity{Colors.RESET}")
        print("━" * 30)
        
        # Combine and sort recent activities
        all_activities = []
        
        # Add availability updates
        for activity in list(self.availability_history)[-5:]:
            all_activities.append({
                'time': activity['timestamp'],
                'type': 'AVAIL',
                'message': f"Agent {activity['agent_id'][:8]}... status: {activity['status']}",
                'color': Colors.GREEN
            })
        
        # Add task activities  
        for activity in list(self.task_history)[-3:]:
            all_activities.append({
                'time': activity['timestamp'],
                'type': 'TASK',
                'message': activity['activity'][:50] + "..." if len(activity['activity']) > 50 else activity['activity'],
                'color': Colors.YELLOW
            })
        
        # Add errors
        for error in list(self.error_history)[-3:]:
            all_activities.append({
                'time': error['timestamp'],
                'type': 'ERROR',
                'message': error['error'][:50] + "..." if len(error['error']) > 50 else error['error'],
                'color': Colors.RED
            })
        
        # Sort by time and show most recent
        all_activities.sort(key=lambda x: x['time'], reverse=True)
        
        for activity in all_activities[:8]:  # Show last 8 activities
            time_str = activity['time'].strftime("%H:%M:%S")
            type_str = f"[{activity['type']}]".ljust(7)
            print(f"{Colors.DIM}{time_str}{Colors.RESET} {activity['color']}{type_str}{Colors.RESET} {activity['message']}")
        
        print()
    
    def draw_mini_graph(self, data: List[int], label: str):
        """Draw a simple ASCII graph"""
        if not data or len(data) < 2:
            return
        
        max_val = max(data) if data else 1
        min_val = min(data) if data else 0
        range_val = max_val - min_val if max_val != min_val else 1
        
        # Normalize to 0-10 scale for display
        normalized = [int(((val - min_val) / range_val) * 10) for val in data]
        
        # Draw graph
        graph_chars = ['▁', '▂', '▃', '▄', '▅', '▆', '▇', '█']
        graph_line = ""
        
        for val in normalized:
            if val == 0:
                graph_line += "▁"
            elif val >= len(graph_chars):
                graph_line += "█"
            else:
                graph_line += graph_chars[val]
        
        print(f"{Colors.CYAN}{graph_line}{Colors.RESET} ({min_val}-{max_val} {label})")
    
    def draw_footer(self):
        """Draw footer with controls"""
        print("━" * 50)
        print(f"{Colors.DIM}Press Ctrl+C to exit | Refresh rate: {self.refresh_rate}s{Colors.RESET}")
    
    def run(self):
        """Main monitoring loop"""
        try:
            while True:
                self.clear_screen()
                self.update_terminal_size()
                
                # Get fresh data
                status = self.get_bzzz_status()
                if status['logs']:
                    self.parse_logs(status['logs'])
                
                # Draw dashboard
                self.draw_header()
                self.draw_p2p_status()
                self.draw_agent_activity()
                self.draw_coordination_status()
                self.draw_recent_activity()
                self.draw_footer()
                
                # Wait for next refresh
                time.sleep(self.refresh_rate)
                
        except KeyboardInterrupt:
            print(f"\n{Colors.BRIGHT_CYAN}🛑 Bzzz Monitor stopped{Colors.RESET}")
            sys.exit(0)
        except Exception as e:
            print(f"\n{Colors.BRIGHT_RED}❌ Error: {e}{Colors.RESET}")
            sys.exit(1)

def main():
    parser = argparse.ArgumentParser(description='Bzzz P2P Coordination Monitor')
    parser.add_argument('--refresh-rate', type=float, default=1.0,
                       help='Refresh rate in seconds (default: 1.0)')
    parser.add_argument('--no-color', action='store_true',
                       help='Disable colored output')
    
    args = parser.parse_args()
    
    # Disable colors if requested
    if args.no_color:
        for attr in dir(Colors):
            if not attr.startswith('_'):
                setattr(Colors, attr, '')
    
    # Check if bzzz service exists
    try:
        result = subprocess.run(['systemctl', 'status', 'bzzz.service'], 
                              capture_output=True, text=True)
        if result.returncode != 0 and 'not be found' in result.stderr:
            print(f"{Colors.BRIGHT_RED}❌ Bzzz service not found. Is it installed and running?{Colors.RESET}")
            sys.exit(1)
    except Exception as e:
        print(f"{Colors.BRIGHT_RED}❌ Error checking bzzz service: {e}{Colors.RESET}")
        sys.exit(1)
    
    print(f"{Colors.BRIGHT_CYAN}🚀 Starting Bzzz Monitor...{Colors.RESET}")
    time.sleep(1)
    
    monitor = BzzzMonitor(refresh_rate=args.refresh_rate)
    monitor.run()

if __name__ == '__main__':
    main()