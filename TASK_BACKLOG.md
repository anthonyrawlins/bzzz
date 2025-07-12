# Bzzz + Antennae Development Task Backlog

Based on the UNIFIED_DEVELOPMENT_PLAN.md, here are the development tasks ready for distribution to the Hive cluster:

## Week 1-2: Foundation Tasks

### Task 1: P2P Networking Foundation 🔧
**Assigned to**: WALNUT (Advanced Coding - starcoder2:15b)
**Priority**: 5 (Critical)
**Objective**: Design and implement core P2P networking foundation for Project Bzzz using libp2p in Go

**Requirements**:
- Use go-libp2p library for mesh networking
- Implement mDNS peer discovery for local network (192.168.1.0/24)
- Create secure encrypted P2P connections with peer identity
- Design pub/sub topics for both task coordination (Bzzz) and meta-discussion (Antennae)
- Prepare for Docker + host networking deployment
- Create modular Go code structure in `/home/tony/AI/projects/Bzzz/`

**Deliverables**:
- `main.go` - Entry point and peer initialization
- `p2p/` - P2P networking module with libp2p integration
- `discovery/` - mDNS peer discovery implementation  
- `pubsub/` - Pub/sub messaging for capability broadcasting
- `go.mod` - Go module definition with dependencies
- `Dockerfile` - Container with host networking support

### Task 2: Distributed Logging System 📊
**Assigned to**: IRONWOOD (Reasoning Analysis - phi4:14b)
**Priority**: 4 (High)
**Dependencies**: Task 1 (P2P Foundation)
**Objective**: Architect and implement Hypercore-based distributed logging system

**Requirements**:
- Design append-only log streams using Hypercore Protocol
- Implement public key broadcasting for log identity
- Create log replication capabilities between peers
- Store both execution logs (Bzzz) and discussion transcripts (Antennae)
- Ensure tamper-proof audit trails for debugging
- Integrate with P2P capability detection module

**Deliverables**:
- `logging/` - Hypercore-based logging module
- `replication/` - Log replication and synchronization
- `audit/` - Tamper-proof audit trail verification
- Documentation on log schema and replication protocol

### Task 3: GitHub Integration Module 📋
**Assigned to**: ACACIA (Code Review/Docs - codellama)
**Priority**: 4 (High)  
**Dependencies**: Task 1 (P2P Foundation)
**Objective**: Implement GitHub integration for atomic task claiming and collaborative workflows

**Requirements**:
- Create atomic issue assignment mechanism (GitHub's native assignment)
- Implement repository forking, branch creation, and commit workflows
- Generate pull requests with discussion transcript links
- Handle task result posting and failure reporting
- Use GitHub API for all interactions
- Include comprehensive error handling and retry logic

**Deliverables**:
- `github/` - GitHub API integration module
- `workflows/` - Repository and branch management
- `tasks/` - Task claiming and result posting
- Integration tests with GitHub API
- Documentation on GitHub workflow process

## Week 3-4: Integration Tasks

### Task 4: Meta-Discussion Implementation 💬
**Assigned to**: IRONWOOD (Reasoning Analysis)
**Priority**: 3 (Medium)
**Dependencies**: Task 1, Task 2
**Objective**: Implement Antennae meta-discussion layer for collaborative reasoning

**Requirements**:
- Create structured messaging for agent collaboration
- Implement "propose plan" and "objection period" logic
- Add hop limits (3 hops) and participant caps for safety
- Design escalation paths to human intervention
- Integrate with Hypercore logging for discussion transcripts

### Task 5: End-to-End Integration 🔄
**Assigned to**: WALNUT (Advanced Coding)
**Priority**: 2 (Normal)
**Dependencies**: All previous tasks
**Objective**: Integrate all components and create working Bzzz+Antennae system

**Requirements**:
- Combine P2P networking, logging, and GitHub integration
- Implement full task lifecycle with meta-discussion
- Create Docker Swarm deployment configuration
- Add monitoring and health checks
- Comprehensive testing across cluster nodes

## Current Status

✅ **Hive Cluster Ready**: 3 agents registered with proper specializations
- walnut: starcoder2:15b (kernel_dev) 
- ironwood: phi4:14b (reasoning)
- acacia: codellama (docs_writer)

✅ **Authentication Working**: Dev user and API access configured

⚠️ **Task Submission**: Need to resolve API endpoint issues for automated task distribution

**Next Steps**: 
1. Fix task creation API endpoint issues
2. Submit tasks to respective agents based on specializations
3. Monitor execution and coordinate between agents
4. Test the collaborative reasoning (Antennae) layer once P2P foundation is complete