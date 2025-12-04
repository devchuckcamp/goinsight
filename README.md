# GoInsight - Customer Feedback Analytics Copilot

GoInsight is an internal analytics tool that provides LLM-powered insights on customer feedback data. Product managers can ask natural language questions about customer feedback, and the service generates SQL queries, analyzes the data, and returns actionable insights with recommendations and suggested action items.

**✨ NEW: Jira Integration** - Automatically convert AI-generated insights into Jira tickets! See [JIRA_INTEGRATION.md](JIRA_INTEGRATION.md) for details.

## 🏗️ Architecture Overview

The service operates in three main steps:

1. **SQL Generation**: Converts natural language questions into safe SQL SELECT queries
2. **Data Retrieval**: Executes queries against a Postgres database containing structured feedback
3. **Insight Generation**: Uses an LLM to analyze results and produce summaries, recommendations, and action items

### Tech Stack
- **Language**: Go 1.22+
- **Database**: PostgreSQL 15
- **HTTP Router**: Chi
- **LLM Provider**: OpenAI (configurable interface)
- **Containerization**: Docker & Docker Compose

## 📁 Project Structure

```
goinsight/
├── cmd/
│   ├── api/              # Main API server entrypoint
│   │   └── main.go
│   └── seed/             # Database seeder utility
│       └── main.go
├── internal/
│   ├── config/           # Configuration & environment variable loading
│   │   └── config.go
│   ├── db/               # Database client and query execution
│   │   ├── postgres.go
│   │   └── migrations.go
│   ├── domain/           # Domain models and types
│   │   └── feedback.go
│   ├── http/             # HTTP handlers and routing
│   │   ├── handlers.go
│   │   └── router.go
│   └── llm/              # LLM client interface and implementations
│       ├── client.go     # Interface definition
│       ├── prompts.go    # System prompts for SQL and insight generation
│       ├── openai_client.go
│       └── mock_client.go
├── migrations/           # SQL migration files
│   ├── 001_init.sql
│   └── 002_seed_feedback.sql
├── Dockerfile
├── docker-compose.yml
├── .env.example          # Example environment variables (NEVER commit real .env!)
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- **Docker** and **Docker Compose** installed
- **Go 1.22+** (for local development)
- **LLM Provider** (choose one):
  - Mock client (built-in, no setup) - for testing
  - **Groq API key** (FREE, recommended) - get at https://console.groq.com
  - **Ollama** (FREE, local) - install from https://ollama.ai
  - OpenAI API key (paid) - get at https://platform.openai.com
- **Jira Cloud** (optional, for ticket creation) - see [JIRA_INTEGRATION.md](JIRA_INTEGRATION.md)

### Setup Instructions

#### 1. Clone the Repository

```bash
cd goinsight
```

#### 2. Create Your Environment File

**CRITICAL**: The `.env` file contains sensitive API keys and must NEVER be committed to version control.

```bash
cp .env.example .env
```

Edit `.env` and fill in your actual values:

```env
# Database Configuration
DATABASE_URL=postgres://goinsight:your_password@postgres:5432/goinsight?sslmode=disable
POSTGRES_PASSWORD=your_password

# LLM Configuration - Using FREE Groq! (RECOMMENDED)
# Get free key: https://console.groq.com/keys
LLM_PROVIDER=groq
GROQ_API_KEY=your_groq_api_key_here
LLM_MODEL=

# Other options:
# Option 1: Mock (No setup, testing only)
# LLM_PROVIDER=mock

# Option 2: Ollama (FREE, local)
# Install: https://ollama.ai
# OLLAMA_URL=http://host.docker.internal:11434
# LLM_PROVIDER=ollama
# LLM_MODEL=llama3

# Option 3: OpenAI (Paid)
# OPENAI_API_KEY=sk-your-key-here
# LLM_PROVIDER=openai
# LLM_MODEL=gpt-4o-mini

# Jira Configuration (Optional - for creating tickets from insights)
# Get your API token: https://id.atlassian.com/manage-profile/security/api-tokens
# JIRA_BASE_URL=https://your-domain.atlassian.net
# JIRA_EMAIL=your-email@company.com
# JIRA_API_TOKEN=your_jira_api_token_here

# Server Configuration
PORT=8080
ENV=development
DEBUG=false
```

**⚠️ IMPORTANT**: 
- The `.env` file is in `.gitignore` and will NOT be committed
- Never hardcode API keys in source code, Dockerfiles, or docker-compose.yml
- Only placeholder/development values appear in committed files

#### 3. Start the Services

```bash
docker compose up --build
```

This will:
- Build the Go application container
- Start PostgreSQL
- Run database migrations automatically
- Start the API server on port 8080

The API will wait for PostgreSQL to be ready before starting.

#### 4. Seed Mock Data

The seed data (33 sample feedback records) is automatically inserted via the `002_seed_feedback.sql` migration.

Alternatively, you can run the seeder manually:

```bash
go run cmd/seed/main.go
```

Or inside Docker:

```bash
docker compose exec api ./seed
```

## 📡 API Endpoints

### Health Check

```bash
curl http://localhost:8080/api/health
```

**Response:**
```json
{
  "status": "healthy"
}
```

### Create Jira Tickets (NEW!)

Convert AI-generated insights into Jira tickets. See the complete guide: [JIRA_INTEGRATION.md](JIRA_INTEGRATION.md)

```bash
curl -X POST http://localhost:8080/api/jira-tickets \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are critical issues?",
    "summary": "Analysis shows...",
    "recommendations": ["Fix X", "Improve Y"],
    "actions": [
      {"title": "Fix Critical Bug", "description": "..."}
    ],
    "meta": {
      "project_key": "PROD",
      "default_issue_type": "Story",
      "default_labels": ["feedback", "ai-insight"]
    }
  }'
```

### Ask a Question

```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the most common billing issues?"
  }'
```

**Response:**
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
    }
  ],
  "summary": "The data shows that refund processing is the most common billing issue with 2 reports, followed by invoice errors and payment method requests.",
  "recommendations": [
    "Prioritize fixing the refund processing workflow",
    "Review invoice generation logic for accuracy",
    "Consider adding PayPal as a payment option"
  ],
  "actions": [
    {
      "title": "Fix Refund Processing System",
      "description": "Investigate and resolve the refund processing delays affecting enterprise customers. This is critical as it's blocking quarterly reconciliation."
    },
    {
      "title": "Audit Invoice Generation",
      "description": "Review the invoice calculation logic to prevent incorrect amounts on subscription upgrades."
    }
  ]
}
```

### Example Questions to Try

```bash
# Performance issues
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "Show me all critical performance issues"}'

# Sentiment analysis
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "What is the sentiment distribution for onboarding feedback?"}'

# Regional analysis
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "Which region has the most negative feedback?"}'

# Customer tier analysis
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "What are enterprise customers complaining about?"}'
```

## 🗄️ Database Schema

The `feedback_enriched` table structure:

```sql
CREATE TABLE feedback_enriched (
    id            TEXT PRIMARY KEY,
    created_at    TIMESTAMPTZ NOT NULL,
    source        TEXT NOT NULL,        -- 'zendesk', 'google_play', 'nps_survey'
    product_area  TEXT NOT NULL,        -- 'billing', 'onboarding', 'performance', etc.
    sentiment     TEXT NOT NULL,        -- 'positive', 'neutral', 'negative'
    priority      INT NOT NULL,         -- 1 (low) to 5 (critical)
    topic         TEXT NOT NULL,        -- 'refund issues', 'slow load times', etc.
    region        TEXT NOT NULL,        -- 'NA', 'EU', 'APAC'
    customer_tier TEXT NOT NULL,        -- 'free', 'pro', 'enterprise'
    summary       TEXT NOT NULL         -- Short feedback summary
);
```

## 🔧 Configuration

All configuration is managed through environment variables:

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Yes | - |
| `LLM_PROVIDER` | LLM provider: `mock`, `groq`, `ollama`, `openai` | No | `mock` |
| `OPENAI_API_KEY` | OpenAI API key (if using OpenAI) | Conditional | - |
| `GROQ_API_KEY` | Groq API key (if using Groq) | Conditional | - |
| `OLLAMA_URL` | Ollama server URL (if using Ollama) | No | `http://localhost:11434` |
| `LLM_MODEL` | LLM model to use | No | Provider-specific default |
| `JIRA_BASE_URL` | Jira Cloud base URL (optional) | No | - |
| `JIRA_EMAIL` | Jira account email (optional) | No | - |
| `JIRA_API_TOKEN` | Jira API token (optional) | No | - |
| `PORT` | HTTP server port | No | `8080` |
| `ENV` | Environment name | No | `development` |
| `DEBUG` | Enable debug logging | No | `false` |

### Configuration Loading

The application uses the `godotenv` package to load `.env` files for local development. In production/Docker environments, environment variables can be set directly without a `.env` file.

Configuration validation happens at startup - the app will not start without required environment variables.

## 🔌 LLM Provider Integration

### Current Implementation

The OpenAI client (`internal/llm/openai_client.go`) reads the `OPENAI_API_KEY` from the environment and makes requests to the OpenAI API.

### Using Mock Client

If `OPENAI_API_KEY` is not set, the application automatically falls back to a mock client that returns placeholder responses. This is useful for:
- Testing the API structure without LLM costs
- Development without API access
- CI/CD pipelines

### Adding a New Provider

To add a new LLM provider (e.g., Claude, Llama):

1. Create a new file: `internal/llm/provider_client.go`
2. Implement the `Client` interface:
   ```go
   type Client interface {
       GenerateSQL(ctx context.Context, question string) (string, error)
       GenerateInsight(ctx context.Context, question string, queryResults []map[string]any) (string, error)
   }
   ```
3. Update `cmd/api/main.go` to initialize your client based on `LLM_PROVIDER` environment variable

## 🧪 Development

### Running Locally (Without Docker)

1. Start PostgreSQL:
   ```bash
   docker compose up postgres -d
   ```

2. Set up your `.env` file with `DATABASE_URL` pointing to `localhost:5432`

3. Run migrations:
   ```bash
   go run cmd/api/main.go
   # Migrations run automatically on startup
   ```

4. Seed data:
   ```bash
   go run cmd/seed/main.go
   ```

5. Run the API:
   ```bash
   go run cmd/api/main.go
   ```

### Running Tests

```bash
go test ./...
```

### Adding New Migrations

Create a new SQL file in `migrations/` with an incremental number:

```bash
# Example: migrations/003_add_sentiment_index.sql
CREATE INDEX idx_sentiment_priority ON feedback_enriched(sentiment, priority);
```

Migrations run automatically on application startup.

## 🔒 Security Considerations

### Environment Variables & Secrets

- **NEVER** commit `.env` files containing real API keys
- **NEVER** hardcode secrets in source code, Dockerfiles, or docker-compose.yml
- Use placeholder values in committed files
- In production, use secret management systems (AWS Secrets Manager, HashiCorp Vault, etc.)

### SQL Injection Protection

The application implements multiple layers of SQL injection protection:

1. **LLM Prompt Engineering**: The SQL generation prompt explicitly instructs the LLM to generate only SELECT queries
2. **Query Validation**: The handler validates that queries start with SELECT
3. **Keyword Blacklist**: Dangerous keywords (DROP, DELETE, INSERT, UPDATE, etc.) are blocked
4. **Read-Only Operations**: Only SELECT queries are allowed

For production use, consider:
- Using a read-only database user
- Implementing query timeouts
- Adding rate limiting
- Auditing all generated queries

## 🐳 Docker Commands

```bash
# Start services
docker compose up -d

# View logs
docker compose logs -f api

# Stop services
docker compose down

# Rebuild after code changes
docker compose up --build

# Execute commands in container
docker compose exec api sh

# Reset database
docker compose down -v  # Removes volumes
docker compose up --build
```

## 📝 Notes

- The seed data includes 33 diverse feedback records covering billing, onboarding, performance, features, security, UI/UX, and integrations
- Sample data spans different regions (NA, EU, APAC), customer tiers (free, pro, enterprise), and sentiment levels
- Migrations are idempotent and safe to run multiple times
- The application includes automatic database connection retry logic for Docker environments

## 🤝 Contributing

When contributing:

1. Never commit real API keys or credentials
2. Test with the mock client before using real LLM APIs
3. Ensure all new environment variables are documented in `.env.example`
4. Add appropriate indexes for new query patterns
5. Follow Go best practices and idiomatic code style

## 📄 License

Internal use only - not for public distribution.

---

**Questions or Issues?** Check the logs with `docker compose logs -f api` or open an issue in the repository.
