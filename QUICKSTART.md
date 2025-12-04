# GoInsight - Quick Start Guide

## 🚀 5-Minute Setup

### 1. Prerequisites Check
```bash
docker --version
docker compose version
```

### 2. Create Your `.env` File
```bash
cp .env.example .env
```

**Edit `.env` and set required values:**
```env
# Database - use any secure password (just for local dev)
DATABASE_URL=postgres://goinsight:your_password@postgres:5432/goinsight?sslmode=disable
POSTGRES_USER=goinsight
POSTGRES_PASSWORD=your_password
POSTGRES_DB=goinsight

# LLM - Choose a FREE option! (see FREE_LLM_GUIDE.md)
LLM_PROVIDER=mock              # Start with mock (no setup)
LLM_MODEL=                     # Optional: auto-defaults based on provider
# OR use Groq (free): GROQ_API_KEY=gsk_key & LLM_PROVIDER=groq
# OR use Ollama (local): OLLAMA_URL=http://host.docker.internal:11434 & LLM_PROVIDER=ollama
```

**Important:** 
- The password in `DATABASE_URL` must match `POSTGRES_PASSWORD`
- See `FREE_LLM_GUIDE.md` for complete LLM setup options (all FREE!)

### 3. Start Everything
```bash
docker compose up --build
```

Wait for: `Server listening on port 8080`

### 4. Test the API

**Health Check:**
```bash
curl http://localhost:8080/api/health
```

**Ask a Question:**
```bash
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "What are the most critical issues from enterprise customers?"}'
```

## 📋 Sample Questions

```bash
# Performance issues
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "Show me all critical performance issues"}'

# Billing analysis
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "What are the most common billing complaints?"}'

# Regional insights
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "Which region has the most negative feedback?"}'

# Onboarding sentiment
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "What is the sentiment distribution for onboarding?"}'

# Enterprise customers
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "What are enterprise customers complaining about?"}'

# Recent feedback
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "Show me the most recent critical feedback"}'
```

## 🔧 Common Commands

```bash
# View logs
docker compose logs -f api

# Stop services
docker compose down

# Restart after code changes
docker compose up --build

# Reset database
docker compose down -v
docker compose up --build

# Run seeder manually
docker compose exec api go run cmd/seed/main.go
```

## 🗂️ Project Structure

```
goinsight/
├── cmd/
│   ├── api/main.go          # API server entrypoint
│   └── seed/main.go         # Database seeder
├── internal/
│   ├── config/              # Environment variable loading
│   ├── db/                  # Database client & migrations
│   ├── domain/              # Data models
│   ├── http/                # HTTP handlers & routing
│   └── llm/                 # LLM client implementations
├── migrations/
│   ├── 001_init.sql         # Create tables
│   └── 002_seed_feedback.sql # Sample data
├── .env.example             # Template (commit this)
├── .env                     # Your secrets (never commit)
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## 🔐 Security Reminders

- ✅ `.env` is in `.gitignore`
- ✅ Never commit API keys
- ✅ Only placeholder values in code
- ✅ All secrets from environment variables

## 🐛 Troubleshooting

### Database Connection Failed
```bash
# Wait a bit longer for Postgres to start, or check logs:
docker compose logs postgres
```

### OpenAI API Errors
```bash
# Verify your API key in .env:
cat .env | grep OPENAI_API_KEY

# Check you have credits: https://platform.openai.com/usage
```

### Port Already in Use
```bash
# Change PORT in .env:
PORT=8081

# Or stop conflicting service:
docker compose down
```

### Mock Client Being Used
If you see "Using mock LLM client", your `OPENAI_API_KEY` is not set correctly in `.env`.

## 📊 Sample Data

The database includes 33 pre-populated feedback records covering:
- **Product Areas**: billing, onboarding, performance, features, security, ui_ux, integrations
- **Regions**: NA, EU, APAC
- **Customer Tiers**: free, pro, enterprise
- **Sentiments**: positive, neutral, negative
- **Priorities**: 1-5 (low to critical)

## 🔄 Making Changes

### Adding New Code
1. Edit files in `internal/`
2. Run: `docker compose up --build`

### Adding Database Columns
1. Create `migrations/003_your_migration.sql`
2. Restart: `docker compose down && docker compose up --build`

### Changing LLM Provider
1. Implement `llm.Client` interface
2. Update `cmd/api/main.go` initialization

## 📚 Full Documentation

- See `README.md` for complete setup instructions
- See `ARCHITECTURE.md` for system design details

## 🎯 Next Steps

1. ✅ Verify health check works
2. ✅ Test with sample questions
3. ✅ Review generated SQL in logs
4. ✅ Examine insight responses
5. 🔧 Customize prompts in `internal/llm/prompts.go`
6. 🔧 Add your own product areas to seed data
7. 🔧 Implement authentication if needed

**Happy analyzing! 🎉**
