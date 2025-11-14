# StreamSpace Bug Bounty Program

**Program Status**: Active
**Last Updated**: 2025-11-14
**Program Type**: Private (invite-only initially, public launch planned for Q2 2025)

---

## Table of Contents

- [Program Overview](#program-overview)
- [Scope](#scope)
- [Rewards](#rewards)
- [Rules of Engagement](#rules-of-engagement)
- [Submission Guidelines](#submission-guidelines)
- [Disclosure Policy](#disclosure-policy)
- [Safe Harbor](#safe-harbor)
- [Hall of Fame](#hall-of-fame)

---

## Program Overview

Welcome to the StreamSpace Bug Bounty Program! We believe that working with skilled security researchers is crucial to keeping our users safe. This program rewards security researchers who discover and responsibly disclose security vulnerabilities in StreamSpace.

### Program Goals

1. **Identify and fix security vulnerabilities** before they can be exploited
2. **Foster collaboration** with the security research community
3. **Improve our security posture** through external validation
4. **Recognize and reward** researchers who help protect our users

### Contact Information

**Security Team Email**: security@streamspace.io
**PGP Key**: Available at https://streamspace.io/.well-known/pgp-key.txt
**Response Time SLA**:
- Initial acknowledgment: Within 24 hours
- Triage and validation: Within 72 hours
- Regular updates: Every 5 business days

---

## Scope

### In-Scope Assets

The following assets are **in scope** for the bug bounty program:

#### 1. Web Application
- **URL**: https://app.streamspace.io
- **Components**:
  - User authentication and authorization
  - Session management
  - API endpoints (`/api/v1/*`)
  - WebSocket connections
  - Admin panel
  - Plugin system

#### 2. API Backend
- **Base URL**: https://api.streamspace.io
- **Components**:
  - REST API endpoints
  - WebSocket proxy
  - Authentication middleware
  - Database interactions
  - File upload/download functionality

#### 3. Source Code Repositories
- **GitHub**: https://github.com/JoshuaAFerguson/streamspace
- **Components**:
  - Go backend code (`api/`, `controller/`)
  - TypeScript/React frontend (`ui/`)
  - Kubernetes manifests (`manifests/`)
  - CI/CD workflows (`.github/workflows/`)

#### 4. Container Images
- **Registry**: ghcr.io/streamspace/*
- **Images**:
  - streamspace-api
  - streamspace-controller
  - streamspace-ui

#### 5. Infrastructure (Limited)
- **In Scope**:
  - Kubernetes configuration security
  - Service mesh (Istio) misconfigurations
  - Network policy bypasses
  - Container escape vulnerabilities

### Out-of-Scope Assets

The following are **explicitly out of scope**:

❌ **Third-Party Services**:
- GitHub.com infrastructure
- Authentik/Keycloak identity providers
- Cloud provider infrastructure (AWS, GCP, Azure)
- DNS providers, CDN services
- Email service providers

❌ **Physical Security**:
- Physical access to offices or data centers
- Social engineering of employees
- Physical theft or destruction

❌ **Denial of Service**:
- Network-level DoS attacks
- Application-level DoS (unless demonstrating a unique technique with minimal impact)
- Resource exhaustion attacks exceeding testing limits

❌ **Non-Security Issues**:
- Functional bugs without security impact
- UI/UX issues
- Feature requests
- Performance issues

❌ **Known Issues**:
- Issues already documented in our public issue tracker
- Vulnerabilities in outdated versions (must test against latest release)
- Third-party dependencies (report directly to the upstream project)

### Exclusions

The following vulnerabilities are **not eligible** for rewards:

- Issues requiring physical access
- Social engineering attacks
- DoS/DDoS attacks
- Issues in third-party applications or websites
- Spam or social engineering of users
- Issues already known to StreamSpace
- Issues reported by employees or contractors
- Duplicate submissions (first reporter wins)

---

## Rewards

### Bounty Tiers

We offer monetary rewards based on the severity and impact of the vulnerability:

| Severity | Description | Bounty Range | Examples |
|----------|-------------|--------------|----------|
| **Critical** | Vulnerabilities that allow complete system compromise or data breach | **$2,000 - $10,000** | - Remote code execution (RCE)<br>- Authentication bypass<br>- SQL injection leading to full database access<br>- Container escape to host |
| **High** | Vulnerabilities that allow unauthorized access to sensitive data or functionality | **$500 - $2,000** | - Privilege escalation (user → admin)<br>- Stored XSS in admin panel<br>- IDOR accessing other users' sessions<br>- JWT secret disclosure |
| **Medium** | Vulnerabilities with limited impact or requiring user interaction | **$100 - $500** | - Reflected XSS<br>- CSRF on state-changing operations<br>- Information disclosure<br>- Rate limiting bypass |
| **Low** | Vulnerabilities with minimal security impact | **$50 - $100** | - Missing security headers<br>- CORS misconfigurations<br>- Self-XSS<br>- Verbose error messages |
| **Informational** | Security best practice recommendations without immediate exploitation | **Swag + Recognition** | - Security recommendations<br>- Defense-in-depth suggestions<br>- Code quality issues |

### Bonus Multipliers

We offer **bonus rewards** for exceptional submissions:

- **+50%**: High-quality write-up with clear reproduction steps and suggested fix
- **+25%**: Proof-of-concept (PoC) exploit code demonstrating impact
- **+25%**: First vulnerability of a new class (e.g., first SQL injection found)
- **+20%**: Creative attack chain combining multiple vulnerabilities
- **+10%**: Providing a working code patch

**Example**: Critical RCE ($5,000) + High-quality write-up (+50%) + PoC exploit (+25%) = **$8,750**

### Severity Assessment

We use the **CVSS v3.1** scoring system to determine severity:

- **Critical**: CVSS 9.0 - 10.0
- **High**: CVSS 7.0 - 8.9
- **Medium**: CVSS 4.0 - 6.9
- **Low**: CVSS 0.1 - 3.9

**Note**: Final severity rating is determined by the StreamSpace security team based on:
- **Impact**: Data exposure, privilege level gained, users affected
- **Exploitability**: Attack complexity, user interaction required
- **Scope**: Blast radius and cascading effects
- **Context**: Specific to StreamSpace architecture and deployment

### Payment Methods

We support the following payment methods:

- ✅ Bank transfer (ACH, wire transfer)
- ✅ PayPal
- ✅ Cryptocurrency (Bitcoin, Ethereum via Coinbase Commerce)
- ✅ Donation to charity of your choice (we match 100%)

**Payment Timeline**:
- Payment processed within **30 days** of vulnerability fix being deployed to production
- Tax forms required for payments >$600 (US researchers, IRS 1099-MISC)

---

## Rules of Engagement

### Testing Guidelines

To ensure safe and responsible testing, please follow these rules:

#### ✅ DO:

1. **Test on staging environment first** (https://staging.streamspace.io)
2. **Use the provided test accounts** (see Submission Guidelines)
3. **Report vulnerabilities immediately** upon discovery
4. **Provide clear reproduction steps** in your report
5. **Give us reasonable time to fix** before public disclosure (90 days minimum)
6. **Respect user privacy** - do not access or modify other users' data
7. **Use automation responsibly** - stay within rate limits
8. **Stop testing** if you encounter PII or sensitive data

#### ❌ DON'T:

1. **Don't test on production** without explicit permission
2. **Don't perform attacks** that degrade service (DoS, spam)
3. **Don't access, modify, or delete** other users' data
4. **Don't exfiltrate data** beyond the minimum needed for PoC
5. **Don't publicly disclose** vulnerabilities before fixes are deployed
6. **Don't use vulnerabilities** for personal gain or malicious purposes
7. **Don't perform social engineering** against employees or users
8. **Don't perform physical attacks** or access restricted areas

### Rate Limiting

To prevent accidental DoS, please adhere to these limits:

- **API requests**: Max 100 requests per minute per IP
- **Login attempts**: Max 10 attempts per hour per account
- **WebSocket connections**: Max 5 concurrent connections per user
- **Brute force testing**: Coordinate with security team for rate limit exemptions

If you need higher limits for testing, email security@streamspace.io with your testing plan.

### Test Accounts

We provide dedicated test accounts for security testing:

**Test Account #1 (Regular User)**:
- Username: `bugbounty-user1`
- Email: `bugbounty-user1@streamspace.io`
- Password: Request via security@streamspace.io
- Permissions: Standard user

**Test Account #2 (Admin User)**:
- Username: `bugbounty-admin1`
- Email: `bugbounty-admin1@streamspace.io`
- Password: Request via security@streamspace.io
- Permissions: Admin (use to test privilege escalation)

**API Test Key**:
- Request via security@streamspace.io
- Scoped to test data only

### Coordination

Before performing any of the following, **please coordinate with our security team**:

- Large-scale automated testing (>1000 requests)
- Testing that may impact availability
- Social engineering tests (with written permission only)
- Testing requiring access to internal networks
- Exploiting vulnerabilities that could affect other users

**Email**: security@streamspace.io with subject line "Bug Bounty - Testing Coordination Request"

---

## Submission Guidelines

### How to Submit

1. **Email** your report to: security@streamspace.io
2. **Subject line**: `[Bug Bounty] <Severity> - <Brief Description>`
   - Example: `[Bug Bounty] CRITICAL - SQL Injection in /api/v1/sessions`
3. **Encrypt** your report using our PGP key (https://streamspace.io/.well-known/pgp-key.txt)
4. **Include** all required information (see Report Template below)

### Report Template

Please use this template for all submissions:

```markdown
# Vulnerability Report

## Summary
[Brief one-sentence description of the vulnerability]

## Severity
[Your assessment: Critical / High / Medium / Low]

## Description
[Detailed explanation of the vulnerability, including:
- What component is affected
- What security control is bypassed
- What the attacker can achieve
]

## Steps to Reproduce
1. [Clear, numbered steps]
2. [Include all necessary details]
3. [Anyone should be able to reproduce]

## Proof of Concept
[Include:
- Screenshots or videos
- Command-line examples
- Code snippets
- Request/response samples
]

## Impact
[Explain the real-world impact:
- How many users are affected?
- What data can be accessed?
- What actions can an attacker perform?
]

## Suggested Fix
[Optional but appreciated:
- How would you recommend fixing this?
- Code patches welcome
]

## References
- [Link to CVE, CWE, OWASP article, etc.]
- [Supporting research or blog posts]

## Researcher Information
- Name: [Your name or handle]
- Email: [Contact email]
- Website/Twitter: [Optional]
- Payment method preference: [Bank/PayPal/Crypto/Charity]
```

### What Makes a Great Report?

**Excellent reports include**:

✅ **Clear title** that immediately conveys the issue
✅ **Detailed steps** that anyone can follow to reproduce
✅ **Screenshots/videos** showing the vulnerability in action
✅ **Impact analysis** explaining why this matters
✅ **PoC exploit** demonstrating the vulnerability (if applicable)
✅ **Suggested fix** or code patch
✅ **Professional tone** and clear writing

**Poor reports lack**:

❌ Vague descriptions like "Your site has XSS"
❌ Missing reproduction steps
❌ No proof of concept or evidence
❌ Unclear impact assessment
❌ Duplicate of known issues

### Example Excellent Report

```markdown
# [Bug Bounty] CRITICAL - Authentication Bypass via JWT Algorithm Confusion

## Summary
An attacker can bypass authentication by exploiting the JWT "none" algorithm
to forge arbitrary tokens and gain unauthorized access.

## Severity
Critical (CVSS 9.8)

## Description
The StreamSpace API accepts JWT tokens signed with the "none" algorithm,
allowing an attacker to forge tokens without a secret. By changing the
algorithm from "RS256" to "none" and removing the signature, an attacker
can authenticate as any user, including administrators.

## Steps to Reproduce
1. Obtain a valid JWT token by logging in as a regular user
2. Decode the JWT using https://jwt.io
3. Change the header from:
   ```json
   {"alg": "RS256", "typ": "JWT"}
   ```
   to:
   ```json
   {"alg": "none", "typ": "JWT"}
   ```
4. Modify the payload to escalate privileges:
   ```json
   {"sub": "admin", "role": "admin", "exp": 9999999999}
   ```
5. Remove the signature portion (everything after the second dot)
6. Base64-encode the header and payload
7. Send request with forged token:
   ```bash
   curl -H "Authorization: Bearer <forged_token>" \
     https://api.streamspace.io/api/v1/admin/users
   ```

## Proof of Concept
[Attached video: auth-bypass-poc.mp4 showing successful admin access]

Forged token used:
eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsImV4cCI6OTk5OTk5OTk5OX0.

## Impact
- **Complete authentication bypass**: Attacker can impersonate any user
- **Full admin access**: Can create/delete users, access all sessions
- **Data breach**: Can access all user data and sessions
- **Affects**: All users of StreamSpace (estimated 10,000+)

## Suggested Fix
Update `api/internal/middleware/auth.go:45` to explicitly validate the
JWT algorithm:

```go
token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
    // Ensure the algorithm is RS256
    if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return jwtPublicKey, nil
})
```

## References
- https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
- CWE-347: Improper Verification of Cryptographic Signature
- OWASP JWT Cheat Sheet

## Researcher Information
- Name: Jane Doe
- Email: jane.doe@security-research.com
- Twitter: @janedoe_sec
- Payment: PayPal preferred
```

---

## Disclosure Policy

We believe in **coordinated disclosure** that protects users while recognizing researchers.

### Our Commitment

When you submit a vulnerability:

1. **Acknowledgment**: We will acknowledge your report within **24 hours**
2. **Validation**: We will validate and triage your report within **72 hours**
3. **Updates**: We will provide updates every **5 business days** on fix progress
4. **Fix Timeline**: We aim to deploy critical fixes within **30 days**, high severity within **60 days**
5. **Bounty Payment**: We will process payment within **30 days** of fix deployment
6. **Public Credit**: We will credit you in our security advisories (if desired)

### Your Commitment

We ask that you:

1. **Give us time**: Wait **90 days minimum** before public disclosure
2. **Coordinate**: Work with us on disclosure timing if additional time is needed
3. **Respect privacy**: Don't publish sensitive user data or PII
4. **Be professional**: Communicate respectfully and constructively

### Disclosure Timeline

**Standard Timeline**:
- **Day 0**: Vulnerability reported
- **Day 1**: Acknowledgment from StreamSpace
- **Day 3**: Validation and severity assessment
- **Day 30**: Fix deployed for critical issues
- **Day 60**: Fix deployed for high severity
- **Day 90**: Coordinated public disclosure (if applicable)

**Expedited Disclosure** (for active exploitation):
- If vulnerability is being actively exploited in the wild
- Coordinate immediate disclosure with security team
- We may issue emergency patch within 24-48 hours

**Extended Timeline**:
- For complex vulnerabilities requiring architectural changes
- We may request extension beyond 90 days
- Researcher has final say on disclosure timing

### Public Disclosure

After the fix is deployed, we will:

1. **Publish security advisory** on GitHub Security Advisories
2. **Credit the researcher** (unless they request anonymity)
3. **Assign CVE** (if applicable)
4. **Update our security page** with lessons learned
5. **Add researcher to Hall of Fame**

Researchers are welcome to publish their own write-ups after coordinated disclosure.

---

## Safe Harbor

StreamSpace commits to the following **Safe Harbor** protections for good-faith security research:

### Legal Protection

We will **not initiate legal action** against you if you:

1. Make a good-faith effort to comply with this policy
2. Do not intentionally harm users or StreamSpace
3. Do not exfiltrate, modify, or delete user data beyond the minimum for PoC
4. Do not publicly disclose vulnerabilities before fix deployment
5. Respect the rules of engagement outlined above

### No Law Enforcement Referrals

We will **not refer good-faith researchers** to law enforcement for:

- Computer Fraud and Abuse Act (CFAA) violations (US)
- Similar computer misuse laws in other jurisdictions

### Authorization

This bug bounty program provides **explicit authorization** for security testing of in-scope assets, provided you follow the rules of engagement.

### Exception Handling

If you inadvertently:

- Access other users' data (stop immediately and report)
- Cause a disruption (notify us immediately)
- Violate a rule unintentionally (communicate with us)

We will work with you constructively. **Communication is key**.

### Your Protections

To maintain Safe Harbor protections:

✅ **Act in good faith** at all times
✅ **Follow the rules** of this policy
✅ **Report vulnerabilities** promptly
✅ **Respect user privacy** and data
✅ **Avoid service disruptions**
✅ **Communicate** with our security team

---

## Hall of Fame

We recognize and thank the following security researchers who have helped make StreamSpace more secure:

### 2025

| Researcher | Vulnerability | Severity | Bounty |
|------------|---------------|----------|--------|
| *Program launching soon* | - | - | - |

### Recognition Tiers

- 🥇 **Gold**: 3+ Critical vulnerabilities
- 🥈 **Silver**: 5+ High vulnerabilities
- 🥉 **Bronze**: 10+ Medium vulnerabilities
- ⭐ **Star**: First bounty recipient

### Anonymity Option

If you prefer to remain anonymous, we will list you as:
- "Anonymous Researcher #1"
- Or a pseudonym of your choice

Let us know your preference when submitting.

---

## FAQ

### Q: Can I test in production?

**A**: No, please use our staging environment (https://staging.streamspace.io) or a local deployment. Production testing requires explicit written permission.

### Q: What if I find a vulnerability in a dependency?

**A**: Report it directly to the upstream project. If it's a zero-day affecting StreamSpace, report it to us as well so we can coordinate with the vendor.

### Q: How long do I have to wait before disclosing?

**A**: We ask for a minimum of 90 days. For complex issues, we may request an extension, but you have the final say.

### Q: Can I report anonymously?

**A**: Yes! You can use a pseudonym or request full anonymity. Just let us know your payment preference.

### Q: What if my report is a duplicate?

**A**: Duplicates are not eligible for bounty, but we'll still acknowledge your effort. First reporter wins.

### Q: Do you accept reports from employees?

**A**: No, employees and contractors are not eligible for bounty rewards (but please report issues you find!).

### Q: Can I donate my bounty to charity?

**A**: Absolutely! We'll match your donation 100% and provide a tax receipt.

### Q: What if I disagree with the severity assessment?

**A**: We're happy to discuss severity ratings. Email security@streamspace.io with your reasoning and any additional evidence.

### Q: Do you offer swag?

**A**: Yes! All researchers who submit valid vulnerabilities receive StreamSpace swag (t-shirt, stickers, etc.). Email us your mailing address.

### Q: Can I blog about my findings?

**A**: Yes, after coordinated disclosure! We encourage write-ups and will link to them from our security advisories.

---

## Updates to This Policy

This policy may be updated periodically. Material changes will be announced via:

- Email to registered researchers
- GitHub repository announcement
- Security mailing list

**Last Updated**: 2025-11-14
**Version**: 1.0

---

## Contact

**Email**: security@streamspace.io
**PGP Key**: https://streamspace.io/.well-known/pgp-key.txt
**GitHub**: https://github.com/JoshuaAFerguson/streamspace/security
**Twitter**: @StreamSpaceIO

**Program Manager**: security-team@streamspace.io

---

## Legal

This bug bounty program is governed by the laws of [Your Jurisdiction]. By participating, you agree to:

1. Comply with all applicable laws and regulations
2. Follow this bug bounty policy in good faith
3. Work constructively with the StreamSpace security team

StreamSpace reserves the right to modify or cancel this program at any time. We reserve the right to determine bounty eligibility and amounts at our sole discretion.

**Thank you for helping us keep StreamSpace secure!** 🔒

---

**End of Bug Bounty Program Policy**
