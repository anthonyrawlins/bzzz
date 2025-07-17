# Bzzz System Architecture & Flow

This document contains diagrams to visualize the architecture and data flows of the Bzzz distributed task coordination system.

---

## 1. Component Architecture Diagram

This diagram shows the main components of the Bzzz ecosystem and their relationships. It illustrates the static structure of the system, including internal modules, external dependencies, and P2P connections.

```mermaid
graph TD
    subgraph "External Systems"
        GitHub[(GitHub Repositories)] -- "Tasks (Issues/PRs)" --> BzzzAgent
        HiveAPI[Hive REST API] -- "Repo Lists & Status Updates" --> BzzzAgent
        N8N([N8N Webhooks])
        Ollama[Ollama API]
    end

    subgraph "Bzzz Agent Node"
        BzzzAgent[Bzzz Agent]
        BzzzAgent -- "Manages" --> P2P
        BzzzAgent -- "Uses" --> Integration
        BzzzAgent -- "Uses" --> Executor
        BzzzAgent -- "Uses" --> Logging

        P2P(P2P/PubSub Layer) -- "Discovers Peers" --> Discovery
        P2P -- "Communicates via" --> Antennae

        Integration(GitHub Integration) -- "Polls for Tasks" --> HiveAPI
        Integration -- "Claims Tasks" --> GitHub

        Executor(Task Executor) -- "Runs Commands In" --> Sandbox
        Executor -- "Gets Next Command From" --> Reasoning

        Reasoning(Reasoning Module) -- "Sends Prompts To" --> Ollama

        Sandbox(Docker Sandbox) -- "Isolated Environment"

        Logging(Hypercore Logging) -- "Creates Audit Trail"

        Discovery(mDNS Discovery)
    end

    BzzzAgent -- "P2P Comms" --> OtherAgent[Other Bzzz Agent]
    OtherAgent -- "P2P Comms" --> BzzzAgent
    Executor -- "Escalates To" --> N8N

    classDef internal fill:#D6EAF8,stroke:#2E86C1,stroke-width:2px;
    class BzzzAgent,P2P,Integration,Executor,Reasoning,Sandbox,Logging,Discovery internal

    classDef external fill:#E8DAEF,stroke:#8E44AD,stroke-width:2px;
    class GitHub,HiveAPI,N8N,Ollama external
end
```

---

## 2. Task Execution Flowchart

This flowchart illustrates the dynamic lifecycle of a single task, from the moment it's available to its final completion and pull request creation.

```mermaid
flowchart TD
    A[Start: Unassigned Task on GitHub] --> B{Bzzz Agent Polls Hive API};
    B --> C{Discovers Active Repositories};
    C --> D{Polls Repos for Suitable Tasks};
    D --> E{Task Found?};
    E -- No --> B;
    E -- Yes --> F[Agent Claims Task via GitHub API];
    F --> G[Report Claim to Hive API];
    G --> H[Announce Claim on P2P PubSub];

    subgraph "Task Execution Loop"
        I[Create Docker Sandbox] --> J[Clone Repository];
        J --> K{Generate Next Command via Reasoning/Ollama};
        K --> L{Is Task Complete?};
        L -- No --> M[Execute Command in Sandbox];
        M --> N[Feed Output Back to Reasoning];
        N --> K;
    end

    H --> I;
    L -- Yes --> O[Create Branch & Commit Changes];
    O --> P[Push Branch to GitHub];
    P --> Q[Create Pull Request];
    Q --> R[Report Completion to Hive API];
    R --> S[Announce Completion on PubSub];
    S --> T[Destroy Docker Sandbox];
    T --> Z[End];

    subgraph "Meta-Discussion (Antennae)"
        direction LR
        MD1{Agent Proposes Plan} -- PubSub --> MD2[Other Agents Review];
        MD2 -- Feedback --> MD1;
        MD1 -- "Stuck?" --> MD3{Escalate to N8N};
    end

    H -.-> MD1;
    K -- "Needs Help" --> MD1;
end
```
