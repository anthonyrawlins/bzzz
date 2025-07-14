#!/usr/bin/env python3
"""
Mock Hive API Server for Bzzz Testing

This simulates what the real Hive API would provide to bzzz agents:
- Active repositories with bzzz-enabled tasks
- Fake GitHub issues with bzzz-task labels
- Task dependencies and coordination scenarios

The real bzzz agents will consume this fake data and do actual coordination.
"""

import json
import random
import time
from datetime import datetime, timedelta
from flask import Flask, jsonify, request
from threading import Thread

app = Flask(__name__)

# Mock data for repositories and tasks
MOCK_REPOSITORIES = [
    {
        "project_id": 1,
        "name": "hive-coordination-platform",
        "git_url": "https://github.com/mock/hive",
        "owner": "mock-org",
        "repository": "hive",
        "branch": "main",
        "bzzz_enabled": True,
        "ready_to_claim": True,
        "private_repo": False,
        "github_token_required": False
    },
    {
        "project_id": 2,
        "name": "bzzz-p2p-system",
        "git_url": "https://github.com/mock/bzzz",
        "owner": "mock-org", 
        "repository": "bzzz",
        "branch": "main",
        "bzzz_enabled": True,
        "ready_to_claim": True,
        "private_repo": False,
        "github_token_required": False
    },
    {
        "project_id": 3,
        "name": "distributed-ai-development",
        "git_url": "https://github.com/mock/distributed-ai-dev",
        "owner": "mock-org",
        "repository": "distributed-ai-dev", 
        "branch": "main",
        "bzzz_enabled": True,
        "ready_to_claim": True,
        "private_repo": False,
        "github_token_required": False
    },
    {
        "project_id": 4,
        "name": "infrastructure-automation",
        "git_url": "https://github.com/mock/infra-automation",
        "owner": "mock-org",
        "repository": "infra-automation",
        "branch": "main", 
        "bzzz_enabled": True,
        "ready_to_claim": True,
        "private_repo": False,
        "github_token_required": False
    }
]

# Mock tasks with realistic coordination scenarios
MOCK_TASKS = {
    1: [  # hive tasks
        {
            "number": 15,
            "title": "Add WebSocket support for real-time coordination",
            "description": "Implement WebSocket endpoints for real-time agent coordination messages",
            "state": "open",
            "labels": ["bzzz-task", "feature", "realtime", "coordination"],
            "created_at": "2025-01-14T10:00:00Z",
            "updated_at": "2025-01-14T10:30:00Z",
            "html_url": "https://github.com/mock/hive/issues/15",
            "is_claimed": False,
            "assignees": [],
            "task_type": "feature",
            "dependencies": [
                {
                    "repository": "bzzz",
                    "task_number": 23,
                    "dependency_type": "api_contract"
                }
            ]
        },
        {
            "number": 16,
            "title": "Implement agent authentication system",
            "description": "Add secure JWT-based authentication for bzzz agents accessing Hive APIs",
            "state": "open",
            "labels": ["bzzz-task", "security", "auth", "high-priority"],
            "created_at": "2025-01-14T09:30:00Z", 
            "updated_at": "2025-01-14T10:45:00Z",
            "html_url": "https://github.com/mock/hive/issues/16",
            "is_claimed": False,
            "assignees": [],
            "task_type": "security",
            "dependencies": []
        },
        {
            "number": 17,
            "title": "Create coordination metrics dashboard", 
            "description": "Build dashboard showing cross-repository coordination statistics",
            "state": "open",
            "labels": ["bzzz-task", "dashboard", "metrics", "ui"],
            "created_at": "2025-01-14T11:00:00Z",
            "updated_at": "2025-01-14T11:15:00Z", 
            "html_url": "https://github.com/mock/hive/issues/17",
            "is_claimed": False,
            "assignees": [],
            "task_type": "feature",
            "dependencies": [
                {
                    "repository": "bzzz",
                    "task_number": 24,
                    "dependency_type": "api_contract"
                }
            ]
        }
    ],
    2: [  # bzzz tasks
        {
            "number": 23,
            "title": "Define coordination API contract",
            "description": "Standardize API contract for cross-repository coordination messaging",
            "state": "open",
            "labels": ["bzzz-task", "api", "coordination", "blocker"],
            "created_at": "2025-01-14T09:00:00Z",
            "updated_at": "2025-01-14T10:00:00Z",
            "html_url": "https://github.com/mock/bzzz/issues/23", 
            "is_claimed": False,
            "assignees": [],
            "task_type": "api_design",
            "dependencies": []
        },
        {
            "number": 24,
            "title": "Implement dependency detection algorithm",
            "description": "Auto-detect task dependencies across repositories using graph analysis",
            "state": "open",
            "labels": ["bzzz-task", "algorithm", "coordination", "complex"],
            "created_at": "2025-01-14T10:15:00Z",
            "updated_at": "2025-01-14T10:30:00Z",
            "html_url": "https://github.com/mock/bzzz/issues/24",
            "is_claimed": False, 
            "assignees": [],
            "task_type": "feature",
            "dependencies": [
                {
                    "repository": "bzzz",
                    "task_number": 23,
                    "dependency_type": "api_contract"
                }
            ]
        },
        {
            "number": 25,
            "title": "Add consensus algorithm for coordination",
            "description": "Implement distributed consensus for multi-agent task coordination",
            "state": "open",
            "labels": ["bzzz-task", "consensus", "distributed-systems", "hard"],
            "created_at": "2025-01-14T11:30:00Z",
            "updated_at": "2025-01-14T11:45:00Z",
            "html_url": "https://github.com/mock/bzzz/issues/25",
            "is_claimed": False,
            "assignees": [],
            "task_type": "feature", 
            "dependencies": []
        }
    ],
    3: [  # distributed-ai-dev tasks
        {
            "number": 8,
            "title": "Add support for bzzz coordination",
            "description": "Integrate with bzzz P2P coordination system for distributed AI development",
            "state": "open",
            "labels": ["bzzz-task", "integration", "p2p", "ai"],
            "created_at": "2025-01-14T10:45:00Z",
            "updated_at": "2025-01-14T11:00:00Z",
            "html_url": "https://github.com/mock/distributed-ai-dev/issues/8",
            "is_claimed": False,
            "assignees": [],
            "task_type": "integration",
            "dependencies": [
                {
                    "repository": "bzzz", 
                    "task_number": 23,
                    "dependency_type": "api_contract"
                },
                {
                    "repository": "hive",
                    "task_number": 16, 
                    "dependency_type": "security"
                }
            ]
        },
        {
            "number": 9,
            "title": "Implement AI model coordination",
            "description": "Enable coordination between AI models across different development environments",
            "state": "open",
            "labels": ["bzzz-task", "ai-coordination", "models", "complex"],
            "created_at": "2025-01-14T11:15:00Z",
            "updated_at": "2025-01-14T11:30:00Z",
            "html_url": "https://github.com/mock/distributed-ai-dev/issues/9",
            "is_claimed": False,
            "assignees": [],
            "task_type": "feature",
            "dependencies": [
                {
                    "repository": "distributed-ai-dev",
                    "task_number": 8,
                    "dependency_type": "integration"
                }
            ]
        }
    ],
    4: [  # infra-automation tasks
        {
            "number": 12,
            "title": "Automate bzzz deployment across cluster",
            "description": "Create automated deployment scripts for bzzz agents on all cluster nodes",
            "state": "open",
            "labels": ["bzzz-task", "deployment", "automation", "devops"],
            "created_at": "2025-01-14T12:00:00Z",
            "updated_at": "2025-01-14T12:15:00Z",
            "html_url": "https://github.com/mock/infra-automation/issues/12",
            "is_claimed": False,
            "assignees": [],
            "task_type": "infrastructure",
            "dependencies": [
                {
                    "repository": "hive",
                    "task_number": 16,
                    "dependency_type": "security"
                }
            ]
        }
    ]
}

# Track claimed tasks
claimed_tasks = {}

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({"status": "healthy", "service": "mock-hive-api", "timestamp": datetime.now().isoformat()})

@app.route('/api/bzzz/active-repos', methods=['GET'])
def get_active_repositories():
    """Return mock active repositories for bzzz consumption"""
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 📡 Bzzz requested active repositories")
    
    # Randomly vary the number of available repos for more realistic testing
    available_repos = random.sample(MOCK_REPOSITORIES, k=random.randint(2, len(MOCK_REPOSITORIES)))
    
    return jsonify({"repositories": available_repos})

@app.route('/api/bzzz/projects/<int:project_id>/tasks', methods=['GET'])
def get_project_tasks(project_id):
    """Return mock bzzz-task labeled issues for a specific project"""
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 📋 Bzzz requested tasks for project {project_id}")
    
    if project_id not in MOCK_TASKS:
        return jsonify([])
    
    # Return tasks, updating claim status
    tasks = []
    for task in MOCK_TASKS[project_id]:
        task_copy = task.copy()
        claim_key = f"{project_id}-{task['number']}"
        
        # Check if task is claimed
        if claim_key in claimed_tasks:
            claim_info = claimed_tasks[claim_key]
            # Tasks expire after 30 minutes if not updated
            if datetime.now() - claim_info['claimed_at'] < timedelta(minutes=30):
                task_copy['is_claimed'] = True
                task_copy['assignees'] = [claim_info['agent_id']]
            else:
                # Claim expired
                del claimed_tasks[claim_key]
                task_copy['is_claimed'] = False
                task_copy['assignees'] = []
        
        tasks.append(task_copy)
    
    return jsonify(tasks)

@app.route('/api/bzzz/projects/<int:project_id>/claim', methods=['POST'])
def claim_task(project_id):
    """Register task claim with mock Hive system"""
    data = request.get_json()
    task_number = data.get('task_number')
    agent_id = data.get('agent_id')
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 🎯 Agent {agent_id} claiming task {project_id}#{task_number}")
    
    if not task_number or not agent_id:
        return jsonify({"error": "task_number and agent_id are required"}), 400
    
    claim_key = f"{project_id}-{task_number}"
    
    # Check if already claimed
    if claim_key in claimed_tasks:
        existing_claim = claimed_tasks[claim_key]
        if datetime.now() - existing_claim['claimed_at'] < timedelta(minutes=30):
            return jsonify({
                "error": "Task already claimed",
                "claimed_by": existing_claim['agent_id'],
                "claimed_at": existing_claim['claimed_at'].isoformat()
            }), 409
    
    # Register the claim
    claim_id = f"{project_id}-{task_number}-{agent_id}-{int(time.time())}"
    claimed_tasks[claim_key] = {
        "agent_id": agent_id,
        "claimed_at": datetime.now(),
        "claim_id": claim_id
    }
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] ✅ Task {project_id}#{task_number} claimed by {agent_id}")
    
    return jsonify({"success": True, "claim_id": claim_id})

@app.route('/api/bzzz/projects/<int:project_id>/status', methods=['PUT'])
def update_task_status(project_id):
    """Update task status in mock Hive system"""
    data = request.get_json()
    task_number = data.get('task_number')
    status = data.get('status')
    metadata = data.get('metadata', {})
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 📊 Task {project_id}#{task_number} status: {status}")
    
    if not task_number or not status:
        return jsonify({"error": "task_number and status are required"}), 400
    
    # Log status update
    if status == "completed":
        claim_key = f"{project_id}-{task_number}"
        if claim_key in claimed_tasks:
            agent_id = claimed_tasks[claim_key]['agent_id']
            print(f"[{datetime.now().strftime('%H:%M:%S')}] 🎉 Task {project_id}#{task_number} completed by {agent_id}")
            del claimed_tasks[claim_key]  # Remove claim
    elif status == "escalated":
        print(f"[{datetime.now().strftime('%H:%M:%S')}] 🚨 Task {project_id}#{task_number} escalated: {metadata}")
    
    return jsonify({"success": True})

@app.route('/api/bzzz/coordination-log', methods=['POST'])
def log_coordination_activity():
    """Log coordination activity for monitoring"""
    data = request.get_json()
    activity_type = data.get('type', 'unknown')
    details = data.get('details', {})
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 🧠 Coordination: {activity_type} - {details}")
    
    # Save coordination activity to file
    save_coordination_work(activity_type, details)
    
    return jsonify({"success": True, "logged": True})

@app.route('/api/bzzz/projects/<int:project_id>/submit-work', methods=['POST'])
def submit_work(project_id):
    """Endpoint for agents to submit their actual work/code/solutions"""
    data = request.get_json()
    task_number = data.get('task_number')
    agent_id = data.get('agent_id')
    work_type = data.get('work_type', 'code')  # code, documentation, configuration, etc.
    content = data.get('content', '')
    files = data.get('files', {})  # Dictionary of filename -> content
    commit_message = data.get('commit_message', '')
    description = data.get('description', '')
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 📝 Work submission: {agent_id} -> Project {project_id} Task {task_number}")
    print(f"   Type: {work_type}, Files: {len(files)}, Content length: {len(content)}")
    
    # Save the actual work content
    work_data = {
        "project_id": project_id,
        "task_number": task_number,
        "agent_id": agent_id,
        "work_type": work_type,
        "content": content,
        "files": files,
        "commit_message": commit_message,
        "description": description,
        "submitted_at": datetime.now().isoformat()
    }
    
    save_agent_work(work_data)
    
    return jsonify({
        "success": True, 
        "work_id": f"{project_id}-{task_number}-{int(time.time())}",
        "message": "Work submitted successfully to mock repository"
    })

@app.route('/api/bzzz/projects/<int:project_id>/create-pr', methods=['POST'])
def create_pull_request(project_id):
    """Endpoint for agents to submit pull request content"""
    data = request.get_json()
    task_number = data.get('task_number')
    agent_id = data.get('agent_id')
    pr_title = data.get('title', '')
    pr_description = data.get('description', '')
    files_changed = data.get('files_changed', {})
    branch_name = data.get('branch_name', f"bzzz-task-{task_number}")
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 🔀 Pull Request: {agent_id} -> Project {project_id}")
    print(f"   Title: {pr_title}")
    print(f"   Files changed: {len(files_changed)}")
    
    # Save the pull request content
    pr_data = {
        "project_id": project_id,
        "task_number": task_number,
        "agent_id": agent_id,
        "title": pr_title,
        "description": pr_description,
        "files_changed": files_changed,
        "branch_name": branch_name,
        "created_at": datetime.now().isoformat(),
        "status": "open"
    }
    
    save_pull_request(pr_data)
    
    return jsonify({
        "success": True,
        "pr_number": random.randint(100, 999),
        "pr_url": f"https://github.com/mock/{get_repo_name(project_id)}/pull/{random.randint(100, 999)}",
        "message": "Pull request created successfully in mock repository"
    })

@app.route('/api/bzzz/projects/<int:project_id>/coordination-discussion', methods=['POST'])
def log_coordination_discussion(project_id):
    """Endpoint for agents to log coordination discussions and decisions"""
    data = request.get_json()
    discussion_type = data.get('type', 'general')  # dependency_analysis, conflict_resolution, etc.
    participants = data.get('participants', [])
    messages = data.get('messages', [])
    decisions = data.get('decisions', [])
    context = data.get('context', {})
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 💬 Coordination Discussion: Project {project_id}")
    print(f"   Type: {discussion_type}, Participants: {len(participants)}, Messages: {len(messages)}")
    
    # Save coordination discussion
    discussion_data = {
        "project_id": project_id,
        "type": discussion_type,
        "participants": participants,
        "messages": messages,
        "decisions": decisions,
        "context": context,
        "timestamp": datetime.now().isoformat()
    }
    
    save_coordination_discussion(discussion_data)
    
    return jsonify({"success": True, "logged": True})

@app.route('/api/bzzz/projects/<int:project_id>/log-prompt', methods=['POST'])
def log_agent_prompt(project_id):
    """Endpoint for agents to log the prompts they are receiving/generating"""
    data = request.get_json()
    task_number = data.get('task_number')
    agent_id = data.get('agent_id')
    prompt_type = data.get('prompt_type', 'task_analysis')  # task_analysis, coordination, meta_thinking
    prompt_content = data.get('prompt_content', '')
    context = data.get('context', {})
    model_used = data.get('model_used', 'unknown')
    
    print(f"[{datetime.now().strftime('%H:%M:%S')}] 🧠 Prompt Log: {agent_id} -> {prompt_type}")
    print(f"   Model: {model_used}, Task: {project_id}#{task_number}")
    print(f"   Prompt length: {len(prompt_content)} chars")
    
    # Save the prompt data
    prompt_data = {
        "project_id": project_id,
        "task_number": task_number,
        "agent_id": agent_id,
        "prompt_type": prompt_type,
        "prompt_content": prompt_content,
        "context": context,
        "model_used": model_used,
        "timestamp": datetime.now().isoformat()
    }
    
    save_agent_prompt(prompt_data)
    
    return jsonify({"success": True, "logged": True})

def save_agent_prompt(prompt_data):
    """Save agent prompts to files for analysis"""
    import os
    timestamp = datetime.now()
    work_dir = "/tmp/bzzz_agent_prompts"
    os.makedirs(work_dir, exist_ok=True)
    
    # Create filename with project, task, and timestamp
    project_id = prompt_data["project_id"]
    task_number = prompt_data["task_number"]
    agent_id = prompt_data["agent_id"].replace("/", "_")  # Clean agent ID for filename
    prompt_type = prompt_data["prompt_type"]
    
    filename = f"prompt_{prompt_type}_p{project_id}_t{task_number}_{agent_id}_{timestamp.strftime('%H%M%S')}.json"
    prompt_file = os.path.join(work_dir, filename)
    
    with open(prompt_file, "w") as f:
        json.dump(prompt_data, f, indent=2)
    
    print(f"   💾 Saved prompt to: {prompt_file}")
    
    # Also save to daily log
    log_file = os.path.join(work_dir, f"agent_prompts_log_{timestamp.strftime('%Y%m%d')}.jsonl")
    with open(log_file, "a") as f:
        f.write(json.dumps(prompt_data) + "\n")

def save_agent_work(work_data):
    """Save actual agent work submissions to files"""
    import os
    timestamp = datetime.now()
    work_dir = "/tmp/bzzz_agent_work"
    os.makedirs(work_dir, exist_ok=True)
    
    # Create filename with project, task, and timestamp
    project_id = work_data["project_id"]
    task_number = work_data["task_number"]
    agent_id = work_data["agent_id"].replace("/", "_")  # Clean agent ID for filename
    
    filename = f"work_p{project_id}_t{task_number}_{agent_id}_{timestamp.strftime('%H%M%S')}.json"
    work_file = os.path.join(work_dir, filename)
    
    with open(work_file, "w") as f:
        json.dump(work_data, f, indent=2)
    
    print(f"   💾 Saved work to: {work_file}")
    
    # Also save to daily log
    log_file = os.path.join(work_dir, f"agent_work_log_{timestamp.strftime('%Y%m%d')}.jsonl")
    with open(log_file, "a") as f:
        f.write(json.dumps(work_data) + "\n")

def save_pull_request(pr_data):
    """Save pull request content to files"""
    import os
    timestamp = datetime.now()
    work_dir = "/tmp/bzzz_pull_requests"
    os.makedirs(work_dir, exist_ok=True)
    
    # Create filename with project, task, and timestamp
    project_id = pr_data["project_id"]
    task_number = pr_data["task_number"]
    agent_id = pr_data["agent_id"].replace("/", "_")  # Clean agent ID for filename
    
    filename = f"pr_p{project_id}_t{task_number}_{agent_id}_{timestamp.strftime('%H%M%S')}.json"
    pr_file = os.path.join(work_dir, filename)
    
    with open(pr_file, "w") as f:
        json.dump(pr_data, f, indent=2)
    
    print(f"   💾 Saved PR to: {pr_file}")
    
    # Also save to daily log
    log_file = os.path.join(work_dir, f"pull_requests_log_{timestamp.strftime('%Y%m%d')}.jsonl")
    with open(log_file, "a") as f:
        f.write(json.dumps(pr_data) + "\n")

def save_coordination_discussion(discussion_data):
    """Save coordination discussions to files"""
    import os
    timestamp = datetime.now()
    work_dir = "/tmp/bzzz_coordination_discussions"
    os.makedirs(work_dir, exist_ok=True)
    
    # Create filename with project and timestamp
    project_id = discussion_data["project_id"]
    discussion_type = discussion_data["type"]
    
    filename = f"discussion_{discussion_type}_p{project_id}_{timestamp.strftime('%H%M%S')}.json"
    discussion_file = os.path.join(work_dir, filename)
    
    with open(discussion_file, "w") as f:
        json.dump(discussion_data, f, indent=2)
    
    print(f"   💾 Saved discussion to: {discussion_file}")
    
    # Also save to daily log
    log_file = os.path.join(work_dir, f"coordination_discussions_{timestamp.strftime('%Y%m%d')}.jsonl")
    with open(log_file, "a") as f:
        f.write(json.dumps(discussion_data) + "\n")

def get_repo_name(project_id):
    """Get repository name from project ID"""
    repo_map = {
        1: "hive",
        2: "bzzz", 
        3: "distributed-ai-dev",
        4: "infra-automation"
    }
    return repo_map.get(project_id, "unknown-repo")

def save_coordination_work(activity_type, details):
    """Save coordination work to files for analysis"""
    timestamp = datetime.now()
    work_dir = "/tmp/bzzz_coordination_work"
    os.makedirs(work_dir, exist_ok=True)
    
    # Create detailed log entry
    work_entry = {
        "timestamp": timestamp.isoformat(),
        "type": activity_type,
        "details": details,
        "session_id": details.get("session_id", "unknown")
    }
    
    # Save to daily log file
    log_file = os.path.join(work_dir, f"coordination_work_{timestamp.strftime('%Y%m%d')}.jsonl")
    with open(log_file, "a") as f:
        f.write(json.dumps(work_entry) + "\n")
    
    # Save individual work items to separate files
    if activity_type in ["code_generation", "task_solution", "pull_request_content"]:
        work_file = os.path.join(work_dir, f"{activity_type}_{timestamp.strftime('%H%M%S')}.json")
        with open(work_file, "w") as f:
            json.dump(work_entry, f, indent=2)

def start_background_task_updates():
    """Background thread to simulate changing task priorities and new tasks"""
    def background_updates():
        while True:
            time.sleep(random.randint(60, 180))  # Every 1-3 minutes
            
            # Occasionally add a new urgent task
            if random.random() < 0.3:  # 30% chance
                project_id = random.choice([1, 2, 3, 4])
                urgent_task = {
                    "number": random.randint(100, 999),
                    "title": f"URGENT: {random.choice(['Critical bug fix', 'Security patch', 'Production issue', 'Integration failure'])}",
                    "description": "High priority task requiring immediate attention",
                    "state": "open",
                    "labels": ["bzzz-task", "urgent", "critical"],
                    "created_at": datetime.now().isoformat(),
                    "updated_at": datetime.now().isoformat(),
                    "html_url": f"https://github.com/mock/repo/issues/{random.randint(100, 999)}",
                    "is_claimed": False,
                    "assignees": [],
                    "task_type": "bug",
                    "dependencies": []
                }
                
                if project_id not in MOCK_TASKS:
                    MOCK_TASKS[project_id] = []
                MOCK_TASKS[project_id].append(urgent_task)
                
                print(f"[{datetime.now().strftime('%H:%M:%S')}] 🚨 NEW URGENT TASK: Project {project_id} - {urgent_task['title']}")
    
    thread = Thread(target=background_updates, daemon=True)
    thread.start()

if __name__ == '__main__':
    print("🚀 Starting Mock Hive API Server for Bzzz Testing")
    print("=" * 50)
    print("This server provides fake projects and tasks to real bzzz agents")
    print("Real bzzz coordination will happen with this simulated data")
    print("")
    print("Available endpoints:")
    print("  GET  /health - Health check")
    print("  GET  /api/bzzz/active-repos - Active repositories")
    print("  GET  /api/bzzz/projects/<id>/tasks - Project tasks")
    print("  POST /api/bzzz/projects/<id>/claim - Claim task")
    print("  PUT  /api/bzzz/projects/<id>/status - Update task status")
    print("  POST /api/bzzz/projects/<id>/submit-work - Submit actual work/code")
    print("  POST /api/bzzz/projects/<id>/create-pr - Submit pull request content")
    print("  POST /api/bzzz/projects/<id>/coordination-discussion - Log coordination discussions")
    print("  POST /api/bzzz/projects/<id>/log-prompt - Log agent prompts and model usage")
    print("  POST /api/bzzz/coordination-log - Log coordination activity")
    print("")
    print("Starting background task updates...")
    start_background_task_updates()
    
    print(f"🌟 Mock Hive API running on http://localhost:5000")
    print("Configure bzzz to use: BZZZ_HIVE_API_URL=http://localhost:5000")
    print("")
    
    app.run(host='0.0.0.0', port=5000, debug=False)