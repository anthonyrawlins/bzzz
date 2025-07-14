#!/usr/bin/env python3
"""
Simple test to check GitHub API access for bzzz-task issues.
"""

import requests
from pathlib import Path

def get_github_token():
    """Get GitHub token from secrets file."""
    try:
        # Try gh-token first
        gh_token_path = Path("/home/tony/AI/secrets/passwords_and_tokens/gh-token")
        if gh_token_path.exists():
            return gh_token_path.read_text().strip()
        
        # Try GitHub token
        github_token_path = Path("/home/tony/AI/secrets/passwords_and_tokens/github-token")
        if github_token_path.exists():
            return github_token_path.read_text().strip()
        
        # Fallback to GitLab token if GitHub token doesn't exist
        gitlab_token_path = Path("/home/tony/AI/secrets/passwords_and_tokens/claude-gitlab-token")
        if gitlab_token_path.exists():
            return gitlab_token_path.read_text().strip()
    except Exception:
        pass
    return None

def test_github_bzzz_tasks():
    """Test fetching bzzz-task issues from GitHub."""
    token = get_github_token()
    if not token:
        print("❌ No GitHub token found")
        return
    
    print("🐙 Testing GitHub API access for bzzz-task issues...")
    
    # Test with the hive repository
    repo = "anthonyrawlins/hive"
    url = f"https://api.github.com/repos/{repo}/issues"
    
    headers = {
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github.v3+json"
    }
    
    # First, get all open issues
    print(f"\n📊 Fetching all open issues from {repo}...")
    response = requests.get(url, headers=headers, params={"state": "open"}, timeout=10)
    
    if response.status_code == 200:
        all_issues = response.json()
        print(f"Found {len(all_issues)} total open issues")
        
        # Show all labels used in the repository
        all_labels = set()
        for issue in all_issues:
            for label in issue.get('labels', []):
                all_labels.add(label['name'])
        
        print(f"All labels in use: {sorted(all_labels)}")
        
    else:
        print(f"❌ Failed to fetch issues: {response.status_code} - {response.text}")
        return
    
    # Now test for bzzz-task labeled issues
    print(f"\n🐝 Fetching bzzz-task labeled issues from {repo}...")
    response = requests.get(url, headers=headers, params={"labels": "bzzz-task", "state": "open"}, timeout=10)
    
    if response.status_code == 200:
        bzzz_issues = response.json()
        print(f"Found {len(bzzz_issues)} issues with 'bzzz-task' label")
        
        if not bzzz_issues:
            print("ℹ️  No issues found with 'bzzz-task' label")
            print("   You can create test issues with this label for testing")
        
        for issue in bzzz_issues:
            print(f"\n  🎫 Issue #{issue['number']}: {issue['title']}")
            print(f"     State: {issue['state']}")
            print(f"     Labels: {[label['name'] for label in issue.get('labels', [])]}")
            print(f"     Assignees: {[assignee['login'] for assignee in issue.get('assignees', [])]}")
            print(f"     URL: {issue['html_url']}")
    else:
        print(f"❌ Failed to fetch bzzz-task issues: {response.status_code} - {response.text}")

def main():
    print("🚀 Simple GitHub API Test for Bzzz Integration")
    print("="*50)
    test_github_bzzz_tasks()

if __name__ == "__main__":
    main()