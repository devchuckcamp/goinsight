# Free LLM Options Guide

GoInsight now supports **three FREE alternatives** to OpenAI! Choose the option that works best for you.

---

## 🆓 Option 1: Mock Client (Testing)

**Best for:** Testing the API structure without any external dependencies

### Setup
```env
LLM_PROVIDER=mock
```

### Features
- ✅ Works immediately, no setup
- ✅ Returns realistic response structure
- ✅ Perfect for testing and development
- ❌ No real AI insights

### Test It
```bash
docker compose down && docker compose up -d
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "What are billing issues?"}'
```

---

## ⚡ Option 2: Groq (FREE Cloud API)

**Best for:** Fast, free AI-powered insights without local setup

### Why Groq?
- ✅ **Completely FREE** tier (generous limits)
- ✅ **Super fast** inference (faster than OpenAI)
- ✅ Multiple open-source models (Llama, Mixtral, Gemma)
- ✅ OpenAI-compatible API
- ✅ No credit card required for free tier

### Setup Steps

1. **Get Free API Key**
   - Visit: https://console.groq.com/keys
   - Sign up (free, no credit card)
   - Create an API key

2. **Update `.env`**
   ```env
   LLM_PROVIDER=groq
   GROQ_API_KEY=gsk_your_actual_key_here
   LLM_MODEL=  # Optional: defaults to llama-3.3-70b-versatile
   ```

3. **Available Models** (all free)
   - `llama-3.1-70b-versatile` (recommended, very capable)
   - `llama-3.1-8b-instant` (faster, lighter)
   - `mixtral-8x7b-32768` (great for long context)
   - `gemma2-9b-it` (good balance)

4. **Restart & Test**
   ```bash
   docker compose down && docker compose up -d
   curl -X POST http://localhost:8080/api/ask \
     -H "Content-Type: application/json" \
     -d '{"question": "What are the most critical billing issues?"}'
   ```

### Rate Limits (Free Tier)
- ~30 requests per minute
- ~14,400 tokens per minute
- Perfect for development and small production workloads

---

## 🏠 Option 3: Ollama (Local, Offline)

**Best for:** Complete privacy, no internet required, unlimited usage

### Why Ollama?
- ✅ **100% FREE** forever
- ✅ Runs locally on your machine
- ✅ Works offline
- ✅ Complete data privacy
- ✅ No API keys needed
- ❌ Requires local resources (RAM/GPU)

### Setup Steps

1. **Install Ollama**
   - Download: https://ollama.ai
   - Or via command:
     ```bash
     # macOS/Linux
     curl -fsSL https://ollama.ai/install.sh | sh
     
     # Windows
     # Download installer from ollama.ai
     ```

2. **Pull a Model**
   ```bash
   # Recommended models
   ollama pull llama3          # 4.7GB, very capable
   ollama pull llama3.1        # 4.7GB, newer version
   ollama pull mistral         # 4.1GB, fast and good
   ollama pull codellama       # 3.8GB, good at SQL
   ```

3. **Start Ollama**
   ```bash
   ollama serve
   # Runs on http://localhost:11434
   ```

4. **Update `.env`**
   ```env
   LLM_PROVIDER=ollama
   OLLAMA_URL=http://localhost:11434
   LLM_MODEL=llama3
   ```

5. **Docker Note**: If using Docker, Ollama must be accessible
   - Mac/Windows Docker Desktop: Use `http://host.docker.internal:11434`
   - Linux Docker: Use `http://172.17.0.1:11434`
   
   ```env
   # For Docker on Mac/Windows
   OLLAMA_URL=http://host.docker.internal:11434
   ```

6. **Restart & Test**
   ```bash
   docker compose down && docker compose up -d
   curl -X POST http://localhost:8080/api/ask \
     -H "Content-Type: application/json" \
     -d '{"question": "Show me enterprise customer complaints"}'
   ```

### Performance Tips
- Use smaller models for faster responses (llama3.1:8b)
- GPU acceleration makes it much faster
- First request is slower (model loading)

---

## 📊 Comparison

| Feature | Mock | Groq | Ollama | OpenAI |
|---------|------|------|--------|--------|
| **Cost** | Free | Free | Free | Paid |
| **Setup** | None | API Key | Install | API Key |
| **Speed** | Instant | Very Fast | Medium | Fast |
| **Quality** | N/A | Excellent | Good-Excellent | Excellent |
| **Privacy** | Local | Cloud | Local | Cloud |
| **Offline** | ✅ | ❌ | ✅ | ❌ |
| **Limits** | None | 30 req/min | Hardware | Pay per use |

---

## 🎯 Recommendations

### For Quick Testing
→ **Use Mock Client** (`LLM_PROVIDER=mock`)

### For Development with Real AI
→ **Use Groq** (free, fast, no setup hassle)

### For Production (Small Scale)
→ **Use Groq** (generous free tier)

### For Privacy/Offline
→ **Use Ollama** (completely local)

### For Production (Large Scale)
→ **Use OpenAI** (most reliable, paid)

---

## 🚀 Quick Start Commands

### Test with Mock (Immediate)
```bash
# Update .env
echo "LLM_PROVIDER=mock" >> .env

# Restart
docker compose down && docker compose up -d

# Test
curl -X POST http://localhost:8080/api/ask \
  -H "Content-Type: application/json" \
  -d '{"question": "Show billing issues"}'
```

### Switch to Groq (2 minutes)
```bash
# 1. Get key from https://console.groq.com/keys
# 2. Update .env
cat >> .env << EOF
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_your_key_here
LLM_MODEL=llama-3.1-70b-versatile
EOF

# 3. Restart
docker compose down && docker compose up -d
```

### Switch to Ollama (5 minutes)
```bash
# 1. Install Ollama from ollama.ai
# 2. Pull model
ollama pull llama3

# 3. Start Ollama
ollama serve &

# 4. Update .env
cat >> .env << EOF
LLM_PROVIDER=ollama
OLLAMA_URL=http://host.docker.internal:11434
LLM_MODEL=llama3
EOF

# 5. Restart
docker compose down && docker compose up -d
```

---

## 🐛 Troubleshooting

### Groq: "API key invalid"
- Verify key at https://console.groq.com/keys
- Ensure no extra spaces in .env file
- Key format: `gsk_...`

### Ollama: "failed to make request"
- Check Ollama is running: `ollama list`
- Verify URL is accessible from Docker
- Use `http://host.docker.internal:11434` for Docker on Mac/Windows

### All: "Invalid LLM_PROVIDER"
- Check spelling in .env
- Valid values: `mock`, `groq`, `ollama`, `openai`
- Restart after changing: `docker compose down && docker compose up -d`

---

## 💡 Tips

1. **Start with Mock** to verify the system works
2. **Try Groq next** - it's free and gives real AI insights
3. **Use Ollama** if you want offline/private usage
4. **Upgrade to OpenAI** only if you need the absolute best quality

All three free options work great for most use cases!

---

## 🔒 Security & Production Precautions

### API Key Management
- **Never commit API keys** to version control
- Use environment variables or secure secret managers
- Rotate keys regularly and monitor usage
- Set up alerts for unusual API consumption

### Rate Limiting & Monitoring
- Groq has generous free limits but monitor usage
- Implement client-side rate limiting to prevent quota exhaustion
- Use API monitoring tools to track costs and performance
- Set up fallback to mock client if API limits are reached

### Data Privacy Considerations
- Review LLM provider privacy policies (Groq, OpenAI)
- Avoid sending sensitive customer data through external APIs
- Consider data residency requirements for compliance
- Use local Ollama for maximum privacy control

### Production Deployment Tips
- Test thoroughly with mock client before production
- Implement circuit breakers for API failures
- Use structured logging for API request/response auditing
- Plan for API provider changes (keys, endpoints, limits)
