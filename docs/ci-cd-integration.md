# CI/CD Integration Guide

## Overview

`go-stats-generator` provides quality gates for continuous integration pipelines, allowing teams to enforce code quality thresholds and prevent regressions. This guide demonstrates how to integrate the tool into common CI/CD systems.

## Key Features for CI/CD

- **Quality Gate Enforcement**: Exit with non-zero code when thresholds are violated
- **Baseline Comparison**: Track metrics over time and detect regressions
- **Configurable Thresholds**: Customize limits per project or team standards
- **Multiple Output Formats**: JSON for machine parsing, console for human review

## Quick Start

### Basic Quality Gate

```bash
# Analyze code with enforcement enabled
go-stats-generator analyze . \
  --max-function-length 30 \
  --max-complexity 10 \
  --max-duplication-ratio 0.10 \
  --max-burden-score 50 \
  --enforce-thresholds

# Exit code:
#   0 = all checks pass
#   1 = quality gate violation
#   2 = analysis error
```

### Baseline Comparison Workflow

```bash
# Step 1: Establish baseline (typically on main branch)
go-stats-generator analyze . --format json --output baseline.json

# Step 2: Analyze current code (on feature branch)
go-stats-generator analyze . --format json --output current.json

# Step 3: Compare and enforce no regressions
go-stats-generator diff baseline.json current.json --enforce-thresholds
```

## Recommended Thresholds

### For New Codebases (Greenfield Projects)

Use strict thresholds from day one to maintain high quality:

```bash
go-stats-generator analyze . \
  --max-function-length 30 \
  --max-complexity 10 \
  --max-duplication-ratio 0.05 \
  --max-burden-score 40 \
  --min-doc-coverage 0.80 \
  --max-undocumented-exports 5 \
  --enforce-thresholds
```

**Rationale**: Starting strict prevents technical debt accumulation.

### For Legacy Codebases

Use progressive tightening strategy:

#### Phase 1: Establish Current State (Month 1)
```bash
# Analyze without enforcement to understand baseline
go-stats-generator analyze . --format json --output baseline.json

# Review metrics and set realistic initial thresholds
# Example: If current duplication is 25%, start at 30% and reduce over time
```

#### Phase 2: Prevent New Violations (Months 2-3)
```bash
go-stats-generator analyze . \
  --max-function-length 50 \
  --max-complexity 15 \
  --max-duplication-ratio 0.30 \
  --max-burden-score 70 \
  --enforce-thresholds
```

#### Phase 3: Gradual Improvement (Months 4-12)
```bash
# Tighten thresholds every sprint/month as team improves code
# Month 4:
go-stats-generator analyze . --max-duplication-ratio 0.25 --enforce-thresholds

# Month 6:
go-stats-generator analyze . --max-duplication-ratio 0.20 --enforce-thresholds

# Month 9:
go-stats-generator analyze . --max-duplication-ratio 0.15 --enforce-thresholds

# Month 12: Reach target
go-stats-generator analyze . --max-duplication-ratio 0.10 --enforce-thresholds
```

**Rationale**: Prevents new technical debt while allowing teams to refactor incrementally.

## CI/CD Platform Examples

### GitHub Actions

#### Basic Quality Gate

```yaml
name: Code Quality Gate

on:
  pull_request:
    branches: [ main, develop ]
  push:
    branches: [ main ]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install go-stats-generator
        run: go install github.com/opd-ai/go-stats-generator@latest

      - name: Run quality analysis
        run: |
          go-stats-generator analyze . \
            --max-function-length 30 \
            --max-complexity 10 \
            --max-duplication-ratio 0.10 \
            --max-burden-score 50 \
            --enforce-thresholds

      - name: Upload analysis report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: quality-report
          path: |
            *.json
            *.html
```

#### Baseline Comparison with Main Branch

```yaml
name: Quality Regression Check

on:
  pull_request:
    branches: [ main ]

jobs:
  regression-check:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout PR code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install go-stats-generator
        run: go install github.com/opd-ai/go-stats-generator@latest

      - name: Analyze current PR
        run: |
          go-stats-generator analyze . \
            --format json \
            --output pr-metrics.json

      - name: Checkout main branch
        uses: actions/checkout@v4
        with:
          ref: main
          path: main-branch

      - name: Analyze main branch baseline
        run: |
          cd main-branch
          go-stats-generator analyze . \
            --format json \
            --output ../main-metrics.json

      - name: Compare and enforce no regressions
        run: |
          go-stats-generator diff main-metrics.json pr-metrics.json \
            --enforce-thresholds

      - name: Generate HTML diff report
        if: always()
        run: |
          go-stats-generator diff main-metrics.json pr-metrics.json \
            --format html \
            --output diff-report.html

      - name: Upload diff report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: regression-report
          path: diff-report.html
```

#### Trend Tracking with Artifacts

```yaml
name: Quality Trend Tracking

on:
  push:
    branches: [ main ]

jobs:
  track-trends:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install go-stats-generator
        run: go install github.com/opd-ai/go-stats-generator@latest

      - name: Download previous baseline
        uses: actions/download-artifact@v4
        with:
          name: quality-baseline
          path: .
        continue-on-error: true

      - name: Save current snapshot
        run: |
          go-stats-generator baseline save \
            --name "commit-${{ github.sha }}" \
            --description "Automated snapshot for commit ${{ github.sha }}"

      - name: Generate trend report
        run: |
          go-stats-generator trend --days 90 > trend-report.txt
          cat trend-report.txt

      - name: Upload baseline database
        uses: actions/upload-artifact@v4
        with:
          name: quality-baseline
          path: metrics.db
```

### GitLab CI

#### `.gitlab-ci.yml`

```yaml
stages:
  - quality

variables:
  GO_VERSION: "1.23"

quality-gate:
  stage: quality
  image: golang:${GO_VERSION}
  before_script:
    - go install github.com/opd-ai/go-stats-generator@latest
  script:
    - |
      go-stats-generator analyze . \
        --max-function-length 30 \
        --max-complexity 10 \
        --max-duplication-ratio 0.10 \
        --max-burden-score 50 \
        --enforce-thresholds
  artifacts:
    when: always
    paths:
      - "*.json"
      - "*.html"
    expire_in: 30 days

regression-check:
  stage: quality
  image: golang:${GO_VERSION}
  before_script:
    - go install github.com/opd-ai/go-stats-generator@latest
  script:
    # Analyze current branch
    - go-stats-generator analyze . --format json --output current.json
    
    # Fetch and analyze main branch
    - git fetch origin main
    - git checkout origin/main
    - go-stats-generator analyze . --format json --output baseline.json
    - git checkout -
    
    # Compare and enforce
    - |
      go-stats-generator diff baseline.json current.json \
        --format html \
        --output diff-report.html \
        --enforce-thresholds
  artifacts:
    when: always
    paths:
      - diff-report.html
    reports:
      # GitLab can parse JSON for quality metrics
      codequality: current.json
  only:
    - merge_requests
```

### Jenkins

#### Jenkinsfile

```groovy
pipeline {
    agent any
    
    environment {
        GO_VERSION = '1.23'
    }
    
    stages {
        stage('Setup') {
            steps {
                sh '''
                    # Install Go if not present
                    which go || {
                        wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
                        sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
                        export PATH=$PATH:/usr/local/go/bin
                    }
                    
                    # Install go-stats-generator
                    go install github.com/opd-ai/go-stats-generator@latest
                '''
            }
        }
        
        stage('Quality Gate') {
            steps {
                sh '''
                    $HOME/go/bin/go-stats-generator analyze . \
                        --max-function-length 30 \
                        --max-complexity 10 \
                        --max-duplication-ratio 0.10 \
                        --max-burden-score 50 \
                        --enforce-thresholds
                '''
            }
        }
        
        stage('Regression Check') {
            when {
                changeRequest()
            }
            steps {
                sh '''
                    # Analyze current PR
                    $HOME/go/bin/go-stats-generator analyze . \
                        --format json \
                        --output pr-metrics.json
                    
                    # Analyze main branch
                    git checkout main
                    $HOME/go/bin/go-stats-generator analyze . \
                        --format json \
                        --output main-metrics.json
                    git checkout -
                    
                    # Compare
                    $HOME/go/bin/go-stats-generator diff \
                        main-metrics.json pr-metrics.json \
                        --format html \
                        --output diff-report.html \
                        --enforce-thresholds
                '''
            }
        }
        
        stage('Publish Reports') {
            steps {
                publishHTML([
                    allowMissing: false,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: '.',
                    reportFiles: 'diff-report.html',
                    reportName: 'Code Quality Report'
                ])
                
                archiveArtifacts artifacts: '*.json,*.html', allowEmptyArchive: true
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
    }
}
```

## Configuration File Approach

For teams managing multiple thresholds, use a configuration file instead of command-line flags.

### `.go-stats-generator.yaml`

```yaml
# Global settings
analysis:
  skip_tests: true
  skip_vendor: true
  max_files: 50000

# Quality thresholds
thresholds:
  max_function_length: 30
  max_complexity: 10
  min_doc_coverage: 0.70
  similarity_threshold: 0.80
  min_block_lines: 6

# Duplication settings
duplication:
  max_ratio: 0.10

# Organization settings  
organization:
  max_file_lines: 500
  max_package_files: 20

# Scoring thresholds
scoring:
  max_burden_score: 50
  max_undocumented_exports: 10

# Regression detection
baseline:
  enforce_no_regressions: true
  allow_function_growth: false
  allow_complexity_growth: false
```

Then run simply:

```bash
# Reads .go-stats-generator.yaml from current directory
go-stats-generator analyze . --enforce-thresholds
```

## Advanced Patterns

### Enforce Different Thresholds Per Package

```bash
# Strict thresholds for core business logic
go-stats-generator analyze ./internal/core \
  --max-complexity 8 \
  --max-burden-score 40 \
  --enforce-thresholds

# More lenient for infrastructure code
go-stats-generator analyze ./internal/infrastructure \
  --max-complexity 12 \
  --max-burden-score 60 \
  --enforce-thresholds
```

### Generate Trend Reports on Schedule

```yaml
# GitHub Actions - Daily trend tracking
name: Daily Quality Trends

on:
  schedule:
    - cron: '0 2 * * *'  # 2 AM UTC daily
  workflow_dispatch:  # Manual trigger

jobs:
  track-trends:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Install go-stats-generator
        run: go install github.com/opd-ai/go-stats-generator@latest
      
      - name: Restore baseline database
        uses: actions/cache@v4
        with:
          path: metrics.db
          key: quality-baseline-${{ github.ref }}
      
      - name: Save daily snapshot
        run: |
          go-stats-generator baseline save \
            --name "daily-$(date +%Y-%m-%d)" \
            --description "Automated daily snapshot"
      
      - name: Generate 90-day trend report
        run: |
          go-stats-generator trend --days 90 \
            --format markdown \
            --output trend-report.md
      
      - name: Post trend report as issue
        if: github.event_name == 'schedule'
        uses: peter-evans/create-issue-from-file@v5
        with:
          title: Weekly Quality Trend Report
          content-filepath: trend-report.md
          labels: quality, automated
      
      - name: Save baseline database
        uses: actions/cache@v4
        with:
          path: metrics.db
          key: quality-baseline-${{ github.ref }}-${{ github.run_id }}
```

### Block PRs with Quality Regressions

```yaml
name: PR Quality Check

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Install go-stats-generator
        run: go install github.com/opd-ai/go-stats-generator@latest
      
      - name: Analyze PR
        id: pr-analysis
        run: |
          go-stats-generator analyze . \
            --format json \
            --output pr-metrics.json \
            --enforce-thresholds
        continue-on-error: true
      
      - name: Fetch main baseline
        run: |
          git fetch origin main:main
          git checkout main
          go-stats-generator analyze . \
            --format json \
            --output main-metrics.json
          git checkout -
      
      - name: Detect regressions
        id: regression-check
        run: |
          go-stats-generator diff main-metrics.json pr-metrics.json \
            --format json \
            --output diff.json \
            --enforce-thresholds
        continue-on-error: true
      
      - name: Comment on PR
        if: steps.regression-check.outcome == 'failure'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const diff = JSON.parse(fs.readFileSync('diff.json', 'utf8'));
            
            let comment = '## ⚠️ Code Quality Regression Detected\n\n';
            comment += 'The following quality metrics have regressed:\n\n';
            
            if (diff.regressions && diff.regressions.length > 0) {
              diff.regressions.forEach(r => {
                comment += `- **${r.category}**: ${r.description}\n`;
              });
            }
            
            comment += '\n\nPlease address these issues before merging.';
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
      
      - name: Fail if regressions detected
        if: steps.regression-check.outcome == 'failure'
        run: exit 1
```

## Monitoring and Alerting

### Send Trend Alerts to Slack

```bash
#!/bin/bash
# save-and-alert.sh

# Save baseline snapshot
go-stats-generator baseline save \
  --name "nightly-$(date +%Y-%m-%d)" \
  --description "Nightly snapshot"

# Generate trend report
TREND_OUTPUT=$(go-stats-generator trend --days 30 --format json)

# Extract key metrics
MBI_TREND=$(echo "$TREND_OUTPUT" | jq -r '.mbi_score_trend')
DUP_TREND=$(echo "$TREND_OUTPUT" | jq -r '.duplication_trend')

# Alert if negative trend
if [[ "$MBI_TREND" == "increasing" ]] || [[ "$DUP_TREND" == "increasing" ]]; then
  curl -X POST -H 'Content-type: application/json' \
    --data "{
      \"text\": \"⚠️ Code quality degrading: MBI trend=$MBI_TREND, Duplication trend=$DUP_TREND\"
    }" \
    $SLACK_WEBHOOK_URL
fi
```

### Dashboard Integration

Export metrics to Prometheus/Grafana:

```bash
# Generate metrics in Prometheus format
go-stats-generator analyze . --format json | jq -r '
  "# HELP code_complexity_avg Average cyclomatic complexity",
  "# TYPE code_complexity_avg gauge",
  ("code_complexity_avg " + (.complexity.average|tostring)),
  "",
  "# HELP code_duplication_ratio Duplication ratio",
  "# TYPE code_duplication_ratio gauge", 
  ("code_duplication_ratio " + (.duplication.ratio|tostring)),
  "",
  "# HELP code_doc_coverage Documentation coverage",
  "# TYPE code_doc_coverage gauge",
  ("code_doc_coverage " + (.documentation.overall_coverage|tostring))
' > metrics.prom

# Push to Pushgateway
cat metrics.prom | curl --data-binary @- http://pushgateway:9091/metrics/job/code-quality
```

## Troubleshooting

### Exit Code 1 But No Violations Shown

**Cause**: `--enforce-thresholds` is set but analysis found violations.

**Solution**: Run without `--enforce-thresholds` first to see which metrics are failing:

```bash
go-stats-generator analyze . \
  --max-complexity 10 \
  --max-duplication-ratio 0.10

# Review output, then add enforcement
go-stats-generator analyze . \
  --max-complexity 10 \
  --max-duplication-ratio 0.10 \
  --enforce-thresholds
```

### Baseline Comparison Shows False Positives

**Cause**: Baseline was created with different flags (e.g., `--skip-tests`).

**Solution**: Always use consistent flags when creating baseline and current snapshots:

```bash
# Baseline
go-stats-generator analyze . --skip-tests --format json --output baseline.json

# Current (must also use --skip-tests)
go-stats-generator analyze . --skip-tests --format json --output current.json

# Diff
go-stats-generator diff baseline.json current.json
```

### Trend Command Shows No Data

**Cause**: No snapshots have been saved to the baseline database.

**Solution**: Save at least 2 snapshots before running trend:

```bash
go-stats-generator baseline save --name "snapshot-1"
# (make some code changes)
go-stats-generator baseline save --name "snapshot-2"

# Now trends available
go-stats-generator trend --days 30
```

## Best Practices

1. **Start with Observation**: Run analysis without `--enforce-thresholds` for 1-2 sprints to understand current state
2. **Use Configuration Files**: Commit `.go-stats-generator.yaml` to version control for team consistency
3. **Progressive Tightening**: For legacy code, tighten thresholds incrementally (quarterly or monthly)
4. **Per-Package Thresholds**: Allow different standards for test/mock code vs production code
5. **Automate Trend Tracking**: Save baselines on every merge to main for historical analysis
6. **Combine with Code Review**: Use diff reports as discussion points in PR reviews
7. **Document Threshold Rationale**: Add comments in config files explaining why thresholds were chosen
8. **Monitor Trends**: Set up alerts when quality metrics degrade over time
9. **Team Buy-In**: Involve the team in setting thresholds; don't impose arbitrary limits
10. **Exemptions Process**: Define how to handle legitimate exceptions (complex algorithms, generated code)

## Reference

### All Available Thresholds

| Flag | Default | Description |
|------|---------|-------------|
| `--max-function-length` | 50 | Maximum lines per function |
| `--max-complexity` | 10 | Maximum cyclomatic complexity |
| `--max-burden-score` | 70 | Maximum MBI burden score |
| `--max-duplication-ratio` | 0.10 | Maximum code duplication (0.0-1.0) |
| `--max-undocumented-exports` | 10 | Maximum undocumented exported symbols |
| `--min-doc-coverage` | 0.70 | Minimum documentation coverage (0.0-1.0) |
| `--similarity-threshold` | 0.80 | Clone detection sensitivity (0.0-1.0) |
| `--min-block-lines` | 6 | Minimum lines for duplication detection |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, all quality gates passed |
| 1 | Quality gate violation (when `--enforce-thresholds` set) |
| 2 | Analysis error (parse failure, file access issues) |

---

**Version**: 1.0.0  
**Last Updated**: 2026-03-03  
**Feedback**: Report issues or suggest improvements at https://github.com/opd-ai/go-stats-generator/issues
