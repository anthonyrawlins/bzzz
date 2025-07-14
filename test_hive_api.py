#!/usr/bin/env python3
"""
Test script for Bzzz-Hive API integration.
Tests the newly created API endpoints for dynamic repository discovery.
"""

import sys
import os
sys.path.append('/home/tony/AI/projects/hive/backend')

from app.services.project_service import ProjectService
import json

def test_project_service():
    """Test the ProjectService with Bzzz integration methods."""
    print("🧪 Testing ProjectService with Bzzz integration...")
    
    service = ProjectService()
    
    # Test 1: Get all projects
    print("\n📁 Testing get_all_projects()...")
    projects = service.get_all_projects()
    print(f"Found {len(projects)} total projects")
    
    # Find projects with GitHub repos
    github_projects = [p for p in projects if p.get('github_repo')]
    print(f"Found {len(github_projects)} projects with GitHub repositories:")
    for project in github_projects:
        print(f"  - {project['name']}: {project['github_repo']}")
    
    # Test 2: Get active repositories for Bzzz
    print("\n🐝 Testing get_bzzz_active_repositories()...")
    try:
        active_repos = service.get_bzzz_active_repositories()
        print(f"Found {len(active_repos)} repositories ready for Bzzz coordination:")
        
        for repo in active_repos:
            print(f"\n  📦 Repository: {repo['name']}")
            print(f"     Owner: {repo['owner']}")
            print(f"     Repository: {repo['repository']}")
            print(f"     Git URL: {repo['git_url']}")
            print(f"     Ready to claim: {repo['ready_to_claim']}")
            print(f"     Project ID: {repo['project_id']}")
            
    except Exception as e:
        print(f"❌ Error testing active repositories: {e}")
    
    # Test 3: Get bzzz-task issues for the hive project specifically
    print("\n🎯 Testing get_bzzz_project_tasks() for 'hive' project...")
    try:
        hive_tasks = service.get_bzzz_project_tasks('hive')
        print(f"Found {len(hive_tasks)} bzzz-task issues in hive project:")
        
        for task in hive_tasks:
            print(f"\n  🎫 Issue #{task['number']}: {task['title']}")
            print(f"     State: {task['state']}")
            print(f"     Labels: {task['labels']}")
            print(f"     Task Type: {task['task_type']}")
            print(f"     Claimed: {task['is_claimed']}")
            if task['assignees']:
                print(f"     Assignees: {', '.join(task['assignees'])}")
            print(f"     URL: {task['html_url']}")
            
    except Exception as e:
        print(f"❌ Error testing hive project tasks: {e}")
    
    # Test 4: Simulate API endpoint response format
    print("\n📡 Testing API endpoint response format...")
    try:
        active_repos = service.get_bzzz_active_repositories()
        api_response = {"repositories": active_repos}
        
        print("API Response Preview (first 500 chars):")
        response_json = json.dumps(api_response, indent=2)
        print(response_json[:500] + "..." if len(response_json) > 500 else response_json)
        
    except Exception as e:
        print(f"❌ Error formatting API response: {e}")

def main():
    print("🚀 Starting Bzzz-Hive API Integration Test")
    print("="*50)
    
    try:
        test_project_service()
        print("\n✅ Test completed successfully!")
        
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()