# Bzzz Antennae Test Suite

This directory contains a comprehensive test suite for the Bzzz antennae coordination system that operates independently of external services like Hive, GitHub, or n8n.

## Components

### 1. Task Simulator (`task_simulator.go`)
- **Purpose**: Generates realistic task scenarios for testing coordination
- **Features**:
  - Mock repositories with cross-dependencies
  - Simulated GitHub issues with bzzz-task labels
  - Coordination scenarios (API integration, security-first, parallel conflicts)
  - Automatic task announcements every 45 seconds
  - Simulated agent responses every 30 seconds

### 2. Antennae Test Suite (`antennae_test.go`)
- **Purpose**: Comprehensive testing of coordination capabilities
- **Test Categories**:
  - Basic task announcement and response
  - Cross-repository dependency detection
  - Multi-repository coordination sessions
  - Conflict resolution between agents
  - Human escalation scenarios
  - Load handling with concurrent sessions

### 3. Test Runner (`cmd/test_runner.go`)
- **Purpose**: Command-line interface for running tests
- **Modes**:
  - `simulator` - Run only the task simulator
  - `testsuite` - Run full coordination tests
  - `interactive` - Interactive testing environment

## Mock Data

### Mock Repositories
1. **hive** - Main coordination platform
   - WebSocket support task (depends on bzzz API)
   - Agent authentication system (security)

2. **bzzz** - P2P coordination system
   - API contract definition (blocks other work)
   - Dependency detection algorithm

3. **distributed-ai-dev** - AI development tools
   - Bzzz integration task (depends on API + auth)

### Coordination Scenarios
1. **Cross-Repository API Integration**
   - Tests coordination when multiple repos implement shared API
   - Verifies proper dependency ordering

2. **Security-First Development**
   - Tests blocking relationships with security requirements
   - Ensures auth work completes before integration

3. **Parallel Development Conflict**
   - Tests conflict resolution when agents work on overlapping features
   - Verifies coordination to prevent conflicts

## Usage

### Build the test runner:
```bash
go build -o test-runner cmd/test_runner.go
```

### Run modes:

#### 1. Full Test Suite (Default)
```bash
./test-runner
# or
./test-runner testsuite
```

#### 2. Task Simulator Only
```bash
./test-runner simulator
```
- Continuously announces mock tasks
- Simulates agent responses
- Runs coordination scenarios
- Useful for manual observation

#### 3. Interactive Mode
```bash
./test-runner interactive
```
Commands available:
- `start` - Start task simulator
- `stop` - Stop task simulator  
- `test` - Run single test
- `status` - Show current status
- `peers` - Show connected peers
- `scenario <name>` - Run specific scenario
- `quit` - Exit

## Test Results

The test suite generates detailed results including:
- **Pass/Fail Status**: Each test's success state
- **Timing Metrics**: Response times and duration
- **Coordination Logs**: Step-by-step coordination activity
- **Quantitative Metrics**: Tasks announced, sessions created, dependencies detected

### Example Output:
```
🧪 Antennae Coordination Test Suite
==================================================

🔬 Running Test 1/6
   📋 Basic Task Announcement
   ✅ PASSED (2.3s)
   Expected: Agents respond to task announcements within 30 seconds
   Actual: Received 2 agent responses

🔬 Running Test 2/6
   🔗 Dependency Detection
   ✅ PASSED (156ms)
   Expected: System detects task dependencies across repositories
   Actual: Detected 3 cross-repository dependencies
```

## Integration with Live System

While the test suite is designed to work independently, it can also be used alongside the live bzzz P2P mesh:

1. **Connect to Live Mesh**: The test runner will discover and connect to existing bzzz nodes (WALNUT, ACACIA, IRONWOOD)

2. **Isolated Test Topics**: Uses separate PubSub topics (`bzzz/test/coordination`, `antennae/test/meta-discussion`) to avoid interfering with production coordination

3. **Real Peer Discovery**: Uses actual mDNS discovery to find peers, testing the full P2P stack

## Benefits

1. **Independent Testing**: No dependencies on external services
2. **Realistic Scenarios**: Based on actual coordination patterns
3. **Comprehensive Coverage**: Tests all aspects of antennae coordination
4. **Quantitative Metrics**: Provides measurable test results
5. **Interactive Development**: Supports manual testing and debugging
6. **Load Testing**: Verifies behavior under concurrent coordination sessions

This test suite enables rapid development and validation of the antennae coordination system without requiring complex external infrastructure.