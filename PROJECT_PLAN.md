🐝 Project: Bzzz — P2P Task Coordination System

## 🔧 Architecture Overview (libp2p + pubsub + JSON)

This system will compliment and partially replace elements of the Hive Software System. This is intended to be a replacement for the multitude of MCP, and API calls to the ollama and gemini-cli agents over port 11434 etc. By replacing the master/slave paradigm with a mesh network we allow each node to trigger workflows or respond to calls for work as availability dictates rather than being stuck in endless timeouts awaiting responses. We also eliminate the central coordinator as a single point of failure.

### 📂 Components

#### 1. **Peer Node**

Each machine runs a P2P agent that:

- Connects to other peers via libp2p
- Subscribes to pubsub topics
- Periodically broadcasts status/capabilities
- Receives and executes tasks
- Publishes task results as GitHub pull requests or issues
- Can request assistance from other peers
- Monitors a GitHub repository for new issues (task source)

Each node uses a dedicated GitHub account with:
- A personal access token (fine-scoped to repo/PRs)
- A configured `.gitconfig` for commit identity

#### 2. **libp2p Network**

- All peers discover each other using mDNS, Bootstrap peers, or DHT
- Peer identity is cryptographic (libp2p peer ID)
- Communication is encrypted end-to-end

#### 3. **GitHub Integration**

- Tasks are sourced from GitHub Issues in a designated repository
- Nodes will claim and respond to tasks by:
  - Forking the repository (once)
  - Creating a working branch
  - Making changes to files as instructed by task input
  - Committing changes using their GitHub identity
  - Creating a pull request or additional GitHub issues
  - Publishing final result as a PR, issue(s), or failure report

#### 4. **PubSub Topics**

| Topic             | Direction        | Purpose                                     |
|------------------|------------------|---------------------------------------------|
| `capabilities`    | Peer → All Peers | Broadcast available models, status          |
| `task_broadcast`  | Peer → All Peers | Publish a GitHub issue as task              |
| `task_claim`      | Peer → All Peers | Claim responsibility for a task             |
| `task_result`     | Peer → All Peers | Share PR, issue, or failure result          |
| `presence_ping`   | Peer → All Peers | Lightweight presence signal                 |
| `task_help_request` | Peer → All Peers | Request assistance for a task               |
| `task_help_response`| Peer → All Peers | Offer help or handle sub-task               |

### 📊 Data Flow Diagram
```
+------------------+       libp2p        +------------------+
|     Peer A       |<------------------->|     Peer B       |
|                  |<------------------->|                  |
| - Publishes:     |                    | - Publishes:     |
|   capabilities   |                    |   task_result    |
|   task_broadcast |                    |   capabilities   |
|   help_request   |                    |   help_response  |
| - Subscribes to: |                    | - Subscribes to: |
|   task_result    |                    |   task_broadcast |
|   help_request   |                    |   help_request   |
+------------------+                    +------------------+
        ^                                        ^
        |                                        |
        |                                        |
        +----------------------+-----------------+
                               |
                               v
                         +------------------+
                         |     Peer C       |
                         +------------------+
```

### 📂 Sample JSON Messages

#### `capabilities`
```json
{
  "type": "capabilities",
  "node_id": "pi-node-1",
  "cpu": 43.5,
  "gpu": 2.3,
  "models": ["llama3", "mistral"],
  "installed": ["ollama", "gemini-cli"],
  "status": "idle",
  "timestamp": "2025-07-12T01:23:45Z"
}
```

#### `task_broadcast`
```json
{
  "type": "task",
  "task_id": "#42",
  "repo": "example-org/task-repo",
  "issue_url": "https://github.com/example-org/task-repo/issues/42",
  "model": "ollama",
  "input": "Add unit tests to utils module",
  "params": {"branch_prefix": "task-42-"},
  "timestamp": "2025-07-12T02:00:00Z"
}
```

#### `task_claim`
```json
{
  "type": "task_claim",
  "task_id": "#42",
  "node_id": "pi-node-2",
  "timestamp": "2025-07-12T02:00:03Z"
}
```

#### `task_result`
```json
{
  "type": "task_result",
  "task_id": "#42",
  "node_id": "pi-node-2",
  "result_type": "pull_request",
  "result_url": "https://github.com/example-org/task-repo/pull/97",
  "duration_ms": 15830,
  "timestamp": "2025-07-12T02:10:05Z"
}
```

#### `task_help_request`
```json
{
  "type": "task_help_request",
  "task_id": "#42",
  "from_node": "pi-node-2",
  "reason": "Long-running task or missing capability",
  "requested_capability": "claude-cli",
  "timestamp": "2025-07-12T02:05:00Z"
}
```

#### `task_help_response`
```json
{
  "type": "task_help_response",
  "task_id": "#42",
  "from_node": "pi-node-3",
  "can_help": true,
  "capabilities": ["claude-cli"],
  "eta_seconds": 30,
  "timestamp": "2025-07-12T02:05:02Z"
}
```

---

## 🚀 Development Brief

### 🧱 Tech Stack

- **Language**: Node.js (or Go/Rust)
- **Networking**: libp2p
- **Messaging**: pubsub with JSON
- **Task Execution**: Local CLI (ollama, gemini, claude)
- **System Monitoring**: `os-utils`, `psutil`, `nvidia-smi`
- **Runtime**: systemd services on Linux
- **GitHub Interaction**: `octokit` (Node), Git CLI

### 🛠 Key Modules

#### 1. `peer_agent.js`

- Initializes libp2p node
- Joins pubsub topics
- Periodically publishes capabilities
- Listens for tasks, runs them, and reports PR/results
- Handles help requests and responses

#### 2. `capability_detector.js`

- Detects:
  - CPU/GPU load
  - Installed models (via `ollama list`)
  - Installed CLIs (`which gemini`, `which claude`)

#### 3. `task_executor.js`

- Parses GitHub issue input
- Forks repo (if needed)
- Creates working branch, applies changes
- Commits changes using local Git identity
- Pushes branch and creates pull request or follow-up issues

#### 4. `github_bot.js`

- Authenticates GitHub API client
- Watches for new issues in repo
- Publishes them as `task_broadcast`
- Handles PR/issue creation and error handling

#### 5. `state_manager.js`

- Keeps internal view of network state
- Tracks peers’ capabilities, liveness
- Matches help requests to eligible peers

### 📆 Milestones

| Week | Deliverables                                                 |
| ---- | ------------------------------------------------------------ |
| 1    | libp2p peer bootstrapping + pubsub skeleton                  |
| 2    | JSON messaging spec + capability broadcasting                |
| 3    | GitHub issue ingestion + task broadcast                      |
| 4    | CLI integration with Ollama/Gemini/Claude                    |
| 5    | GitHub PR/issue/failure workflows                            |
| 6    | Help request/response logic, delegation framework            |
| 7    | systemd setup, CLI utilities, and resilience                 |
| 8    | End-to-end testing, GitHub org coordination, deployment guide|

---

Would you like a prototype `task_help_request` matchmaking function or sample test matrix for capability validation?

