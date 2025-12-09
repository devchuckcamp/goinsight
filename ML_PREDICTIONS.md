# ML Predictions Integration Guide

GoInsight now integrates with **tens-insight** (TensorFlow-based ML trainer) to surface predictive account health and product-area priority signals.

## Overview

The ML integration provides two key capabilities:
1. **Account Churn Risk**: Predict which accounts are at risk of churning
2. **Product Area Prioritization**: Identify which product areas need attention by customer segment

## Architecture

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│  tens-insight   │────────▶│   PostgreSQL     │◀────────│   goinsight     │
│  (ML Trainer)   │  writes │  ML Predictions  │  reads  │  (API Server)   │
└─────────────────┘         └──────────────────┘         └─────────────────┘
     TensorFlow                                              LLM + REST API
```

- **tens-insight**: Python/TensorFlow app that trains models and writes predictions to Postgres
- **goinsight**: Go API that reads predictions and combines them with LLM-generated insights

## Database Schema

### account_risk_scores

Stores ML predictions for account churn risk and health scores (populated by tens-insight):

```sql
CREATE TABLE account_risk_scores (
    account_id         VARCHAR PRIMARY KEY,
    churn_probability  FLOAT (0-1),        -- Probability of churn
    health_score       FLOAT (0-100),      -- Health score (inverse of churn)
    risk_category      VARCHAR,            -- 'low', 'medium', 'high', 'critical'
    predicted_at       TIMESTAMPTZ,
    model_version      VARCHAR
);
```

### product_area_impact

Stores ML predictions for product area priority scores by segment (populated by tens-insight):

```sql
CREATE TABLE product_area_impact (
    product_area        VARCHAR,
    segment             VARCHAR,           -- 'enterprise', 'smb', 'pro'
    priority_score      FLOAT (0-100),     -- Higher = more important
    feedback_count      INTEGER,
    avg_sentiment_score FLOAT (-1 to 1),   -- Negative to positive
    negative_count      INTEGER,
    critical_count      INTEGER,
    predicted_at        TIMESTAMPTZ,
    model_version       VARCHAR,
    PRIMARY KEY (product_area, segment)
);
```

## API Endpoints

### GET /api/accounts/{id}/health

Returns ML-based health and risk metrics for a specific account.

**Request:**
```bash
curl http://localhost:8080/api/accounts/acc_ent_001/health
```

**Response:**
```json
{
  "account_id": "acc_ent_001",
  "churn_probability": 0.78,
  "health_score": 32.0,
  "risk_category": "high",
  "recent_negative_feedback_count": 5,
  "predicted_at": "2025-12-08T10:30:00Z",
  "model_version": "v1.2.3"
}
```

**Use Cases:**
- Identify at-risk customers for proactive outreach
- Prioritize customer success efforts
- Track account health trends over time

### GET /api/priorities/product-areas

Returns ML-based priority scores for product areas, optionally filtered by segment.

**Request:**
```bash
# All segments
curl http://localhost:8080/api/priorities/product-areas

# Filter by segment
curl http://localhost:8080/api/priorities/product-areas?segment=enterprise
```

**Response:**
```json
{
  "product_areas": [
    {
      "product_area": "billing",
      "segment": "enterprise",
      "priority_score": 92.5,
      "feedback_count": 145,
      "avg_sentiment_score": -0.65,
      "negative_count": 98,
      "critical_count": 34,
      "predicted_at": "2025-12-08T10:30:00Z",
      "model_version": "v1.2.3"
    },
    {
      "product_area": "performance",
      "segment": "enterprise",
      "priority_score": 88.9,
      "feedback_count": 178,
      "avg_sentiment_score": -0.52,
      "negative_count": 89,
      "critical_count": 28,
      "predicted_at": "2025-12-08T10:30:00Z",
      "model_version": "v1.2.3"
    }
  ]
}
```

**Query Parameters:**
- `segment` (optional): Filter by customer segment (e.g., `enterprise`, `smb`, `pro`)

**Use Cases:**
- Prioritize product roadmap by impact
- Understand segment-specific pain points
- Data-driven sprint planning

## LLM-Enhanced Queries

The `/api/ask` endpoint now understands the new ML prediction tables. You can ask natural language questions that combine feedback data with ML predictions.

### Example Questions

**Churn Risk + Feedback Analysis:**
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Which enterprise accounts are at highest churn risk and what themes show up in their feedback?"
  }'
```

The LLM will generate SQL joining `account_risk_scores` with `feedback_enriched` to provide insights like:
- Top at-risk accounts
- Common feedback themes from those accounts
- Recommended actions to reduce churn risk

**Product Prioritization:**
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What top 3 product areas should we prioritize for SMB accounts to improve NPS?"
  }'
```

The LLM will query `product_area_impact` filtered by `segment='smb'` and provide:
- Top 3 areas ranked by priority score
- Volume and sentiment data
- Specific recommendations

**Cross-Analysis:**
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Summarize the main drivers of high churn risk across all segments"
  }'
```

The LLM will analyze both ML predictions and feedback data to identify:
- Product areas most correlated with churn
- Segment-specific patterns
- Actionable recommendations

## Integration Workflow

### 1. tens-insight Updates Predictions

The ML trainer runs periodically (e.g., daily) to:
1. Load feedback data from `feedback_enriched`
2. Train/update TensorFlow models
3. Generate predictions
4. Write to `account_risk_scores` and `product_area_impact`

### 2. goinsight Reads Predictions

The API server:
1. Reads predictions on-demand via structured endpoints
2. Combines predictions with feedback data for LLM queries
3. Returns actionable insights to product managers

### 3. Product Managers Take Action

PMs can:
1. Query structured data endpoints for dashboards
2. Ask natural language questions combining ML + feedback
3. Create Jira tickets from insights
4. Track trends over time

## Development

### Running Migrations

The ML prediction tables are created automatically when you start the service:

```bash
docker compose up --build
```

Migrations `003_add_account_risk_scores.sql` and `004_add_product_area_impact.sql` will run on startup.

### Testing Locally

Since tens-insight populates the tables, you'll need either:

**Option 1: Mock Data (for development)**
- The migrations include sample INSERT statements
- Restart the service to reset to sample data

**Option 2: Run tens-insight**
- Set up the TensorFlow trainer locally
- Configure it to write to the same Postgres instance
- See tens-insight documentation for setup

### Monitoring

Check that predictions are up-to-date:

```sql
-- Check latest prediction timestamps
SELECT MAX(predicted_at) as latest_account_prediction
FROM account_risk_scores;

SELECT MAX(predicted_at) as latest_priority_prediction
FROM product_area_impact;
```

If predictions are stale, check the tens-insight service status.

## Example Use Cases

### 1. Proactive Customer Success

**Goal**: Identify at-risk enterprise accounts and understand their pain points

```bash
# Get all high-risk enterprise accounts
curl http://localhost:8080/api/priorities/product-areas?segment=enterprise

# Deep dive on specific account
curl http://localhost:8080/api/accounts/acc_ent_001/health

# Ask LLM for insights
curl -X POST http://localhost:8080/api/ask \
  -d '{"question": "Show me feedback from high-risk enterprise accounts about billing"}'
```

### 2. Sprint Planning

**Goal**: Prioritize product areas for next sprint

```bash
# Get priorities for target segment
curl http://localhost:8080/api/priorities/product-areas?segment=smb

# Ask LLM for recommendations
curl -X POST http://localhost:8080/api/ask \
  -d '{"question": "What are the top 3 product areas we should focus on for SMB customers?"}'

# Create Jira tickets from insights
curl -X POST http://localhost:8080/api/jira-tickets \
  -d '{ /* insights payload */ }'
```

### 3. Executive Reporting

**Goal**: Understand overall health and priorities

```bash
# Ask high-level questions
curl -X POST http://localhost:8080/api/ask \
  -d '{"question": "What percentage of our enterprise accounts are at high churn risk?"}'

curl -X POST http://localhost:8080/api/ask \
  -d '{"question": "Which product area has the highest priority score across all segments?"}'
```

## Troubleshooting

### No data in ML tables

**Problem**: Queries return empty results

**Solutions**:
- Check if migrations ran: `docker compose logs api | grep migration`
- Verify tens-insight is running and writing predictions
- Check database connectivity between tens-insight and Postgres

### Stale predictions

**Problem**: `predicted_at` timestamps are old

**Solutions**:
- Check tens-insight scheduler/cron configuration
- Verify tens-insight has access to feedback data
- Review tens-insight logs for errors

### LLM not using ML tables

**Problem**: Natural language queries don't leverage ML data

**Solutions**:
- Be explicit in questions: "using churn predictions" or "based on priority scores"
- Check LLM prompt includes ML table documentation (in `internal/llm/prompts.go`)
- Try more specific questions that clearly need ML data

## Best Practices

1. **Keep predictions fresh**: Run tens-insight at least daily
2. **Monitor data quality**: Set up alerts for stale predictions
3. **Combine data sources**: Use ML predictions + feedback + Jira for full picture
4. **Validate insights**: Cross-reference LLM insights with structured endpoint data
5. **Track trends**: Query historical predictions to see changes over time

## Related Documentation

- [JIRA_INTEGRATION.md](JIRA_INTEGRATION.md) - Creating tickets from insights
- [FREE_LLM_GUIDE.md](FREE_LLM_GUIDE.md) - LLM provider setup
- [EXAMPLES.md](EXAMPLES.md) - More API examples
- tens-insight README - ML trainer setup (separate repo)

---

**Questions?** Open an issue or check the logs: `docker compose logs -f api`
