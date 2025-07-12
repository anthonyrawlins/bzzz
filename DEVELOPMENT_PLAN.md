# Project Bzzz: Decentralized Task Execution Network - Development Plan

## 1. Overview & Vision

This document outlines the development plan for **Project Bzzz**, a decentralized task execution network designed to enhance the existing **Hive Cluster**.

The vision is to evolve from a centrally coordinated system to a resilient, peer-to-peer (P2P) mesh of autonomous agents. This architecture eliminates single points of failure, improves scalability, and allows for dynamic, collaborative task resolution. Bzzz will complement the existing N8N orchestration layer, acting as a powerful, self-organizing execution fabric.

---

## 2. Core Architecture

The system is built on three key pillars: decentralized networking, GitHub-native task management, and verifiable, distributed logging.

| Component | Technology | Purpose |
| :--- | :--- | :--- |
| **Networking** | **libp2p** | For peer discovery (mDNS, DHT), identity, and secure P2P communication. |
| **Task Management** | **GitHub Issues** | The single source of truth for task definition, allocation, and tracking. |
| **Messaging** | **libp2p Pub/Sub** | For broadcasting capabilities and coordinating collaborative help requests. |
| **Logging** | **Hypercore Protocol** | For creating a tamper-proof, decentralized, and replicable logging system for debugging. |

---

## 3. Architectural Refinements & Key Features

Based on our analysis, the following refinements will be adopted:

### 3.1. Task Allocation via GitHub Assignment

To prevent race conditions and simplify logic, we will use GitHub's native issue assignment mechanism as an atomic lock. The `task_claim` pub/sub topic is no longer needed.

**Workflow:**
1.  A `bzzz-agent` discovers a new, *unassigned* issue in the target repository.
2.  The agent immediately attempts to **assign itself** to the issue via the GitHub API.
3.  **Success:** If the assignment succeeds, the agent has exclusive ownership of the task and begins execution.
4.  **Failure:** If the assignment fails (because another agent was faster), the agent logs the contention and looks for another task.

### 3.2. Collaborative Task Execution with Hop Limit

The `task_help_request` feature enables agents to collaborate on complex tasks. To prevent infinite request loops and network flooding, we will implement a **hop limit**.

-   **Hop Limit:** A `task_help_request` will be discarded after being forwarded **3 times**.
-   If a task cannot be completed after 3 help requests, it will be marked as "failed," and a comment will be added to the GitHub issue for manual review.

### 3.3. Decentralized Logging with Hypercore

To solve the challenge of debugging a distributed system, each agent will manage its own secure, append-only log stream using the Hypercore Protocol.

-   **Log Creation:** Each agent generates a `hypercore` and broadcasts its public key via the `capabilities` message.
-   **Log Replication:** Any other agent (or a dedicated monitoring node) can use this key to replicate the log stream in real-time or after the fact.
-   **Benefits:** This creates a verifiable and resilient audit trail for every agent's actions, which is invaluable for debugging without relying on a centralized logging server.

---

## 4. Integration with the Hive Ecosystem

Bzzz is designed to integrate seamlessly with the existing cluster infrastructure.

### 4.1. Deployment Strategy: Docker + Host Networking (PREFERRED APPROACH)

Based on comprehensive analysis of the existing Hive infrastructure and Bzzz's P2P requirements, we will use a **hybrid deployment approach** that combines Docker containerization with host networking:

```yaml
# Docker Compose configuration for bzzz-agent
services:
  bzzz-agent:
    image: bzzz-agent:latest
    network_mode: "host"  # Direct host network access for P2P
    volumes:
      - ./data:/app/data
      - /var/run/docker.sock:/var/run/docker.sock  # Docker API access
    environment:
      - NODE_ID=${HOSTNAME}
      - GITHUB_TOKEN_FILE=/run/secrets/github-token
    secrets:
      - github-token
    restart: unless-stopped
    deploy:
      placement:
        constraints:
          - node.role == worker  # Deploy on all worker nodes
```

**Rationale for Docker + Host Networking:**
- ✅ **P2P Networking Advantages**: Direct access to host networking enables efficient mDNS discovery, NAT traversal, and lower latency communication
- ✅ **Infrastructure Consistency**: Maintains Docker Swarm deployment patterns and existing operational procedures
- ✅ **Resource Efficiency**: Eliminates Docker overlay network overhead for P2P communication
- ✅ **Best of Both Worlds**: Container portability and management with native network performance

### 4.2. Cluster Integration Points

-   **Phased Rollout:** Deploy `bzzz-agent` containers across all cluster nodes (ACACIA, WALNUT, IRONWOOD, ROSEWOOD, FORSTEINET) using Docker Swarm
-   **Network Architecture**: Leverages existing 192.168.1.0/24 LAN for P2P mesh communication
-   **Resource Coordination**: Agents discover and utilize existing Ollama endpoints (port 11434) and CLI tools
-   **Storage Integration**: Uses NFS shares (/rust/containers/) for shared configuration and Hypercore log storage

### 4.3. Integration with Existing Services

-   **N8N as a Task Initiator:** High-level workflows in N8N will now terminate by creating a detailed GitHub Issue. This action triggers the Bzzz mesh, which handles the execution and reports back by creating a Pull Request.
-   **Hive Coexistence**: Bzzz will run alongside existing Hive services on different ports, allowing gradual migration of workloads
-   **The "Mesh Visualizer":** A dedicated monitoring dashboard will be created. It will:
    1.  Subscribe to the `capabilities` pub/sub topic to visualize the live network topology.
    2.  Replicate and display the Hypercore log streams from all active agents.
    3.  Integrate with existing Grafana dashboards for unified monitoring

---

## 5. Security Strategy

-   **GitHub Token Management:** Agents will use short-lived, fine-grained Personal Access Tokens. These tokens will be stored securely in **HashiCorp Vault** or a similar secrets management tool, and retrieved by the agent at runtime.
-   **Network Security:** All peer-to-peer communication is automatically **encrypted end-to-end** by `libp2p`.

---

## 6. Recommended Tech Stack

| Category | Recommendation | Notes |
| :--- | :--- | :--- |
| **Language** | **Go** or **Rust** | Strongly recommended for performance, concurrency, and system-level programming. |
| **Networking** | `go-libp2p` / `rust-libp2p` | The official and most mature implementations. |
| **Logging** | `hypercore-go` / `hypercore-rs` | Libraries for implementing the Hypercore Protocol. |
| **GitHub API** | `go-github` / `octokit.rs` | Official and community-maintained clients for interacting with GitHub. |

---

## 7. Development Milestones

This 8-week plan incorporates the refined architecture.

| Week | Deliverables | Key Features |
| :--- | :--- | :--- |
| **1** | **P2P Foundation & Logging** | Setup libp2p peer discovery and establish a **Hypercore log stream** for each agent. |
| **2** | **Capability Broadcasting** | Implement `capability_detector` and broadcast agent status via pub/sub. |
| **3** | **GitHub Task Claiming** | Ingest issues from GitHub and implement the **assignment-based task claiming** logic. |
| **4** | **Core Task Execution** | Integrate local CLIs (Ollama, etc.) to perform basic tasks based on issue content. |
| **5** | **GitHub Result Workflow** | Implement logic to create Pull Requests or follow-up issues upon task completion. |
| **6** | **Collaborative Help** | Implement the `task_help_request` and `task_help_response` flow with the **hop limit**. |
| **7** | **Monitoring & Visualization** | Build the first version of the **Mesh Visualizer** dashboard to display agent status and logs. |
| **8** | **Deployment & Testing** | Package the agent as a Docker container with host networking, write Docker Swarm deployment guide, and conduct end-to-end testing across cluster nodes. |

---

## 8. Potential Risks & Mitigation

-   **Network Partitions ("Split-Brain"):**
    -   **Risk:** A network partition could lead to two separate meshes trying to work on the same task.
    -   **Mitigation:** Using GitHub's issue assignment as the atomic lock effectively solves this. The first agent to successfully claim the issue wins, regardless of network state.
-   **Dependency on GitHub:**
    -   **Risk:** The system's ability to acquire new tasks is dependent on the availability of the GitHub API.
    -   **Mitigation:** This is an accepted trade-off for gaining a robust, native task management platform. Agents can be designed to continue working on already-claimed tasks during a GitHub outage.
-   **Debugging Complexity:**
    -   **Risk:** Debugging distributed systems remains challenging.
    -   **Mitigation:** The Hypercore-based logging system provides a powerful and verifiable audit trail, which is a significant step towards mitigating this complexity. The Mesh Visualizer will also be a critical tool.
-   **Docker Host Networking Security:**
    -   **Risk:** Host networking mode exposes containers directly to the host network, reducing isolation.
    -   **Mitigation:** 
        - Implement strict firewall rules on each node
        - Use libp2p's built-in encryption for all P2P communication
        - Run containers with restricted user privileges (non-root)
        - Regular security audits of exposed ports and services

---

## 9. Migration Strategy from Hive

### 9.1. Gradual Transition Plan

1. **Phase 1: Parallel Deployment** (Weeks 1-2)
   - Deploy Bzzz agents alongside existing Hive infrastructure
   - Use different port ranges to avoid conflicts
   - Monitor resource usage and network performance

2. **Phase 2: Simple Task Migration** (Weeks 3-4)
   - Route basic code generation tasks through GitHub issues → Bzzz
   - Keep complex multi-agent workflows in existing Hive + n8n
   - Compare performance metrics between systems

3. **Phase 3: Workflow Integration** (Weeks 5-6)
   - Modify n8n workflows to create GitHub issues as final step
   - Implement Bzzz → Hive result reporting for hybrid workflows
   - Test end-to-end task lifecycle

4. **Phase 4: Full Migration** (Weeks 7-8)
   - Migrate majority of workloads to Bzzz mesh
   - Retain Hive for monitoring and dashboard functionality
   - Plan eventual deprecation of centralized coordinator

### 9.2. Compatibility Layer

-   **API Bridge**: Maintain existing Hive API endpoints that proxy to Bzzz mesh
-   **Data Migration**: Export task history and agent configurations from PostgreSQL
-   **Monitoring Continuity**: Integrate Bzzz metrics into existing Grafana dashboards
