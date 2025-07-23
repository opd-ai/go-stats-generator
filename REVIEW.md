# PROJECT: Go Codebase Production Readiness Analysis & Implementation Plan

## OBJECTIVE:
Analyze an existing Go codebase and create a comprehensive, actionable plan to transform it into a production-ready system. This includes identifying gaps in code quality, security, performance, observability, and operational readiness, then providing specific implementation steps to address each issue. Place the plan into a ROADMAP.md file and present it in a single unbroken ~~~~ codeblock.

## TECHNICAL SPECIFICATIONS:
- Language: Go
- Type: Analysis and Remediation Plan
- Scope: Complete codebase assessment (architecture, code quality, dependencies, deployment)
- Output: Prioritized implementation roadmap
- Timeline: Phased approach with clear milestones and effort estimates

## ANALYSIS FRAMEWORK:

### Phase 1: Codebase Assessment
**Required Actions:**
1. **Code Quality Analysis:**
   - Run `go mod tidy` and analyze dependencies for security vulnerabilities and maintenance status
   - Map current directory structure against Go community standards
   - Identify overly complex functions and deeply nested code
   - Check for consistent error handling patterns and proper resource management

2. **Testing & Coverage Analysis:**
   - Generate test coverage reports and identify critical paths without adequate testing
   - Assess test quality, organization, and coverage of edge cases
   - Find missing mock implementations for external dependencies

3. **Security Assessment:**
   - Check for hardcoded credentials and sensitive data exposure
   - Identify potential injection vulnerabilities and missing input validation
   - Review authentication and authorization implementations
   - Assess secure communication practices within application boundaries

**IMPORTANT SECURITY SCOPE NOTE:**
- **DO NOT recommend TLS, HTTPS, or transport-layer encryption** if not already present
- **DO NOT suggest certificate management or SSL/TLS configuration**
- Focus only on application-layer security concerns (input validation, authentication, authorization, data sanitization)
- Transport security is assumed to be handled by reverse proxies, load balancers, or deployment infrastructure

## IMPLEMENTATION PLAN:

### PHASE 1: Critical Foundation (High Priority)
**Focus: Essential production requirements**

#### Task 1.1: Application Security and Error Handling
```go
// Implementation Requirements:
1. Externalize all configuration and secrets
2. Implement comprehensive input validation
3. Establish consistent error handling patterns
4. Add timeout handling for all external operations

// Required Pattern:
func ProcessData(ctx context.Context, input string) error {
    if input == "" {
        return fmt.Errorf("invalid input: empty string")
    }
    
    result, err := externalService.Process(ctx, input)
    if err != nil {
        return fmt.Errorf("processing failed for input %q: %w", input, err)
    }
    
    return nil
}
```

#### Task 1.2: Observability Foundation
```go
// Implementation Requirements:
1. Implement structured logging throughout application
2. Add application metrics and health indicators
3. Create health check endpoints
4. Implement request correlation for debugging

// Logging Pattern:
logger.Info("processing request",
    zap.String("operation", "process_data"),
    zap.String("request_id", requestID),
    zap.Duration("duration", time.Since(start)),
)
```

### PHASE 2: Performance & Reliability (Medium Priority)
**Focus: Production resilience**

#### Task 2.1: Resource Management
```go
// Implementation Requirements:
1. Implement connection pooling for external services
2. Add circuit breaker patterns for external dependencies
3. Configure appropriate timeouts and retry logic
4. Optimize resource usage and cleanup

// Resource Management:
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

resource, err := pool.Acquire(ctx)
if err != nil {
    return fmt.Errorf("failed to acquire resource: %w", err)
}
defer resource.Close()
```

#### Task 2.2: Configuration Management
```go
// Implementation Requirements:
1. Centralize configuration management
2. Implement configuration validation at startup
3. Support environment-specific configurations
4. Document all configuration options
```

### PHASE 3: Operational Excellence (Lower Priority)
**Focus: Long-term maintainability**

#### Task 3.1: Testing and Quality
```go
// Implementation Requirements:
1. Achieve comprehensive test coverage for critical paths
2. Implement integration testing for external dependencies
3. Add performance testing for critical operations
4. Establish code quality standards and automation
```

## VALIDATION CHECKLIST:

### Application Security Requirements:
- [ ] No hardcoded secrets or credentials
- [ ] Input validation for all external data
- [ ] Proper authentication and authorization within application
- [ ] No sensitive data in logs or error messages
- [ ] SQL injection prevention through parameterized queries
- [ ] XSS prevention through proper output encoding

### Reliability Requirements:
- [ ] Comprehensive error handling with context
- [ ] Circuit breakers for external dependencies
- [ ] Appropriate timeout configurations
- [ ] Graceful shutdown and resource cleanup

### Performance Requirements:
- [ ] Connection pooling for external services
- [ ] No blocking operations without timeouts
- [ ] Resource limits and memory management
- [ ] Performance monitoring and profiling

### Observability Requirements:
- [ ] Structured logging with correlation IDs
- [ ] Application metrics and health indicators
- [ ] Health check endpoints
- [ ] Error tracking and alerting

### Testing Requirements:
- [ ] Unit tests for business logic
- [ ] Integration tests for external dependencies
- [ ] Performance tests for critical operations
- [ ] Test data management and isolation

### Deployment Requirements:
- [ ] Environment-specific configuration
- [ ] Automated deployment procedures
- [ ] Resource limits and monitoring
- [ ] Rollback capabilities

## OUTPUT FORMAT:

```markdown
# PRODUCTION READINESS ASSESSMENT: [Codebase Name]

## CRITICAL ISSUES
### Application Security Concerns:
- [Specific application-layer security issues with location references]
- [Note: Transport security (TLS/HTTPS) is outside scope - assumed handled by infrastructure]

### Reliability Concerns:
- [Specific issues with location references]

### Performance Concerns:
- [Specific issues with location references]

## IMPLEMENTATION ROADMAP

### Phase 1: Foundation
**Duration:** [timeframe]
**Tasks:**
1. [Specific task with acceptance criteria]
2. [Specific task with acceptance criteria]

### Phase 2: Performance & Reliability
**Duration:** [timeframe]
**Tasks:**
1. [Specific task with acceptance criteria]
2. [Specific task with acceptance criteria]

### Phase 3: Operational Excellence
**Duration:** [timeframe]
**Tasks:**
1. [Specific task with acceptance criteria]

## RECOMMENDED LIBRARIES
[Specific recommendations with justification]

## SUCCESS CRITERIA
[Measurable production readiness indicators]

## RISK ASSESSMENT
[Potential risks and mitigation strategies]

## SECURITY SCOPE CLARIFICATION
- Analysis focuses on application-layer security only
- Transport encryption (TLS/HTTPS) assumed to be handled by deployment infrastructure
- No recommendations for certificate management or SSL/TLS configuration
```

## LIBRARY SELECTION GUIDANCE:
**Prefer well-established, actively maintained libraries:**

- Configuration: Choose based on complexity (stdlib, viper, etc.)
- Logging: Structured logging solutions (zerolog, slog)
- HTTP: Evaluate stdlib vs framework needs
- Database: Consider query patterns and ORM requirements
- Testing: Community standard frameworks
- Monitoring: Standard observability patterns

**Document selection rationale based on:**
- Maintenance activity and community health
- Documentation quality and completeness
- Performance characteristics
- Integration complexity
- Long-term support considerations

**TRANSPORT SECURITY EXCLUSION:**
- Do not recommend TLS/SSL libraries or configuration
- Do not suggest HTTPS enforcement at application level
- Do not recommend certificate management solutions
- Assume transport security is handled by reverse proxies, load balancers, or container orchestration platforms