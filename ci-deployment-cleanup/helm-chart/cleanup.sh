#!/bin/bash
set -e

# Debug logging function
debug_log() {
    if [ "$DEBUG_MODE" = "true" ]; then
        echo "[DEBUG] $*" >&2
    fi
}

echo "Starting Renku CI deployment cleanup..."
debug_log "Debug mode is enabled"
debug_log "Environment variables: DRY_RUN=$DRY_RUN, MAX_AGE_HOURS=$MAX_AGE_HOURS"

echo "Max age: $MAX_AGE_HOURS hours"
echo "Exemption label: $EXEMPTION_LABEL"

if [ "$ENFORCE_NAME_PATTERNS" = "true" ]; then
    echo "Name pattern enforcement: enabled"
    echo "Allowed patterns:"
    # NAMESPACE_PATTERNS should be a space-separated list
    for pattern in $NAMESPACE_PATTERNS; do
        echo "  - $pattern"
    done
else
    echo "Name pattern enforcement: disabled"
fi

if [ "$PR_CLEANUP_ENABLED" = "true" ]; then
    echo "PR-based cleanup: enabled"
    echo "Repository mappings:"
    # PR_REPOSITORIES should be formatted as "pattern1:repo1 pattern2:repo2"
    for mapping in $PR_REPOSITORIES; do
        pattern=$(echo "$mapping" | cut -d':' -f1)
        repo=$(echo "$mapping" | cut -d':' -f2)
        echo "  - $pattern -> $repo"
    done
else
    echo "PR-based cleanup: disabled"
fi

if [ "$DRY_RUN" = "true" ]; then
    echo "DRY RUN MODE: No actual deletions will be performed"
fi

debug_log "Initialization complete, starting namespace discovery"

# Function to calculate age in seconds
calculate_age() {
    local timestamp="$1"
    local current_time=$(date +%s)
    
    debug_log "Calculating age for timestamp: $timestamp"
    
    # Kubernetes timestamps are in ISO 8601 format, need to handle them properly
    local creation_time
    if command -v gdate >/dev/null 2>&1; then
        # Use GNU date if available (Linux with coreutils)
        creation_time=$(gdate -d "$timestamp" +%s 2>/dev/null || echo "0")
        debug_log "Used gdate for timestamp parsing"
    else
        # For Alpine Linux/BusyBox date, we need to parse the ISO 8601 format manually
        # Format: 2025-05-28T13:50:39Z
        local year month day hour minute second
        year=$(echo "$timestamp" | cut -d'-' -f1)
        month=$(echo "$timestamp" | cut -d'-' -f2)
        day=$(echo "$timestamp" | cut -d'T' -f1 | cut -d'-' -f3)
        hour=$(echo "$timestamp" | cut -d'T' -f2 | cut -d':' -f1)
        minute=$(echo "$timestamp" | cut -d':' -f2)
        second=$(echo "$timestamp" | cut -d':' -f3 | sed 's/Z$//')
        
        debug_log "Parsed timestamp components: $year-$month-$day $hour:$minute:$second"
        
        # Use BusyBox date with explicit format
        local formatted_timestamp="${year}-${month}-${day} ${hour}:${minute}:${second}"
        creation_time=$(date -d "$formatted_timestamp" +%s 2>/dev/null || echo "0")
        debug_log "Used BusyBox date for timestamp parsing"
    fi
    
    if [ "$creation_time" = "0" ]; then
        debug_log "Failed to parse timestamp, returning age 0"
        echo "0"
    else
        local age=$((current_time - creation_time))
        debug_log "Calculated age: $age seconds"
        echo "$age"
    fi
}

# Function to format age for display
format_age() {
    local age_seconds="$1"
    local age_hours=$((age_seconds / 3600))
    local age_days=$((age_hours / 24))
    
    if [ $age_days -gt 0 ]; then
        echo "${age_days}d $((age_hours % 24))h"
    else
        echo "${age_hours}h"
    fi
}

# Function to format hours to days+hours for thresholds
format_hours_threshold() {
    local hours="$1"
    local days=$((hours / 24))
    
    if [ $days -gt 0 ]; then
        echo "${days}d ($((hours % 24))h)"
    else
        echo "${hours}h"
    fi
}

# Function to check if namespace matches any allowed pattern
matches_pattern() {
    local namespace="$1"
    debug_log "Checking if namespace '$namespace' matches any allowed patterns"
    
    if [ "$ENFORCE_NAME_PATTERNS" = "true" ]; then
        for pattern in $NAMESPACE_PATTERNS; do
            debug_log "Testing pattern: $pattern"
            if [[ "$namespace" =~ $pattern ]]; then
                debug_log "Namespace matches pattern: $pattern"
                return 0
            fi
        done
        debug_log "Namespace does not match any patterns"
        return 1
    else
        # Pattern enforcement disabled, always return true
        debug_log "Pattern enforcement disabled, allowing all namespaces"
        return 0
    fi
}

# Function to check GitHub PR status
check_pr_status() {
    local namespace="$1"
    local github_token="${GITHUB_TOKEN}"
    
    debug_log "Checking PR status for namespace: $namespace"
    
    if [ -z "$github_token" ]; then
        echo "  → GitHub token not configured, skipping PR status check"
        debug_log "No GitHub token available"
        return 1
    fi
    
    # Check each repository mapping
    for mapping in $PR_REPOSITORIES; do
        local pattern=$(echo "$mapping" | cut -d':' -f1)
        local repo=$(echo "$mapping" | cut -d':' -f2)
        
        debug_log "Checking mapping: $pattern -> $repo"
        
        if [[ "$namespace" =~ $pattern ]]; then
            debug_log "Namespace matches PR pattern: $pattern"
            
            # Extract PR number (assuming first capture group)
            local pr_number="${BASH_REMATCH[1]}"
            
            if [ -z "$pr_number" ]; then
                echo "  → Could not extract PR number from namespace $namespace"
                debug_log "Failed to extract PR number"
                return 1
            fi
            
            echo "  → Checking PR #$pr_number status in $repo"
            debug_log "Querying GitHub API for PR #$pr_number in $repo"
            
            # Query GitHub API for PR status
            local pr_response
            pr_response=$(curl -s -H "Authorization: token $github_token" \
                "https://api.github.com/repos/$repo/pulls/$pr_number" 2>/dev/null)
            
            if [ $? -ne 0 ]; then
                echo "  → Failed to query GitHub API for PR #$pr_number"
                debug_log "GitHub API request failed"
                return 1
            fi
            
            debug_log "GitHub API response received"
            
            # Check if PR exists and get its state
            local pr_state
            pr_state=$(echo "$pr_response" | grep -o '"state":[[:space:]]*"[^"]*"' | sed 's/"state":[[:space:]]*"\([^"]*\)"/\1/')
            
            if [ -z "$pr_state" ]; then
                echo "  → PR #$pr_number not found in $repo"
                debug_log "PR not found in repository"
                # Set global variable for dry run messaging
                PR_CLEANUP_REASON="PR #$pr_number not found in $repo"
                return 0  # PR doesn't exist, can clean up
            fi
            
            echo "  → PR #$pr_number state: $pr_state"
            debug_log "PR state: $pr_state"
            
            # Check if PR is closed or merged
            if [ "$pr_state" = "closed" ]; then
                # For closed PRs, check if it was merged
                local merged
                merged=$(echo "$pr_response" | grep -o '"merged":[[:space:]]*[^,}]*' | sed 's/"merged":[[:space:]]*\([^,}]*\)/\1/')
                debug_log "PR merged status: $merged"
                
                if [ "$merged" = "true" ]; then
                    echo "  → PR #$pr_number is merged, eligible for cleanup"
                    PR_CLEANUP_REASON="PR #$pr_number is merged in $repo"
                else
                    echo "  → PR #$pr_number is closed but not merged, eligible for cleanup"
                    PR_CLEANUP_REASON="PR #$pr_number is closed (not merged) in $repo"
                fi
                return 0  # Can clean up
            elif [ "$pr_state" = "open" ]; then
                echo "  → PR #$pr_number is still open, skipping cleanup"
                debug_log "PR is still open, cannot clean up"
                return 1  # Cannot clean up
            else
                echo "  → PR #$pr_number has unknown state: $pr_state"
                debug_log "Unknown PR state: $pr_state"
                return 1  # Unknown state, skip cleanup
            fi
        fi
    done
    
    echo "  → Namespace $namespace does not match any PR cleanup patterns"
    debug_log "No matching PR cleanup patterns"
    return 1  # No matching pattern
}

# Get maximum age in seconds
MAX_AGE_SECONDS=$(( MAX_AGE_HOURS * 3600 ))
debug_log "Maximum age threshold: $MAX_AGE_SECONDS seconds ($MAX_AGE_HOURS hours)"

# Find and process all namespaces
debug_log "Starting namespace enumeration"
kubectl get namespaces \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.creationTimestamp}{"\t"}{.metadata.labels}{"\n"}{end}' | \
  while IFS=$'\t' read -r namespace timestamp labels; do
    if [ -z "$namespace" ] || [ -z "$timestamp" ]; then
      debug_log "Skipping empty namespace or timestamp"
      continue
    fi
    
    debug_log "Processing namespace: $namespace"
    
    age_seconds=$(calculate_age "$timestamp")
    age_display=$(format_age "$age_seconds")
    
    echo "Checking namespace: $namespace (age: $age_display)"
    
    # Check if namespace is exempt from cleanup
    if [[ "$labels" == *"$EXEMPTION_LABEL"* ]]; then
      echo "  → Namespace $namespace is exempt from cleanup (has exemption label), skipping"
      debug_log "Namespace has exemption label: $EXEMPTION_LABEL"
      continue
    fi
    
    # Check if namespace matches allowed patterns
    if ! matches_pattern "$namespace"; then
      echo "  → Namespace $namespace does not match any allowed patterns, skipping"
      continue
    fi
    
    # For matching namespaces, show age comparison with culling threshold
    remaining_seconds=$((MAX_AGE_SECONDS - age_seconds))
    remaining_hours=$((remaining_seconds / 3600))
    threshold_display=$(format_hours_threshold $MAX_AGE_HOURS)
    
    if [ "$remaining_seconds" -gt 0 ]; then
      echo "  → Namespace $namespace has ${remaining_hours}h remaining before cleanup (${age_display} < ${threshold_display} threshold)"
      debug_log "Namespace is within age threshold"
    else
      overdue_hours=$((-remaining_hours))
      echo "  → Namespace $namespace is ${overdue_hours}h overdue for cleanup (${age_display} > ${threshold_display} threshold)"
      debug_log "Namespace exceeds age threshold by ${overdue_hours}h"
    fi
    
    # Check cleanup conditions
    should_cleanup=false
    cleanup_reason=""
    
    # Initialize PR cleanup reason variable
    PR_CLEANUP_REASON=""
    
    # Check age-based cleanup
    if [ "$age_seconds" -gt "$MAX_AGE_SECONDS" ]; then
      should_cleanup=true
      cleanup_reason="age-based (${age_display} > ${threshold_display})"
      debug_log "Age-based cleanup triggered"
    fi
    
    # Check PR-based cleanup if enabled
    if [ "$PR_CLEANUP_ENABLED" = "true" ]; then
      if check_pr_status "$namespace"; then
        should_cleanup=true
        if [ -n "$cleanup_reason" ]; then
          cleanup_reason="$cleanup_reason and PR-based ($PR_CLEANUP_REASON)"
        else
          cleanup_reason="PR-based ($PR_CLEANUP_REASON)"
        fi
        debug_log "PR-based cleanup triggered: $PR_CLEANUP_REASON"
      fi
    fi
    
    if [ "$should_cleanup" = "true" ]; then
      echo "  → Namespace $namespace eligible for cleanup: $cleanup_reason"
      debug_log "Namespace eligible for cleanup with reason: $cleanup_reason"
      
      if [ "$DRY_RUN" = "true" ]; then
        echo "  → DRY RUN: Would clean up namespace $namespace ($cleanup_reason)"
        debug_log "Dry run mode: would clean up namespace"
      else
        debug_log "Performing actual cleanup"
        # Use rdu cleanup command with force flag to avoid interactive prompts
        if command -v rdu >/dev/null 2>&1; then
          echo "  → Using rdu cleanup-deployment for namespace $namespace"
          debug_log "Using rdu for cleanup"
          # Create .kube directory and empty config file to satisfy rdu's expectations
          mkdir -p /home/appuser/.kube
          touch /home/appuser/.kube/config
          # Unset KUBECONFIG to force rdu to use in-cluster config
          unset KUBECONFIG
          echo "yes" | rdu cleanup-deployment --namespace "$namespace" --delete-namespace || {
            echo "  → Warning: rdu cleanup failed for $namespace"
            debug_log "rdu cleanup failed, checking remaining resources"
            
            # Check what resources still exist in the namespace
            echo "  → Checking remaining resources in namespace $namespace:"
            if kubectl get all -n "$namespace" 2>/dev/null | grep -v "^NAME" | grep -v "No resources found"; then
              kubectl get all -n "$namespace" 2>/dev/null || echo "    No standard resources found"
            else
              echo "    No standard resources found"
            fi
            
            # Also check for other common resources
            echo "  → Checking for PVCs, secrets, and configmaps:"
            kubectl get pvc,secrets,configmaps -n "$namespace" 2>/dev/null | grep -v "^NAME" | grep -v "No resources found" || echo "    No PVCs, secrets, or configmaps found"
            
            # Check for any finalizers that might be blocking deletion
            echo "  → Checking namespace finalizers:"
            kubectl get namespace "$namespace" -o jsonpath='{.spec.finalizers}' 2>/dev/null | grep -q . && {
              echo "    Finalizers found: $(kubectl get namespace "$namespace" -o jsonpath='{.spec.finalizers}' 2>/dev/null)"
            } || echo "    No finalizers found"
            
            echo "  → Attempting manual cleanup"
            debug_log "Attempting manual namespace deletion"
            kubectl delete namespace "$namespace" --timeout=300s || echo "  → Failed to delete namespace $namespace"
          }
        else
          echo "  → rdu not available, performing manual cleanup"
          debug_log "rdu not available, using kubectl for cleanup"
          kubectl delete namespace "$namespace" --timeout=300s || echo "  → Failed to delete namespace $namespace"
        fi
        echo "  → Cleanup completed for namespace: $namespace"
        debug_log "Cleanup completed for namespace: $namespace"
      fi
    else
      debug_log "Namespace does not meet cleanup criteria"
    fi
  done

debug_log "Namespace processing completed"
echo "Renku CI deployment cleanup completed"