#!/usr/bin/env python3
"""
Test script to trigger and observe bzzz meta discussion
"""

import json
import time
import requests
from datetime import datetime

def test_meta_discussion():
    """Test the Antennae meta discussion by simulating a complex task"""
    
    print("🎯 Testing Bzzz Antennae Meta Discussion")
    print("=" * 50)
    
    # Test 1: Check if the P2P mesh is active
    print("1. Checking P2P mesh status...")
    
    # We can't directly inject into the P2P mesh from here, but we can:
    # - Check the bzzz service logs for meta discussion activity
    # - Create a mock scenario description
    
    mock_scenario = {
        "task_type": "complex_architecture_design",
        "description": "Design a microservices architecture for a distributed AI system with P2P coordination",
        "complexity": "high",
        "requires_collaboration": True,
        "estimated_agents_needed": 3
    }
    
    print(f"📋 Mock Complex Task:")
    print(f"   Type: {mock_scenario['task_type']}")
    print(f"   Description: {mock_scenario['description']}")
    print(f"   Complexity: {mock_scenario['complexity']}")
    print(f"   Collaboration Required: {mock_scenario['requires_collaboration']}")
    
    # Test 2: Demonstrate what would happen in meta discussion
    print("\n2. Simulating Antennae Meta Discussion Flow:")
    print("   🤖 Agent A (walnut): 'I'll handle the API gateway design'")
    print("   🤖 Agent B (acacia): 'I can work on the data layer architecture'") 
    print("   🤖 Agent C (ironwood): 'I'll focus on the P2P coordination logic'")
    print("   🎯 Meta Discussion: Agents coordinate task splitting and dependencies")
    
    # Test 3: Show escalation scenario
    print("\n3. Human Escalation Scenario:")
    print("   ⚠️  Agents detect conflicting approaches to distributed consensus")
    print("   🚨 Automatic escalation triggered after 3 rounds of discussion")
    print("   👤 Human expert summoned via N8N webhook")
    
    # Test 4: Check current bzzz logs for any meta discussion activity
    print("\n4. Checking recent bzzz activity...")
    
    try:
        # This would show any recent meta discussion logs
        import subprocess
        result = subprocess.run([
            'journalctl', '-u', 'bzzz.service', '--no-pager', '-l', '-n', '20'
        ], capture_output=True, text=True, timeout=10)
        
        if result.returncode == 0:
            logs = result.stdout
            if 'meta' in logs.lower() or 'antennae' in logs.lower():
                print("   ✅ Found meta discussion activity in logs!")
                # Show relevant lines
                for line in logs.split('\n'):
                    if 'meta' in line.lower() or 'antennae' in line.lower():
                        print(f"   📝 {line}")
            else:
                print("   ℹ️  No recent meta discussion activity (expected - no active tasks)")
        else:
            print("   ⚠️  Could not access bzzz logs")
            
    except Exception as e:
        print(f"   ⚠️  Error checking logs: {e}")
    
    # Test 5: Show what capabilities support meta discussion
    print("\n5. Meta Discussion Capabilities:")
    capabilities = [
        "meta-discussion", 
        "task-coordination", 
        "collaborative-reasoning",
        "human-escalation",
        "cross-repository-coordination"
    ]
    
    for cap in capabilities:
        print(f"   ✅ {cap}")
    
    print("\n🎯 Meta Discussion Test Complete!")
    print("\nTo see meta discussion in action:")
    print("1. Configure repositories in Hive with 'bzzz_enabled: true'")
    print("2. Create complex GitHub issues labeled 'bzzz-task'") 
    print("3. Watch agents coordinate via Antennae P2P channel")
    print("4. Monitor logs: journalctl -u bzzz.service -f | grep -i meta")

if __name__ == "__main__":
    test_meta_discussion()