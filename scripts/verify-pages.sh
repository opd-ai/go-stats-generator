#!/usr/bin/env bash
#
# GitHub Pages Deployment Verification Script
#
# This script checks if GitHub Pages is properly configured and deployed
# for the go-stats-generator WebAssembly application.

set -e

REPO_OWNER="opd-ai"
REPO_NAME="go-stats-generator"
EXPECTED_URL="https://${REPO_OWNER}.github.io/${REPO_NAME}/"

echo "=== GitHub Pages Deployment Verification ==="
echo

# Check if we have curl
if ! command -v curl &> /dev/null; then
    echo "❌ curl is required but not installed"
    exit 1
fi

# Check if we have jq (optional, for better API response parsing)
HAS_JQ=false
if command -v jq &> /dev/null; then
    HAS_JQ=true
fi

echo "📋 Repository: ${REPO_OWNER}/${REPO_NAME}"
echo "🌐 Expected URL: ${EXPECTED_URL}"
echo

# Step 1: Check GitHub Pages API
echo "1️⃣  Checking GitHub Pages API..."
PAGES_API_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/pages"
PAGES_RESPONSE=$(curl -s -w "\n%{http_code}" "${PAGES_API_URL}")
PAGES_HTTP_CODE=$(echo "${PAGES_RESPONSE}" | tail -n 1)
PAGES_BODY=$(echo "${PAGES_RESPONSE}" | sed '$d')

if [ "${PAGES_HTTP_CODE}" -eq 200 ]; then
    echo "   ✅ GitHub Pages is enabled"
    
    if [ "${HAS_JQ}" = true ]; then
        PAGES_URL=$(echo "${PAGES_BODY}" | jq -r '.html_url // empty')
        PAGES_STATUS=$(echo "${PAGES_BODY}" | jq -r '.status // empty')
        PAGES_SOURCE=$(echo "${PAGES_BODY}" | jq -r '.source.type // empty')
        
        if [ -n "${PAGES_URL}" ]; then
            echo "   📍 URL: ${PAGES_URL}"
        fi
        if [ -n "${PAGES_STATUS}" ]; then
            echo "   📊 Status: ${PAGES_STATUS}"
        fi
        if [ -n "${PAGES_SOURCE}" ]; then
            echo "   🔧 Source: ${PAGES_SOURCE}"
            
            if [ "${PAGES_SOURCE}" != "workflow" ]; then
                echo "   ⚠️  Warning: Source should be 'workflow' (GitHub Actions)"
                echo "      Current source: ${PAGES_SOURCE}"
                echo "      Action required: Set source to 'GitHub Actions' in repository settings"
            fi
        fi
    fi
elif [ "${PAGES_HTTP_CODE}" -eq 404 ]; then
    echo "   ❌ GitHub Pages is NOT enabled"
    echo
    echo "   📖 Setup Instructions:"
    echo "      1. Go to: https://github.com/${REPO_OWNER}/${REPO_NAME}/settings/pages"
    echo "      2. Under 'Source', select 'GitHub Actions'"
    echo "      3. Save the changes"
    echo
    echo "   📄 See docs/GITHUB_PAGES_SETUP.md for detailed instructions"
    exit 1
else
    echo "   ⚠️  Unexpected API response: HTTP ${PAGES_HTTP_CODE}"
    echo "      This may indicate a permissions issue or GitHub API problem"
fi

echo

# Step 2: Check if site is accessible
echo "2️⃣  Checking if site is accessible..."
SITE_RESPONSE=$(curl -s -w "\n%{http_code}" -L "${EXPECTED_URL}")
SITE_HTTP_CODE=$(echo "${SITE_RESPONSE}" | tail -n 1)
SITE_BODY=$(echo "${SITE_RESPONSE}" | sed '$d')

if [ "${SITE_HTTP_CODE}" -eq 200 ]; then
    echo "   ✅ Site is accessible at ${EXPECTED_URL}"
    
    # Check for expected content
    if echo "${SITE_BODY}" | grep -q "go-stats-generator"; then
        echo "   ✅ Site contains expected content (go-stats-generator)"
    else
        echo "   ⚠️  Site accessible but may not contain expected content"
    fi
elif [ "${SITE_HTTP_CODE}" -eq 404 ]; then
    echo "   ❌ Site returns 404 Not Found"
    echo "      This may mean:"
    echo "      - GitHub Pages is not enabled"
    echo "      - No successful deployment has occurred yet"
    echo "      - DNS propagation is in progress (wait a few minutes)"
else
    echo "   ⚠️  Unexpected response: HTTP ${SITE_HTTP_CODE}"
fi

echo

# Step 3: Check GitHub Actions workflow
echo "3️⃣  Checking GitHub Actions workflow..."
WORKFLOW_API_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/actions/workflows"
WORKFLOW_RESPONSE=$(curl -s "${WORKFLOW_API_URL}")

if [ "${HAS_JQ}" = true ]; then
    DEPLOY_WORKFLOW=$(echo "${WORKFLOW_RESPONSE}" | jq -r '.workflows[] | select(.name == "Deploy to GitHub Pages") | .state')
    
    if [ -n "${DEPLOY_WORKFLOW}" ]; then
        echo "   ✅ 'Deploy to GitHub Pages' workflow found"
        echo "   📊 State: ${DEPLOY_WORKFLOW}"
    else
        echo "   ⚠️  'Deploy to GitHub Pages' workflow not found"
        echo "      Expected workflow file: .github/workflows/deploy-pages.yml"
    fi
    
    # Check latest workflow run
    RUNS_API_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/actions/workflows/deploy-pages.yml/runs?per_page=1"
    RUNS_RESPONSE=$(curl -s "${RUNS_API_URL}")
    LATEST_RUN_STATUS=$(echo "${RUNS_RESPONSE}" | jq -r '.workflow_runs[0].status // empty')
    LATEST_RUN_CONCLUSION=$(echo "${RUNS_RESPONSE}" | jq -r '.workflow_runs[0].conclusion // empty')
    
    if [ -n "${LATEST_RUN_STATUS}" ]; then
        echo "   📊 Latest run status: ${LATEST_RUN_STATUS}"
        
        if [ "${LATEST_RUN_STATUS}" = "completed" ]; then
            if [ "${LATEST_RUN_CONCLUSION}" = "success" ]; then
                echo "   ✅ Latest deployment: SUCCESS"
            else
                echo "   ❌ Latest deployment: ${LATEST_RUN_CONCLUSION}"
                echo "      Check workflow logs: https://github.com/${REPO_OWNER}/${REPO_NAME}/actions/workflows/deploy-pages.yml"
            fi
        fi
    fi
else
    # Without jq, just check if workflow file exists locally
    if [ -f ".github/workflows/deploy-pages.yml" ]; then
        echo "   ✅ Workflow file exists: .github/workflows/deploy-pages.yml"
    else
        echo "   ❌ Workflow file not found: .github/workflows/deploy-pages.yml"
    fi
fi

echo

# Step 4: Check WASM binary and manifest
echo "4️⃣  Checking WASM binary and manifest..."
MANIFEST_URL="${EXPECTED_URL}wasm/wasm-manifest.json"
MANIFEST_RESPONSE=$(curl -s -w "\n%{http_code}" "${MANIFEST_URL}")
MANIFEST_HTTP_CODE=$(echo "${MANIFEST_RESPONSE}" | tail -n 1)
MANIFEST_BODY=$(echo "${MANIFEST_RESPONSE}" | sed '$d')

if [ "${MANIFEST_HTTP_CODE}" -eq 200 ]; then
    echo "   ✅ wasm-manifest.json found"
    
    if [ "${HAS_JQ}" = true ]; then
        WASM_FILE=$(echo "${MANIFEST_BODY}" | jq -r '.wasmFile // empty')
        
        if [ -n "${WASM_FILE}" ]; then
            echo "   📦 WASM file: ${WASM_FILE}"
            
            # Try to fetch the WASM binary
            WASM_URL="${EXPECTED_URL}wasm/${WASM_FILE}"
            WASM_RESPONSE=$(curl -s -w "\n%{http_code}" -I "${WASM_URL}")
            WASM_HTTP_CODE=$(echo "${WASM_RESPONSE}" | tail -n 1)
            
            if [ "${WASM_HTTP_CODE}" -eq 200 ]; then
                echo "   ✅ WASM binary is accessible"
                
                # Check content type
                CONTENT_TYPE=$(echo "${WASM_RESPONSE}" | grep -i "content-type:" | cut -d' ' -f2- | tr -d '\r')
                if [ -n "${CONTENT_TYPE}" ]; then
                    echo "   📋 Content-Type: ${CONTENT_TYPE}"
                fi
                
                # Check content length
                CONTENT_LENGTH=$(echo "${WASM_RESPONSE}" | grep -i "content-length:" | cut -d' ' -f2- | tr -d '\r')
                if [ -n "${CONTENT_LENGTH}" ]; then
                    SIZE_MB=$(awk "BEGIN {printf \"%.2f\", ${CONTENT_LENGTH}/1024/1024}")
                    echo "   📏 Size: ${SIZE_MB} MB"
                fi
            else
                echo "   ❌ WASM binary not found: HTTP ${WASM_HTTP_CODE}"
            fi
        fi
    fi
else
    echo "   ⚠️  wasm-manifest.json not found (HTTP ${MANIFEST_HTTP_CODE})"
    echo "      This may indicate the site hasn't been deployed yet"
fi

echo

# Summary
echo "=== Summary ==="
echo

if [ "${PAGES_HTTP_CODE}" -eq 200 ] && [ "${SITE_HTTP_CODE}" -eq 200 ] && [ "${MANIFEST_HTTP_CODE}" -eq 200 ]; then
    echo "✅ GitHub Pages deployment is fully operational!"
    echo
    echo "🎉 Visit your site at: ${EXPECTED_URL}"
else
    echo "⚠️  Some checks failed. Review the output above for details."
    echo
    echo "📖 For setup instructions, see: docs/GITHUB_PAGES_SETUP.md"
    echo "🔧 Repository settings: https://github.com/${REPO_OWNER}/${REPO_NAME}/settings/pages"
    echo "🔍 Workflow logs: https://github.com/${REPO_OWNER}/${REPO_NAME}/actions/workflows/deploy-pages.yml"
fi

echo
