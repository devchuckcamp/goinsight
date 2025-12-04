# API Examples & Expected Responses

## Example 1: Billing Issues

### Request
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the most common billing issues?"
  }'
```

### Expected Response Structure
```json
{
  "question": "What are the most common billing issues?",
  "data_preview": [
    {
      "topic": "refund processing",
      "count": 2
    },
    {
      "topic": "invoice errors",
      "count": 1
    },
    {
      "topic": "payment methods",
      "count": 1
    },
    {
      "topic": "subscription cancellation",
      "count": 1
    }
  ],
  "summary": "Analysis of billing feedback reveals that refund processing is the most critical issue with 2 reports from enterprise customers. Invoice errors and subscription cancellation problems are also significant concerns affecting pro-tier customers across multiple regions.",
  "recommendations": [
    "Prioritize refund processing workflow improvements - currently blocking enterprise customer operations",
    "Implement automated invoice validation to catch calculation errors before delivery",
    "Add self-service subscription cancellation to reduce support burden",
    "Expand payment method options to include PayPal based on user requests"
  ],
  "actions": [
    {
      "title": "Fix Critical Refund Processing Delays",
      "description": "Investigate and resolve refund processing system that is causing 30+ day delays for enterprise customers. This is blocking quarterly reconciliation and has escalation risk. Priority: Critical."
    },
    {
      "title": "Audit Invoice Calculation Logic",
      "description": "Review and fix invoice generation system that produces incorrect amounts during subscription upgrades. Implement automated validation checks."
    },
    {
      "title": "Implement Self-Service Cancellation",
      "description": "Add in-app subscription cancellation flow to reduce support tickets and improve user experience, particularly for mobile users."
    }
  ]
}
```

## Example 2: Performance Issues

### Request
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Show me critical performance issues from the last week"
  }'
```

### Expected Response Structure
```json
{
  "question": "Show me critical performance issues from the last week",
  "data_preview": [
    {
      "id": "fb-011",
      "created_at": "2025-12-02T...",
      "product_area": "performance",
      "sentiment": "negative",
      "priority": 5,
      "topic": "app crashes",
      "customer_tier": "enterprise",
      "summary": "App crashes when loading large datasets, making it unusable"
    },
    {
      "id": "fb-015",
      "created_at": "2025-11-28T...",
      "product_area": "performance",
      "sentiment": "negative",
      "priority": 5,
      "topic": "app crashes",
      "customer_tier": "pro",
      "summary": "App crashes immediately on launch after latest update"
    }
  ],
  "summary": "There are 2 critical performance issues (priority 5) in the last week, both involving app crashes. These affect both enterprise and pro-tier customers across APAC and NA regions. One relates to large dataset handling, the other to a recent update causing launch failures.",
  "recommendations": [
    "Immediately rollback or hotfix the latest update causing launch crashes for APAC pro customers",
    "Implement lazy loading for large datasets to prevent enterprise customer crashes",
    "Add crash reporting and monitoring to catch these issues before customer reports",
    "Establish performance testing as a mandatory release gate for all updates"
  ],
  "actions": [
    {
      "title": "Emergency Hotfix: App Launch Crash",
      "description": "Critical production issue - app crashes on launch after latest update for APAC pro-tier customers. Investigate update changes and deploy fix or rollback immediately."
    },
    {
      "title": "Optimize Large Dataset Loading",
      "description": "Implement pagination and lazy loading for enterprise customers working with large datasets. Current implementation causes memory overflow and crashes."
    }
  ]
}
```

## Example 3: Enterprise Customer Feedback

### Request
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are enterprise customers complaining about most?"
  }'
```

### Expected Response Structure
```json
{
  "question": "What are enterprise customers complaining about most?",
  "data_preview": [
    {
      "product_area": "security",
      "count": 3,
      "avg_priority": 4.67
    },
    {
      "product_area": "performance",
      "count": 2,
      "avg_priority": 4.5
    },
    {
      "product_area": "billing",
      "count": 2,
      "avg_priority": 5.0
    }
  ],
  "summary": "Enterprise customers are primarily concerned with security (3 reports, avg priority 4.67) and billing issues (2 reports, priority 5.0). Security concerns focus on compliance requirements (SOC2) and access controls, while billing issues involve critical refund processing delays.",
  "recommendations": [
    "Fast-track SOC2 compliance documentation - multiple customers requiring this for audits",
    "Implement SSO and role-based access controls as high-priority security features",
    "Resolve refund processing system immediately - affecting enterprise operations",
    "Create enterprise-specific SLA for critical issues (< 24hr response time)",
    "Establish quarterly business reviews with enterprise customers to proactively address concerns"
  ],
  "actions": [
    {
      "title": "Complete SOC2 Compliance Audit",
      "description": "Multiple enterprise customers requesting SOC2 documentation for their audits. Engage compliance team to complete certification process and publish documentation."
    },
    {
      "title": "Implement SSO Integration",
      "description": "Enterprise requirement for Single Sign-On. Integrate with common providers (Okta, Azure AD, Google Workspace). Include role-based access control."
    },
    {
      "title": "Fix Enterprise Refund Processing",
      "description": "Critical billing system issue causing 30+ day refund delays. This is blocking quarterly reconciliation for enterprise customers and has escalation risk."
    }
  ]
}
```

## Example 4: Sentiment Analysis

### Request
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is the sentiment distribution across product areas?"
  }'
```

### Expected Response Structure
```json
{
  "question": "What is the sentiment distribution across product areas?",
  "data_preview": [
    {
      "product_area": "billing",
      "sentiment": "negative",
      "count": 5
    },
    {
      "product_area": "performance",
      "sentiment": "negative",
      "count": 5
    },
    {
      "product_area": "onboarding",
      "sentiment": "positive",
      "count": 2
    },
    {
      "product_area": "onboarding",
      "sentiment": "negative",
      "count": 2
    }
  ],
  "summary": "Sentiment analysis shows billing and performance as the most problematic areas with 100% negative sentiment (5 reports each). Onboarding has mixed sentiment (50/50), while features and UI/UX are predominantly positive. Security and integrations are mostly negative, indicating systemic issues.",
  "recommendations": [
    "Declare billing and performance as crisis areas requiring immediate product intervention",
    "Maintain and amplify onboarding investments - it's working well with positive feedback",
    "Address security concerns proactively before they become customer blockers",
    "Celebrate and document what's working well in features and UI/UX to replicate success"
  ],
  "actions": [
    {
      "title": "Launch Billing Task Force",
      "description": "Create dedicated task force to address 100% negative sentiment in billing. Focus on refund processing, invoice accuracy, and subscription management. Weekly executive updates."
    },
    {
      "title": "Performance Optimization Sprint",
      "description": "Dedicate next sprint to performance issues. Focus on app stability, load times, and memory management. Set measurable improvement targets."
    }
  ]
}
```

## Example 5: Regional Analysis

### Request
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Which region has the most feedback and what are the top issues?"
  }'
```

### Expected Response Structure
```json
{
  "question": "Which region has the most feedback and what are the top issues?",
  "data_preview": [
    {
      "region": "NA",
      "count": 13,
      "negative_count": 8,
      "avg_priority": 3.46
    },
    {
      "region": "EU",
      "count": 11,
      "negative_count": 6,
      "avg_priority": 3.27
    },
    {
      "region": "APAC",
      "count": 9,
      "negative_count": 5,
      "avg_priority": 3.44
    }
  ],
  "summary": "North America (NA) generates the most feedback with 13 reports, of which 62% are negative. EU and APAC show similar patterns with 11 and 9 reports respectively. Priority levels are consistent across regions (avg ~3.4), suggesting similar severity of issues globally.",
  "recommendations": [
    "Investigate why NA has disproportionately high feedback volume - could indicate larger user base or specific regional issues",
    "Ensure support coverage aligns with feedback volumes (NA > EU > APAC)",
    "Look for region-specific issues (e.g., payment methods, compliance requirements)",
    "Consider regional infrastructure improvements if performance issues cluster in APAC"
  ],
  "actions": [
    {
      "title": "NA Region Deep Dive Analysis",
      "description": "Conduct detailed analysis of why NA region has highest feedback volume and negative sentiment. Interview customer success team and review support tickets."
    },
    {
      "title": "Regional Support Optimization",
      "description": "Adjust support staffing and hours to match feedback volumes by region. Ensure NA has adequate coverage given higher ticket volume."
    }
  ]
}
```

## Notes

### Data Preview
- Limited to first 10 rows or aggregated results
- Can include counts, averages, or full records depending on the query
- LLM-generated SQL determines the structure

### Summary
- 2-3 sentence overview of findings
- Data-driven and specific
- Highlights key patterns or anomalies

### Recommendations
- 3-5 actionable suggestions
- Based on data patterns
- Prioritized by impact

### Actions
- 2-4 specific tasks that could become tickets
- Include title and detailed description
- Often include priority or urgency indicators
- Ready to copy into Jira or similar tools

### Variations
The exact structure of `data_preview` will vary based on:
- The natural language question
- The SQL query generated
- Whether the query aggregates or returns raw rows
- The LLM's interpretation of the optimal way to present the data

All responses follow the consistent schema defined in `internal/domain/feedback.go`.

## Example 6: Creating Jira Tickets from Insights

### Step 1: Get Insights (same as examples above)
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the most critical billing issues?"
  }'
```

### Step 2: Create Jira Tickets with Auto-Calculated Priority

```bash
curl -X POST http://localhost:8080/api/jira-tickets \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the most critical billing issues?",
    "summary": "Analysis shows refund processing delays affecting enterprise customers...",
    "recommendations": [
      "Improve the refund processing system to reduce delays",
      "Provide clearer communication about refund timelines"
    ],
    "actions": [
      {
        "title": "Investigate Refund Processing Delays",
        "description": "Analyze customer feedback to identify root causes...",
        "magnitude": 7.5
      },
      {
        "title": "Update Refund Process Documentation",
        "description": "Review and update documentation for clarity...",
        "magnitude": 4.5
      },
      {
        "title": "Critical Payment Gateway Outage",
        "description": "Emergency fix for payment gateway that is down",
        "magnitude": 9.5
      }
    ],
    "meta": {
      "project_key": "SASS",
      "default_issue_type": "Story",
      "default_labels": ["feedback", "ai-insight", "billing"]
    }
  }'
```

### Expected Response Structure
```json
{
  "ticket_specs": [
    {
      "project_key": "SASS",
      "issue_type": "Story",
      "summary": "Critical Payment Gateway Outage",
      "description": "Context: Analysis shows refund processing delays...\n\n### Impact\nEmergency fix for payment gateway that is down\n\n### Acceptance Criteria\n- Investigate gateway status\n- Implement emergency fix\n- Test payment processing\n- Monitor for stability",
      "priority": "Highest",
      "labels": ["feedback", "ai-insight", "billing", "payment", "critical"],
      "components": ["billing-system"]
    },
    {
      "project_key": "SASS",
      "issue_type": "Story",
      "summary": "Investigate Refund Processing Delays",
      "description": "Context: Analysis shows refund processing delays...\n\n### Impact\nDelays affecting enterprise customers\n\n### Acceptance Criteria\n- Analyze customer feedback\n- Identify root causes\n- Document findings",
      "priority": "High",
      "labels": ["feedback", "ai-insight", "billing", "refunds"],
      "components": ["billing-system"]
    },
    {
      "project_key": "SASS",
      "issue_type": "Story",
      "summary": "Update Refund Process Documentation",
      "description": "Context: Analysis shows refund processing delays...\n\n### Impact\nImprove clarity for support team and customers\n\n### Acceptance Criteria\n- Review current documentation\n- Update refund process details\n- Publish updated docs",
      "priority": "Medium",
      "labels": ["feedback", "ai-insight", "billing", "documentation"],
      "components": ["documentation"]
    }
  ],
  "created_tickets": [
    {
      "id": "10234",
      "key": "SASS-123",
      "self": "https://your-domain.atlassian.net/rest/api/2/issue/10234"
    },
    {
      "id": "10235",
      "key": "SASS-124",
      "self": "https://your-domain.atlassian.net/rest/api/2/issue/10235"
    },
    {
      "id": "10236",
      "key": "SASS-125",
      "self": "https://your-domain.atlassian.net/rest/api/2/issue/10236"
    }
  ],
  "errors": []
}
```

### How Magnitude Affects Priority

The system calculates a magnitude score (0-10) for each action based on keywords:

| Magnitude Score | Priority | Example Keywords |
|----------------|----------|------------------|
| ≥ 8.0 | **Highest** | critical, emergency, security, outage, crash, breach |
| ≥ 6.5 | **High** | investigate, billing, revenue, customer, payment |
| ≥ 4.0 | **Medium** | improve, optimize, implement, analyze |
| < 4.0 | **Low** | documentation, update, review, comment |

**Magnitude calculation factors:**
- Urgent keywords (+2.5): critical, emergency, security, outage, crash
- Impact keywords (+1.5): revenue, customer, payment, billing, compliance
- Investigation tasks (+1.0): investigate, analyze, research
- Documentation tasks (-1.0): documentation, update, review
- Recommendation mentions (+0.5 each)

For more details, see [JIRA_INTEGRATION.md](JIRA_INTEGRATION.md).
