#!/bin/bash

# Intensive coordination test to generate lots of dashboard activity
# This creates rapid-fire coordination scenarios for monitoring

LOG_DIR="/tmp/bzzz_logs"
TEST_LOG="$LOG_DIR/intensive_test_$(date +%Y%m%d_%H%M%S).log"

mkdir -p "$LOG_DIR"

echo "🚀 Starting Intensive Coordination Test"
echo "======================================"
echo "This will generate rapid coordination activity for dashboard monitoring"
echo "Test Log: $TEST_LOG"
echo ""

# Function to log test events
log_test() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local event="$1"
    echo "[$timestamp] $event" | tee -a "$TEST_LOG"
}

# Function to simulate rapid task announcements
simulate_task_burst() {
    local scenario="$1"
    local count="$2"
    
    log_test "BURST_START: $scenario - announcing $count tasks rapidly"
    
    for i in $(seq 1 $count); do
        log_test "TASK_ANNOUNCE: repo-$i/task-$i - $scenario scenario task $i"
        sleep 0.5
    done
    
    log_test "BURST_COMPLETE: $scenario burst finished"
}

# Function to simulate agent coordination chatter
simulate_agent_chatter() {
    local duration="$1"
    local end_time=$(($(date +%s) + duration))
    
    log_test "CHATTER_START: Simulating agent coordination discussion for ${duration}s"
    
    local agent_responses=(
        "I can handle this task"
        "This conflicts with my current work"
        "Need clarification on requirements"
        "Dependencies detected with repo-X"
        "Proposing different execution order"
        "Ready to start immediately"
        "This requires security review first"
        "API contract needed before implementation"
        "Coordination with team required"
        "Escalating to human review"
    )
    
    local agents=("walnut-agent" "acacia-agent" "ironwood-agent" "test-agent-1" "test-agent-2")
    
    while [ $(date +%s) -lt $end_time ]; do
        local agent=${agents[$((RANDOM % ${#agents[@]}))]}
        local response=${agent_responses[$((RANDOM % ${#agent_responses[@]}))]}
        
        log_test "AGENT_RESPONSE: $agent: $response"
        sleep $((1 + RANDOM % 3))  # Random 1-3 second delays
    done
    
    log_test "CHATTER_COMPLETE: Agent discussion simulation finished"
}

# Function to simulate coordination session lifecycle
simulate_coordination_session() {
    local session_id="coord_$(date +%s)_$RANDOM"
    local repos=("hive" "bzzz" "distributed-ai-dev" "n8n-workflows" "monitoring-tools")
    local selected_repos=(${repos[@]:0:$((2 + RANDOM % 3))})  # 2-4 repos
    
    log_test "SESSION_START: $session_id with repos: ${selected_repos[*]}"
    
    # Dependency analysis phase
    sleep 1
    log_test "SESSION_ANALYZE: $session_id - analyzing cross-repository dependencies"
    
    sleep 2
    log_test "SESSION_DEPS: $session_id - detected $((1 + RANDOM % 4)) dependencies"
    
    # Agent coordination phase
    sleep 1
    log_test "SESSION_COORD: $session_id - agents proposing execution plan"
    
    sleep 2
    local outcome=$((RANDOM % 4))
    case $outcome in
        0|1) 
            log_test "SESSION_SUCCESS: $session_id - consensus reached, plan approved"
            ;;
        2)
            log_test "SESSION_ESCALATE: $session_id - escalated to human review"
            ;;
        3)
            log_test "SESSION_TIMEOUT: $session_id - coordination timeout, retrying"
            ;;
    esac
    
    log_test "SESSION_COMPLETE: $session_id finished"
}

# Function to simulate error scenarios
simulate_error_scenarios() {
    local errors=(
        "Failed to connect to repository API"
        "GitHub rate limit exceeded"
        "Task dependency cycle detected"
        "Agent coordination timeout"
        "Invalid task specification"
        "Network partition detected"
        "Consensus algorithm failure"
        "Authentication token expired"
    )
    
    for error in "${errors[@]}"; do
        log_test "ERROR_SIM: $error"
        sleep 2
    done
}

# Main test execution
main() {
    log_test "TEST_START: Intensive coordination test beginning"
    
    echo "🎯 Phase 1: Rapid Task Announcements (30 seconds)"
    simulate_task_burst "Cross-Repository API Integration" 8 &
    sleep 15
    simulate_task_burst "Security-First Development" 6 &
    
    echo ""
    echo "🤖 Phase 2: Agent Coordination Chatter (45 seconds)"
    simulate_agent_chatter 45 &
    
    echo ""
    echo "🔄 Phase 3: Multiple Coordination Sessions (60 seconds)"
    for i in {1..5}; do
        simulate_coordination_session &
        sleep 12
    done
    
    echo ""
    echo "❌ Phase 4: Error Scenario Simulation (20 seconds)" 
    simulate_error_scenarios &
    
    echo ""
    echo "⚡ Phase 5: High-Intensity Burst (30 seconds)"
    # Rapid-fire everything
    for i in {1..3}; do
        simulate_coordination_session &
        sleep 3
        simulate_task_burst "Parallel-Development-Conflict" 4 &
        sleep 7
    done
    
    # Wait for background processes
    wait
    
    log_test "TEST_COMPLETE: Intensive coordination test finished"
    
    echo ""
    echo "📊 TEST SUMMARY"
    echo "==============="
    echo "Total Events: $(grep -c '\[.*\]' "$TEST_LOG")"
    echo "Task Announcements: $(grep -c 'TASK_ANNOUNCE' "$TEST_LOG")"
    echo "Agent Responses: $(grep -c 'AGENT_RESPONSE' "$TEST_LOG")"
    echo "Coordination Sessions: $(grep -c 'SESSION_START' "$TEST_LOG")"
    echo "Simulated Errors: $(grep -c 'ERROR_SIM' "$TEST_LOG")"
    echo ""
    echo "🎯 Watch your dashboard for all this activity!"
    echo "📝 Detailed log: $TEST_LOG"
}

# Trap Ctrl+C
trap 'echo ""; echo "🛑 Test interrupted"; exit 0' INT

# Run the intensive test
main