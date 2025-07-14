#!/bin/bash

# Script to trigger coordination activity with mock API data
# This simulates task updates to cause real bzzz coordination

MOCK_API="http://localhost:5000"

echo "🎯 Triggering Mock Coordination Test"
echo "===================================="
echo "This will cause real bzzz agents to coordinate on fake tasks"
echo ""

# Function to simulate task claim attempts
simulate_task_claims() {
    echo "📋 Simulating task claim attempts..."
    
    # Try to claim tasks from different projects
    for project_id in 1 2 3; do
        for task_num in 15 23 8; do
            echo "🎯 Agent attempting to claim project $project_id task $task_num"
            
            curl -s -X POST "$MOCK_API/api/bzzz/projects/$project_id/claim" \
                -H "Content-Type: application/json" \
                -d "{\"task_number\": $task_num, \"agent_id\": \"test-agent-$project_id\"}" | jq .
            
            sleep 2
        done
    done
}

# Function to simulate task status updates
simulate_task_updates() {
    echo ""
    echo "📊 Simulating task status updates..."
    
    # Update task statuses to trigger coordination
    curl -s -X PUT "$MOCK_API/api/bzzz/projects/1/status" \
        -H "Content-Type: application/json" \
        -d '{"task_number": 15, "status": "in_progress", "metadata": {"progress": 25}}' | jq .
    
    sleep 3
    
    curl -s -X PUT "$MOCK_API/api/bzzz/projects/2/status" \
        -H "Content-Type: application/json" \
        -d '{"task_number": 23, "status": "completed", "metadata": {"completion_time": "2025-01-14T12:00:00Z"}}' | jq .
    
    sleep 3
    
    curl -s -X PUT "$MOCK_API/api/bzzz/projects/3/status" \
        -H "Content-Type: application/json" \
        -d '{"task_number": 8, "status": "escalated", "metadata": {"reason": "dependency_conflict"}}' | jq .
}

# Function to add urgent tasks
add_urgent_tasks() {
    echo ""
    echo "🚨 Adding urgent tasks to trigger immediate coordination..."
    
    # The mock API has background task generation, but we can trigger it manually
    # by checking repositories multiple times rapidly
    for i in {1..5}; do
        echo "🔄 Repository refresh $i/5"
        curl -s "$MOCK_API/api/bzzz/active-repos" > /dev/null
        curl -s "$MOCK_API/api/bzzz/projects/1/tasks" > /dev/null
        curl -s "$MOCK_API/api/bzzz/projects/2/tasks" > /dev/null
        sleep 1
    done
}

# Function to check bzzz response
check_bzzz_activity() {
    echo ""
    echo "📡 Checking recent bzzz activity..."
    
    # Check last 30 seconds of bzzz logs for API calls
    echo "Recent bzzz log entries:"
    journalctl -u bzzz.service --since "30 seconds ago" -n 10 | grep -E "(API|repository|task|coordination)" || echo "No recent coordination activity"
}

# Main execution
main() {
    echo "🔍 Testing mock API connectivity..."
    curl -s "$MOCK_API/health" | jq .
    
    echo ""
    echo "📋 Current active repositories:"
    curl -s "$MOCK_API/api/bzzz/active-repos" | jq .repositories[].name
    
    echo ""
    echo "🎯 Phase 1: Task Claims"
    simulate_task_claims
    
    echo ""
    echo "📊 Phase 2: Status Updates"  
    simulate_task_updates
    
    echo ""
    echo "🚨 Phase 3: Urgent Tasks"
    add_urgent_tasks
    
    echo ""
    echo "📡 Phase 4: Check Results"
    check_bzzz_activity
    
    echo ""
    echo "✅ Mock coordination test complete!"
    echo ""
    echo "🎯 Watch your monitoring dashboard for:"
    echo "   - Task claim attempts"
    echo "   - Status update processing"
    echo "   - Coordination session activity"
    echo "   - Agent availability changes"
    echo ""
    echo "📝 Check mock API server output for request logs"
}

# Run the test
main