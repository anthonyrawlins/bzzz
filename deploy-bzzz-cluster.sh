#!/bin/bash

# Bzzz P2P Service Cluster Deployment Script
# Deploys updated Bzzz binary from walnut to other cluster nodes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BZZZ_DIR="/home/tony/AI/projects/Bzzz"
# Exclude walnut (192.168.1.27) since this IS walnut
CLUSTER_NODES=("192.168.1.72" "192.168.1.113" "192.168.1.132")
CLUSTER_NAMES=("ACACIA" "IRONWOOD" "ROSEWOOD")
SSH_USER="tony"
SSH_PASS="silverfrond[1392]"

# Logging functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if bzzz binary exists
check_binary() {
    log "Checking for Bzzz binary on walnut..."
    
    if [ ! -f "$BZZZ_DIR/bzzz" ]; then
        error "Bzzz binary not found at $BZZZ_DIR/bzzz"
        echo "   Please build the binary first with: go build -o bzzz main.go"
        exit 1
    fi
    
    success "Bzzz binary found and ready for deployment"
}

# Update walnut's own service
update_walnut() {
    log "Updating Bzzz service on walnut (local)..."
    
    # Check if binary has been built recently
    if [ ! -f "$BZZZ_DIR/bzzz" ]; then
        error "Bzzz binary not found. Building..."
        cd "$BZZZ_DIR"
        go build -o bzzz main.go || { error "Build failed"; return 1; }
    fi
    
    # Stop the service
    sudo systemctl stop bzzz.service 2>/dev/null || true
    
    # Backup old binary
    sudo cp /usr/local/bin/bzzz /usr/local/bin/bzzz.backup 2>/dev/null || true
    
    # Install new binary
    sudo cp "$BZZZ_DIR/bzzz" /usr/local/bin/bzzz
    sudo chmod +x /usr/local/bin/bzzz
    sudo chown root:root /usr/local/bin/bzzz
    
    # Start the service
    sudo systemctl start bzzz.service
    
    # Check if service started successfully
    sleep 3
    if sudo systemctl is-active bzzz.service > /dev/null 2>&1; then
        success "✓ WALNUT (local) - Binary updated and service restarted"
    else
        error "✗ WALNUT (local) - Service failed to start"
        return 1
    fi
}

# Check cluster connectivity
check_cluster_connectivity() {
    log "Checking cluster connectivity from walnut..."
    
    for i in "${!CLUSTER_NODES[@]}"; do
        node="${CLUSTER_NODES[$i]}"
        name="${CLUSTER_NAMES[$i]}"
        
        log "Testing connection to $name ($node)..."
        
        if sshpass -p "$SSH_PASS" ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no "$SSH_USER@$node" "echo 'Connection test successful'" > /dev/null 2>&1; then
            success "✓ $name ($node) - Connected"
        else
            warning "✗ $name ($node) - Connection failed"
        fi
    done
}

# Deploy bzzz binary to remote cluster nodes
deploy_bzzz_binary() {
    log "Deploying Bzzz binary from walnut to remote cluster nodes..."
    
    # Make sure binary is executable
    chmod +x "$BZZZ_DIR/bzzz"
    
    for i in "${!CLUSTER_NODES[@]}"; do
        node="${CLUSTER_NODES[$i]}"
        name="${CLUSTER_NAMES[$i]}"
        
        log "Deploying to $name ($node)..."
        
        # Copy the binary
        if sshpass -p "$SSH_PASS" scp -o StrictHostKeyChecking=no "$BZZZ_DIR/bzzz" "$SSH_USER@$node:/tmp/bzzz-new"; then
            
            # Install the binary and restart service
            sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_USER@$node" "
                # Stop the service
                sudo systemctl stop bzzz.service 2>/dev/null || true
                
                # Backup old binary
                sudo cp /usr/local/bin/bzzz /usr/local/bin/bzzz.backup 2>/dev/null || true
                
                # Install new binary
                sudo mv /tmp/bzzz-new /usr/local/bin/bzzz
                sudo chmod +x /usr/local/bin/bzzz
                sudo chown root:root /usr/local/bin/bzzz
                
                # Start the service
                sudo systemctl start bzzz.service
                
                # Check if service started successfully
                sleep 3
                if sudo systemctl is-active bzzz.service > /dev/null 2>&1; then
                    echo 'Service started successfully'
                else
                    echo 'Service failed to start'
                    exit 1
                fi
            "
            
            if [ $? -eq 0 ]; then
                success "✓ $name - Binary deployed and service restarted"
            else
                error "✗ $name - Deployment failed"
            fi
        else
            error "✗ $name - Failed to copy binary"
        fi
    done
}

# Verify cluster status after deployment
verify_cluster_status() {
    log "Verifying cluster status after deployment..."
    
    sleep 10  # Wait for services to fully start
    
    # Check walnut (local)
    log "Checking WALNUT (local) status..."
    if sudo systemctl is-active bzzz.service > /dev/null 2>&1; then
        success "✓ WALNUT (local) - Service is running"
    else
        error "✗ WALNUT (local) - Service is not running"
    fi
    
    # Check remote nodes
    for i in "${!CLUSTER_NODES[@]}"; do
        node="${CLUSTER_NODES[$i]}"
        name="${CLUSTER_NAMES[$i]}"
        
        log "Checking $name ($node) status..."
        
        status=$(sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_USER@$node" "
            if sudo systemctl is-active bzzz.service > /dev/null 2>&1; then
                echo 'RUNNING'
            else
                echo 'FAILED'
            fi
        " 2>/dev/null || echo "CONNECTION_FAILED")
        
        case $status in
            "RUNNING")
                success "✓ $name - Service is running"
                ;;
            "FAILED")
                error "✗ $name - Service is not running"
                ;;
            "CONNECTION_FAILED")
                error "✗ $name - Cannot connect to check status"
                ;;
        esac
    done
}

# Test Hive connectivity from all nodes
test_hive_connectivity() {
    log "Testing Hive API connectivity from all cluster nodes..."
    
    # Test from walnut (local)
    log "Testing Hive connectivity from WALNUT (local)..."
    if curl -s -o /dev/null -w '%{http_code}' --connect-timeout 10 https://hive.home.deepblack.cloud/health 2>/dev/null | grep -q "200"; then
        success "✓ WALNUT (local) - Can reach Hive API"
    else
        warning "✗ WALNUT (local) - Cannot reach Hive API"
    fi
    
    # Test from remote nodes
    for i in "${!CLUSTER_NODES[@]}"; do
        node="${CLUSTER_NODES[$i]}"
        name="${CLUSTER_NAMES[$i]}"
        
        log "Testing Hive connectivity from $name ($node)..."
        
        result=$(sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_USER@$node" "
            curl -s -o /dev/null -w '%{http_code}' --connect-timeout 10 https://hive.home.deepblack.cloud/health 2>/dev/null || echo 'FAILED'
        " 2>/dev/null || echo "CONNECTION_FAILED")
        
        case $result in
            "200")
                success "✓ $name - Can reach Hive API"
                ;;
            "FAILED"|"CONNECTION_FAILED"|*)
                warning "✗ $name - Cannot reach Hive API (response: $result)"
                ;;
        esac
    done
}

# Main deployment function
main() {
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                 Bzzz Cluster Deployment                     ║"
    echo "║                                                              ║"
    echo "║  Deploying updated Bzzz binary from WALNUT to cluster       ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    log "Starting deployment from walnut to P2P mesh cluster..."
    
    # Run deployment steps
    check_binary
    update_walnut
    check_cluster_connectivity
    deploy_bzzz_binary
    verify_cluster_status
    test_hive_connectivity
    
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                 Deployment Completed!                       ║"
    echo "║                                                              ║"
    echo "║  🐝 Bzzz P2P mesh is now running with updated binary        ║"
    echo "║  🔗 Hive integration: https://hive.home.deepblack.cloud     ║"
    echo "║  📡 Check logs for P2P mesh formation and task discovery    ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "status")
        log "Checking cluster status..."
        echo -e "\n${BLUE}=== WALNUT (local) ===${NC}"
        sudo systemctl status bzzz.service --no-pager -l
        
        for i in "${!CLUSTER_NODES[@]}"; do
            node="${CLUSTER_NODES[$i]}"
            name="${CLUSTER_NAMES[$i]}"
            echo -e "\n${BLUE}=== $name ($node) ===${NC}"
            sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_USER@$node" "sudo systemctl status bzzz.service --no-pager -l" 2>/dev/null || echo "Connection failed"
        done
        ;;
    "logs")
        if [ -z "$2" ]; then
            echo "Usage: $0 logs <node_name>"
            echo "Available nodes: WALNUT ${CLUSTER_NAMES[*]}"
            exit 1
        fi
        
        if [ "$2" = "WALNUT" ]; then
            log "Showing logs from WALNUT (local)..."
            sudo journalctl -u bzzz -f
            exit 0
        fi
        
        # Find remote node by name
        for i in "${!CLUSTER_NAMES[@]}"; do
            if [ "${CLUSTER_NAMES[$i]}" = "$2" ]; then
                node="${CLUSTER_NODES[$i]}"
                log "Showing logs from $2 ($node)..."
                sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_USER@$node" "sudo journalctl -u bzzz -f"
                exit 0
            fi
        done
        error "Node '$2' not found. Available: WALNUT ${CLUSTER_NAMES[*]}"
        ;;
    "test")
        log "Testing Hive connectivity..."
        test_hive_connectivity
        ;;
    *)
        echo "Usage: $0 {deploy|status|logs <node_name>|test}"
        echo ""
        echo "Commands:"
        echo "  deploy        - Deploy updated Bzzz binary from walnut to cluster"
        echo "  status        - Show service status on all nodes"
        echo "  logs <node>   - Show logs from specific node (WALNUT ${CLUSTER_NAMES[*]})"
        echo "  test          - Test Hive API connectivity from all nodes"
        exit 1
        ;;
esac