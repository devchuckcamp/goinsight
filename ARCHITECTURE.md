# GoInsight Architecture

## High-Level Architecture

```
┌─────────────────┐
│  Product Manager │
│    (Client)      │
└────────┬─────────┘
         │
         │ HTTP POST /api/ask
         │ {"question": "..."}
         ▼
┌─────────────────────────────────────────┐
│           Go API Server                  │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  1. HTTP Handler                   │ │
│  │     - Validates request            │ │
│  │     - Orchestrates workflow        │ │
│  └─────────┬──────────────────────────┘ │
│            │                             │
│            ▼                             │
│  ┌────────────────────────────────────┐ │
│  │  2. LLM Client (OpenAI)            │ │
│  │     - Generate SQL from question   │ │
│  │     - Uses OPENAI_API_KEY env var  │ │
│  └─────────┬──────────────────────────┘ │
│            │                             │
│            │ SQL Query                   │
│            ▼                             │
│  ┌────────────────────────────────────┐ │
│  │  3. SQL Validator                  │ │
│  │     - Ensure SELECT only           │ │
│  │     - Block dangerous keywords     │ │
│  └─────────┬──────────────────────────┘ │
│            │                             │
│            ▼                             │
│  ┌────────────────────────────────────┐ │
│  │  4. Database Client                │ │
│  │     - Execute query                │ │
│  │     - Return rows as maps          │ │
│  └─────────┬──────────────────────────┘ │
│            │                             │
│            │ Query Results               │
│            ▼                             │
│  ┌────────────────────────────────────┐ │
│  │  5. LLM Client (OpenAI)            │ │
│  │     - Analyze results              │ │
│  │     - Generate insights            │ │
│  │     - Create recommendations       │ │
│  │     - Suggest action items         │ │
│  └─────────┬──────────────────────────┘ │
│            │                             │
│            ▼                             │
│  ┌────────────────────────────────────┐ │
│  │  6. Response Builder               │ │
│  │     - Format JSON response         │ │
│  │     - Include data preview         │ │
│  └─────────┬──────────────────────────┘ │
└────────────┼─────────────────────────────┘
             │
             │ JSON Response
             ▼
┌─────────────────────────────────────────┐
│  Response:                               │
│  {                                       │
│    "question": "...",                    │
│    "data_preview": [...],                │
│    "summary": "...",                     │
│    "recommendations": [...],             │
│    "actions": [...]                      │
│  }                                       │
└─────────────────────────────────────────┘

         │
         ▼
┌─────────────────┐
│   PostgreSQL     │
│                  │
│  feedback_       │
│  enriched table  │
│                  │
│  (ETL Populated) │
└─────────────────┘
```

## Data Flow

### Step 1: SQL Generation
- Input: Natural language question
- Process: LLM converts question to SQL SELECT
- Output: SQL query string
- Security: Prompt engineering to ensure safe queries

### Step 2: Query Validation
- Input: Generated SQL query
- Process: 
  - Verify starts with SELECT
  - Check for forbidden keywords (DROP, DELETE, etc.)
- Output: Validated query or error

### Step 3: Query Execution
- Input: Validated SQL query
- Process: Execute against Postgres `feedback_enriched` table
- Output: Array of result rows (as maps)

### Step 4: Insight Generation
- Input: 
  - Original question
  - Query results
- Process: LLM analyzes data and generates:
  - Summary (2-3 sentences)
  - Recommendations (3-5 items)
  - Action items (2-4 tickets)
- Output: Structured JSON insight

### Step 5: Response Assembly
- Input: All components
- Process: Build final response
- Output: Complete JSON with preview, summary, recommendations, and actions

## Component Responsibilities

### Configuration Layer (`internal/config`)
- Load environment variables from `.env` or system
- Validate required configuration
- Provide type-safe access to config values
- **Critical**: Never hardcodes API keys - always from env

### Domain Layer (`internal/domain`)
- Define core data structures
- Request/response types
- Domain models matching database schema

### Database Layer (`internal/db`)
- Connection management with retry logic
- Query execution
- Result serialization to maps
- Migration runner
- Health checks

### LLM Layer (`internal/llm`)
- `Client` interface for provider abstraction
- OpenAI implementation (reads `OPENAI_API_KEY` from env)
- Mock implementation for testing
- System prompts for SQL and insight generation

### HTTP Layer (`internal/http`)
- Request routing (Chi router)
- Request validation
- Workflow orchestration
- Response formatting
- Error handling

### Application Layer (`cmd/api`)
- Application bootstrap
- Dependency injection
- Graceful shutdown
- Migration execution

## Security Model

### Environment Variable Security
1. **Development**: `.env` file (gitignored)
2. **Docker**: `env_file` directive + environment substitution
3. **Production**: External secret management (AWS Secrets Manager, Vault, etc.)
4. **Code**: Zero hardcoded secrets - all from environment

### SQL Injection Prevention
1. **Prompt Engineering**: Instruct LLM to generate only SELECT
2. **Query Validation**: Programmatic checks for SELECT prefix
3. **Keyword Blacklist**: Block dangerous SQL keywords
4. **Future Enhancement**: Use read-only DB user in production

### API Key Management
- `OPENAI_API_KEY` loaded from environment only
- Validation at startup (fails fast if missing)
- Never logged or exposed in responses
- Placeholder values in all committed files

## Configuration Management

### Environment Variables
- Loaded via `godotenv` for local development
- Direct env vars in Docker/production
- Validation ensures required vars present
- Type conversion with defaults

### Configuration Files
- `.env.example`: Template with placeholders
- `.env`: Local secrets (gitignored)
- `docker-compose.yml`: References env vars, no secrets
- `Dockerfile`: No secrets

## Database Schema

### Table: `feedback_enriched`
Populated by external ETL pipeline with structured customer feedback.

**Columns:**
- `id`: Unique identifier
- `created_at`: Timestamp
- `source`: Origin (zendesk, google_play, nps_survey)
- `product_area`: Feature area (billing, onboarding, etc.)
- `sentiment`: positive/neutral/negative
- `priority`: 1-5 (low to critical)
- `topic`: Specific issue tag
- `region`: Geographic region
- `customer_tier`: free/pro/enterprise
- `summary`: Text summary

**Indexes:**
- product_area, sentiment, priority (filtering)
- created_at DESC (recency)
- customer_tier, region (segmentation)

## Deployment Considerations

### Local Development
- Use `.env` file
- Docker Compose for dependencies
- `go run` for fast iteration

### Docker Deployment
- Multi-stage build (smaller images)
- Health checks for dependencies
- Automatic migrations on startup
- Graceful shutdown handling

### Production Recommendations
1. Use managed Postgres (RDS, Cloud SQL)
2. Implement connection pooling
3. Add request timeouts
4. Use read-only DB credentials
5. Implement rate limiting
6. Add structured logging
7. Set up monitoring/alerting
8. Use secret management service
9. Enable audit logging for queries
10. Implement query cost limits

## Extension Points

### Adding New LLM Providers
1. Implement `llm.Client` interface
2. Add provider-specific env vars
3. Update main.go initialization logic

### Adding Authentication
1. Add middleware in router.go
2. Implement JWT or API key validation
3. Add user context to requests

### Adding Analytics
1. Log generated queries
2. Track query patterns
3. Monitor LLM token usage
4. Measure response times

### Adding Caching
1. Cache frequent queries
2. Cache LLM responses (with TTL)
3. Use Redis for distributed cache
