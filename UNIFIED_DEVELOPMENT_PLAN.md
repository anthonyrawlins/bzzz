# Project Bzzz & Antennae: Integrated Development Plan

## 1. Unified Vision

This document outlines a unified development plan for **Project Bzzz** and its integrated meta-discussion layer, **Project Antennae**. The vision is to build a decentralized task execution network where autonomous agents can not only **act** but also **reason and collaborate** before acting.

-   **Bzzz** provides the core P2P execution fabric (task claiming, execution, results).
-   **Antennae** provides the collaborative "social brain" (task clarification, debate, knowledge sharing).

By developing them together, we create a system that is both resilient and intelligent.

---

## 2. Core Architecture

The combined architecture remains consistent with the principles of decentralization, leveraging a unified tech stack.

| Component | Technology | Purpose |
| :--- | :--- | :--- |
| **Networking** | **libp2p** | Peer discovery, identity, and secure P2P communication. |
| **Task Management** | **GitHub Issues** | The single source of truth for task definition and atomic allocation via assignment. |
| **Messaging** | **libp2p Pub/Sub** | Used for both `bzzz` (capabilities) and `antennae` (meta-discussion) topics. |
| **Logging** | **Hypercore Protocol** | A single, tamper-proof log stream per agent will store both execution logs (Bzzz) and discussion transcripts (Antennae). |

---

## 3. Key Features & Refinements

### 3.1. Task Lifecycle with Meta-Discussion

The agent's task lifecycle will be enhanced to include a reasoning step:

1.  **Discover & Claim:** An agent discovers an unassigned GitHub issue and claims it by assigning itself.
2.  **Open Meta-Channel:** The agent immediately joins a dedicated pub/sub topic: `bzzz/meta/issue/{id}`.
3.  **Propose Plan:** The agent posts its proposed plan of action to the channel. *e.g., "I will address this by modifying `file.py` and adding a new function `x()`."*
4.  **Listen & Discuss:** The agent waits for a brief "objection period" (e.g., 30 seconds). Other agents can chime in with suggestions, corrections, or questions. This is the core loop of the Antennae layer.
5.  **Execute:** If no major objections are raised, the agent proceeds with its plan.
6.  **Report:** The agent creates a Pull Request. The PR description will include a link to the Hypercore log containing the full transcript of the pre-execution discussion.

### 3.2. Safeguards and Structured Messaging

-   **Combined Safeguards:** Hop limits, participant caps, and TTLs will apply to all meta-discussions to prevent runaway conversations.
-   **Structured Messages:** To improve machine comprehension, `meta_msg` payloads will be structured.

    ```json
    {
      "type": "meta_msg",
      "issue_id": 42,
      "node_id": "bzzz-07",
      "msg_id": "abc123",
      "parent_id": null,
      "hop_count": 1,
      "content": {
        "query_type": "clarification_needed",
        "text": "What is the expected output format?",
        "parameters": { "field": "output_format" }
      }
    }
    ```

### 3.3. Human Escalation Path

-   A dedicated pub/sub topic (`bzzz/meta/escalation`) will be used to flag discussions requiring human intervention.
-   An N8N workflow will monitor this topic and create alerts in a designated Slack channel or project management tool.

---

## 4. Integrated Development Milestones

This 8-week plan merges the development of both projects into a single, cohesive timeline.

| Week | Core Deliverable | Key Features & Integration Points |
| :--- | :--- | :--- |
| **1** | **P2P Foundation & Logging** | Establish the core agent identity and a unified **Hypercore log stream** for both action and discussion events. |
| **2** | **Capability Broadcasting** | Agents broadcast capabilities, including which reasoning models they have available (e.g., `claude-3-opus`). |
| **3** | **GitHub Task Claiming & Channel Creation** | Implement assignment-based task claiming. Upon claim, the agent **creates and subscribes to the meta-discussion channel**. |
| **4** | **Pre-Execution Discussion** | Implement the "propose plan" and "listen for objections" logic. This is the first functional version of the Antennae layer. |
| **5** | **Result Workflow with Logging** | Implement PR creation. The PR body **must link to the Hypercore discussion log**. |
| **6** | **Full Collaborative Help** | Implement the full `task_help_request` and `meta_msg` response flow, respecting all safeguards (hop limits, TTLs). |
| **7** | **Unified Monitoring** | The Mesh Visualizer dashboard will display agent status, execution logs, and **live meta-discussion transcripts**. |
| **8** | **End-to-End Scenario Testing** | Conduct comprehensive tests for combined scenarios: task clarification, collaborative debugging, and successful escalation to a human. |

---

## 5. Conclusion

By integrating Antennae from the outset, we are not just building a distributed task runner; we are building a **distributed reasoning system**. This approach will lead to a more robust, intelligent, and auditable Hive, where agents think and collaborate before they act.
