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
    
    return jsonify({"success": True, "logged": True})

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
    print("")
    print("Starting background task updates...")
    start_background_task_updates()
    
    print(f"🌟 Mock Hive API running on http://localhost:5000")
    print("Configure bzzz to use: BZZZ_HIVE_API_URL=http://localhost:5000")
    print("")
    
    app.run(host='0.0.0.0', port=5000, debug=False)