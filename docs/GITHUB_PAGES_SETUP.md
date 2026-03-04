# GitHub Pages Deployment Setup

This document provides step-by-step instructions for enabling GitHub Pages deployment for the `go-stats-generator` WebAssembly application.

## Prerequisites

- Repository admin access to `opd-ai/go-stats-generator`
- GitHub Actions enabled for the repository
- `.github/workflows/deploy-pages.yml` workflow file committed to the `main` branch (✅ already present)

## One-Time Configuration

GitHub Pages must be configured to use GitHub Actions as the deployment source. This is a **one-time manual step** that cannot be automated via code.

### Step 1: Navigate to Repository Settings

1. Go to the GitHub repository: https://github.com/opd-ai/go-stats-generator
2. Click **Settings** (requires admin access)
3. In the left sidebar, scroll down to **Code and automation** section
4. Click **Pages**

### Step 2: Configure Source

On the GitHub Pages settings page:

1. Under **Build and deployment** section
2. Find the **Source** dropdown
3. Select **GitHub Actions** (NOT "Deploy from a branch")
4. No additional configuration is required - GitHub will automatically detect the `deploy-pages.yml` workflow

### Step 3: Verify Configuration

After selecting GitHub Actions as the source:

1. Navigate to **Actions** tab in the repository
2. You should see the "Deploy to GitHub Pages" workflow listed
3. Trigger a manual run:
   - Click on "Deploy to GitHub Pages" workflow
   - Click "Run workflow" button
   - Select `main` branch
   - Click "Run workflow"

Alternatively, push any commit to the `main` branch to trigger automatic deployment.

### Step 4: Monitor First Deployment

1. Go to **Actions** tab
2. Click on the running "Deploy to GitHub Pages" workflow
3. Monitor the build steps:
   - Set up Go
   - Build WASM binary
   - Copy static assets
   - Optimize WASM (optional, may be skipped)
   - Upload artifact
   - Deploy to GitHub Pages

First deployment typically takes 2-3 minutes.

### Step 5: Access Deployed Site

Once deployment succeeds:

1. The workflow output will show the deployment URL
2. Typically: `https://opd-ai.github.io/go-stats-generator/`
3. Visit the URL to verify the application loads

You should see:
- The `go-stats-generator` web interface
- Input fields for repository URL, ref, and GitHub token
- "Analyze" button
- Footer with rate limit status

## Verification Checklist

Use this checklist to confirm successful deployment:

- [ ] GitHub Pages source is set to "GitHub Actions" in repository settings
- [ ] "Deploy to GitHub Pages" workflow appears in Actions tab
- [ ] Workflow runs successfully on push to `main` branch
- [ ] Deployment produces a site URL (e.g., `https://opd-ai.github.io/go-stats-generator/`)
- [ ] Site loads in browser without errors
- [ ] Browser console shows no JavaScript errors
- [ ] WASM binary loads successfully (check Network tab)
- [ ] Entering a GitHub repository URL and clicking "Analyze" works
- [ ] HTML report renders correctly after analysis

## Troubleshooting

### Workflow Fails: "pages build and deployment"

**Symptom**: The "Deploy to GitHub Pages" workflow step fails with permissions error.

**Solution**: The workflow requires specific permissions. Verify in repository Settings → Actions → General → Workflow permissions:
- Set to "Read and write permissions"
- Enable "Allow GitHub Actions to create and approve pull requests"

### Site Shows 404 Error

**Symptom**: Visiting the GitHub Pages URL shows "404 There isn't a GitHub Pages site here."

**Solution**:
1. Verify GitHub Pages source is set to "GitHub Actions" (not branch-based)
2. Check that at least one workflow run completed successfully
3. Wait 1-2 minutes after deployment - GitHub Pages may have propagation delay

### WASM Binary Fails to Load

**Symptom**: Browser console shows error loading `.wasm` file.

**Solution**:
1. Check Network tab - verify `go-stats-generator.<hash>.wasm` returns 200 OK
2. Verify `wasm-manifest.json` exists and contains correct filename
3. Check if file is being served with correct MIME type (`application/wasm`)
4. Clear browser cache and hard reload (Ctrl+Shift+R)

### Rate Limit Errors on Analysis

**Symptom**: Analysis fails immediately with "Rate limit exceeded" error.

**Solution**:
1. GitHub API has a 60 requests/hour limit for unauthenticated requests
2. Provide a GitHub Personal Access Token in the UI
3. Create token at: https://github.com/settings/tokens (no special scopes needed for public repos)
4. Enter token in "GitHub Token" field before clicking "Analyze"

## Automated Deployment

After initial setup, deployment is fully automated:

1. **On every push to `main`**: Workflow automatically rebuilds and deploys
2. **Manual trigger**: Click "Run workflow" in Actions tab to deploy on-demand
3. **Content hashing**: WASM binary filename includes SHA-256 hash for cache-busting
4. **Optimization**: Optional `wasm-opt` step reduces binary size (skipped if not available)

## Custom Domain (Optional)

To use a custom domain instead of `opd-ai.github.io`:

1. Add a `CNAME` file to `web/` directory with your domain (e.g., `stats.example.com`)
2. Update GitHub Pages settings with custom domain
3. Configure DNS records at your domain registrar:
   - CNAME record pointing to `opd-ai.github.io`
4. Wait for DNS propagation and GitHub's SSL certificate provisioning

## Maintenance

### Updating the Application

Changes to the WebAssembly application are automatically deployed when merged to `main`:

1. Make changes to `cmd/wasm/`, `web/`, or analyzer code
2. Commit and push to a feature branch
3. Create pull request to `main`
4. After merge, GitHub Actions automatically rebuilds and deploys

### Monitoring Deployments

- **Actions tab**: View deployment history and logs
- **Environments**: Settings → Environments → `github-pages` shows deployment history
- **Status badge**: Add to README with: `[![Deploy](https://github.com/opd-ai/go-stats-generator/actions/workflows/deploy-pages.yml/badge.svg)](https://github.com/opd-ai/go-stats-generator/actions/workflows/deploy-pages.yml)`

### Rolling Back

To roll back to a previous version:

1. Go to Actions → Deploy to GitHub Pages
2. Find the successful deployment you want to restore
3. Click "Re-run all jobs"
4. GitHub Pages will deploy that specific commit's artifacts

## Security Considerations

- **No secrets in WASM**: The WASM binary is publicly downloadable - never include API keys or secrets
- **Token handling**: User-provided GitHub tokens are stored in `sessionStorage` only (cleared on page close)
- **CORS**: GitHub API supports CORS for browser requests - no proxy needed
- **Content Security Policy**: Consider adding CSP headers via `_headers` file for enhanced security

## Further Reading

- [GitHub Pages documentation](https://docs.github.com/en/pages)
- [GitHub Actions for Pages](https://github.com/actions/deploy-pages)
- [WebAssembly with Go](https://github.com/golang/go/wiki/WebAssembly)
- [WASM optimization with wasm-opt](https://github.com/WebAssembly/binaryen)
