# 🐝 Bzzz-Hive Integration TODOs

**Updated**: January 13, 2025  
**Context**: Dynamic Multi-Repository Task Discovery via Hive API

---

## 🎯 **HIGH PRIORITY: Dynamic Repository Management**

### **1. Hive API Client Integration**
- [ ] **Create Hive API client**
  ```go
  // pkg/hive/client.go
  type HiveClient struct {
      BaseURL    string
      APIKey     string
      HTTPClient *http.Client
  }
  
  func (c *HiveClient) GetActiveRepositories() ([]Repository, error)
  func (c *HiveClient) GetProjectTasks(projectID int) ([]Task, error) 
  func (c *HiveClient) ClaimTask(projectID, taskID int, agentID string) error
  func (c *HiveClient) UpdateTaskStatus(projectID, taskID int, status string) error
  ```

### **2. Configuration Management**
- [ ] **Remove hardcoded repository configuration**
  - [ ] Remove repository settings from main.go
  - [ ] Create Hive API endpoint configuration
  - [ ] Add API authentication configuration
  - [ ] Support multiple simultaneous repository polling

- [ ] **Environment/Config file support**
  ```go
  type Config struct {
      HiveAPI struct {
          BaseURL string `yaml:"base_url"`
          APIKey  string `yaml:"api_key"`
      } `yaml:"hive_api"`
      
      Agent struct {
          ID           string   `yaml:"id"`
          Capabilities []string `yaml:"capabilities"`
          PollInterval string   `yaml:"poll_interval"`
      } `yaml:"agent"`
  }
  ```

### **3. Multi-Repository Task Coordination**
- [ ] **Enhance GitHub Integration service**
  ```go
  // github/integration.go modifications
  type Integration struct {
      hiveClient    *hive.HiveClient
      repositories  map[int]*RepositoryClient  // projectID -> GitHub client
      // ... existing fields
  }
  
  func (i *Integration) pollHiveForRepositories() error
  func (i *Integration) syncRepositoryClients() error
  func (i *Integration) aggregateTasksFromAllRepos() ([]*Task, error)
  ```

---

## 🔧 **MEDIUM PRIORITY: Enhanced Task Management**

### **4. Repository-Aware Task Processing**
- [ ] **Extend Task structure**
  ```go
  type Task struct {
      // Existing fields...
      ProjectID    int    `json:"project_id"`
      ProjectName  string `json:"project_name"`
      GitURL       string `json:"git_url"`
      Owner        string `json:"owner"`
      Repository   string `json:"repository"`
      Branch       string `json:"branch"`
  }
  ```

### **5. Intelligent Task Routing**
- [ ] **Project-aware task filtering**
  - [ ] Filter tasks by agent capabilities per project
  - [ ] Consider project-specific requirements
  - [ ] Implement project priority weighting
  - [ ] Add load balancing across projects

### **6. Cross-Repository Coordination**
- [ ] **Enhanced meta-discussion for multi-project coordination**
  ```go
  type ProjectContext struct {
      ProjectID   int
      GitURL      string
      TaskCount   int
      ActiveAgents []string
  }
  
  func (i *Integration) announceProjectStatus(ctx ProjectContext) error
  func (i *Integration) coordinateAcrossProjects() error
  ```

---

## 🚀 **LOW PRIORITY: Advanced Features**

### **7. Project-Specific Configuration**
- [ ] **Per-project agent specialization**
  - [ ] Different capabilities per project type
  - [ ] Project-specific model preferences
  - [ ] Custom escalation rules per project
  - [ ] Project-aware conversation limits

### **8. Enhanced Monitoring & Metrics**
- [ ] **Multi-project performance tracking**
  - [ ] Tasks completed per project
  - [ ] Agent efficiency across projects
  - [ ] Cross-project collaboration metrics
  - [ ] Project-specific escalation rates

### **9. Advanced Task Coordination**
- [ ] **Cross-project dependencies**
  - [ ] Detect related tasks across repositories
  - [ ] Coordinate dependent task execution
  - [ ] Share knowledge between project contexts
  - [ ] Manage resource allocation across projects

---

## 📋 **IMPLEMENTATION PLAN**

### **Phase 1: Core Hive Integration (Week 1)**
1. **Day 1-2**: Create Hive API client and configuration management
2. **Day 3-4**: Modify GitHub integration to use dynamic repositories  
3. **Day 5**: Test with single active project from Hive
4. **Day 6-7**: Multi-repository polling and task aggregation

### **Phase 2: Enhanced Coordination (Week 2)**  
1. **Day 1-3**: Repository-aware task processing and routing
2. **Day 4-5**: Cross-repository meta-discussion enhancements
3. **Day 6-7**: Project-specific escalation and configuration

### **Phase 3: Advanced Features (Week 3)**
1. **Day 1-3**: Performance monitoring and metrics
2. **Day 4-5**: Cross-project dependency management  
3. **Day 6-7**: Production testing and optimization

---

## 🔧 **CODE STRUCTURE CHANGES**

### **New Files to Create:**
```
pkg/
├── hive/
│   ├── client.go          # Hive API client
│   ├── models.go          # Hive data structures
│   └── config.go          # Hive configuration
├── config/
│   ├── config.go          # Configuration management
│   └── defaults.go        # Default configuration
└── repository/
    ├── manager.go         # Multi-repository management
    ├── router.go          # Task routing logic
    └── coordinator.go     # Cross-repository coordination
```

### **Files to Modify:**
```
main.go                    # Remove hardcoded repo config
github/integration.go     # Add Hive client integration
github/client.go          # Support multiple repository configs
pubsub/pubsub.go          # Enhanced project context messaging
```

---

## 📊 **TESTING STRATEGY**

### **Unit Tests**
- [ ] Hive API client functionality
- [ ] Multi-repository configuration loading
- [ ] Task aggregation and routing logic
- [ ] Project-aware filtering algorithms

### **Integration Tests**  
- [ ] End-to-end Hive API communication
- [ ] Multi-repository GitHub integration
- [ ] Cross-project task coordination
- [ ] P2P coordination with project context

### **System Tests**
- [ ] Full workflow: Hive project activation → Bzzz task discovery → coordination
- [ ] Performance under multiple active projects
- [ ] Failure scenarios (Hive API down, GitHub rate limits)
- [ ] Escalation workflows across different projects

---

## ✅ **SUCCESS CRITERIA**

### **Phase 1 Complete When:**
- [ ] Bzzz agents query Hive API for active repositories
- [ ] Agents can discover tasks from multiple GitHub repositories
- [ ] Task claims are reported back to Hive system
- [ ] Configuration is fully dynamic (no hardcoded repositories)

### **Phase 2 Complete When:**
- [ ] Agents coordinate effectively across multiple projects
- [ ] Task routing considers project-specific requirements
- [ ] Meta-discussions include project context
- [ ] Performance metrics track multi-project activity

### **Full Integration Complete When:**
- [ ] System scales to 10+ active projects simultaneously
- [ ] Cross-project coordination is seamless
- [ ] Escalation workflows are project-aware
- [ ] Analytics provide comprehensive project insights

---

## 🔧 **IMMEDIATE NEXT STEPS**

1. **Create Hive API client** (`pkg/hive/client.go`)
2. **Implement configuration management** (`pkg/config/config.go`)
3. **Modify main.go** to use dynamic repository discovery
4. **Test with single Hive project** to validate integration
5. **Extend to multiple repositories** once basic flow works

---

**Next Action**: Implement Hive API client and remove hardcoded repository configuration from main.go.