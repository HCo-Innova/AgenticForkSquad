# 07-LLM-INTEGRATION.md

```markdown
# 07-LLM-INTEGRATION.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** LLM Integration Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md), 
[03-SYSTEM-ARCHITECTURE.md](03-SYSTEM-ARCHITECTURE.md)

---

## 📖 Table of Contents

1. [LLM Integration Overview](#llm-integration-overview)
2. [Vertex AI Models Overview](#vertex-ai-models-overview)
3. [Vertex AI Client - gemini-2.5-pro](#vertex-ai-client---gemini-25-pro)
4. [Common Client Interface](#common-client-interface)
5. [Rate Limiting (Vertex AI Quotas)](#️-rate-limiting-vertex-ai-quotas)
6. [Error Handling](#error-handling)
7. [Cost Tracking](#cost-tracking)
8. [Best Practices](#best-practices)

---

## 🤖 LLM Integration Overview

### Purpose

**Enable AI agents to analyze queries, generate optimizations, and make 
intelligent decisions through a unified Vertex AI Client using multiple models.**

**Vertex AI Models (por defecto):**
- **gemini-2.5-pro** - Planificación/razonamiento crítico y QA
- **gemini-2.5-flash** - Generación de SQL/código y pruebas (baja latencia)
- **gemini-2.0-flash** - Tareas masivas de bajo riesgo y boilerplate

---

### Integration Architecture

**Layer Organization:**

```
┌─────────────────────────────────────┐
│  Agent Layer (Infrastructure)       │
│  - Agent (gemini-2.5-pro)           │
│  - Agent (gemini-2.5-flash)         │
│  - Agent (gemini-2.0-flash)         │
└──────────────┬──────────────────────┘
               │ uses
               ▼
┌─────────────────────────────────────┐
│  LLM Client Layer (Infrastructure)  │
│  - VertexClient                     │
└──────────────┬──────────────────────┘
               │ implements
               ▼
┌─────────────────────────────────────┐
│  LLMClient Interface (Domain)       │
│  - SendMessage()                    │
│  - SendMessageWithJSON()            │
└──────────────┬──────────────────────┘
               │ calls
               ▼
┌─────────────────────────────────────┐
│  Vertex AI APIs (External)          │
│  - aiplatform.googleapis.com        │
└─────────────────────────────────────┘
```

**Separation of Concerns:**
- Agents: Business logic (what to ask)
- Clients: Technical integration (how to ask)
- Interface: Contract (standardization)
- APIs: Vertex AI endpoints (model-specific)

---

### Common Workflow

**All agents follow same LLM interaction pattern:**

**Step 1: Prepare Context**
```
Gather information:
- EXPLAIN plan output
- Schema information
- Table statistics
- Query text
- Business constraints
```

**Step 2: Build Prompt**
```
Construct prompt:
- System prompt (role definition)
- User prompt (task description)
- Context data (EXPLAIN, schema, etc.)
- Output format specification (JSON)
```

**Step 3: Call LLM API**
```
Send request:
- Authenticate with API key
- Set parameters (temperature, max_tokens, etc.)
- Include prompt
- Handle timeout
```

**Step 4: Parse Response**
```
Extract data:
- Strip markdown formatting
- Parse JSON
- Validate schema
- Handle errors
```

**Step 5: Validate & Use**
```
Business validation:
- Check required fields
- Validate data types
- Verify business rules
- Return structured result
```

---

## 📊 Vertex AI Models Overview

### Feature Matrix (por defecto)

| Feature | gemini-2.5-pro | gemini-2.5-flash | gemini-2.0-flash |
|---------|-----------------|------------------|------------------|
| **Context Window** | 1M tokens | 1M tokens | 1M tokens |
| **Max Output** | 8K tokens | 8K tokens | 8K tokens |
| **JSON Mode** | Native (`response_mime_type`) | Native | Native |
| **Rate Limit** | Project-dependent | Project-dependent | Project-dependent |
| **Latency (P50)** | 2-3 sec | ~1-2 sec | ~1 sec |
| **Best For** | Planificación/QA | Generación operativa | Tareas masivas |

**Context Window Implications (Gemini 1M):**
 - Suficiente para EXPLAIN detallados y esquema relevante
 - No es necesario incluir todo el DDL; priorizar contexto crítico

---

### Cost Comparison (Typical AFS Usage)

**Assumptions per agent execution:**
- Input: 5,000 tokens (EXPLAIN + schema + prompt)
- Output: 1,000 tokens (JSON response)

**Cost per execution:**

**gemini-2.5-pro:**
```
Input:  (5,000 / 1,000,000) * $10.00 = $0.050
Output: (1,000 / 1,000,000) * $30.00 = $0.030
Total: $0.080 per execution
```
Nota: Ajustar con tu plan de precios actual.

---

## 🔵 Deprecated: Legacy Client (removed)

### API Configuration

Este cliente fue deprecado. El sistema usa exclusivamente modelos Gemini. Ver sección "Vertex AI Client - gemini-2.5-pro".

### Authentication

Use ADC (Application Default Credentials).

**Environment Variables:**
```
VERTEX_PROJECT_ID=your-gcp-project
VERTEX_LOCATION=us-central1
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

**Header Format:**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

---

### Request Format

**Endpoint:**
```
POST https://{LOCATION}-aiplatform.googleapis.com/v1/projects/{PROJECT}/locations/{LOCATION}/publishers/google/models/gemini-2.5-pro:generateContent
```

**Body:**
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {
          "text": "Analyze this EXPLAIN plan... Output ONLY the JSON."
        }
      ]
    }
  ],
  "systemInstruction": {
    "parts": [
      {
        "text": "You are an expert PostgreSQL DBA..."
      }
    ]
  },
  "generationConfig": {
    "temperature": 0.0,
    "maxOutputTokens": 4096
  }
}
```

**Parameter Descriptions:**

| Parameter | Value | Purpose |
|-----------|-------|---------|
| maxOutputTokens | 4096 | Maximum response length |
| temperature | 0.0 | Deterministic (no randomness) |
| systemInstruction | object | Role definition |
| contents | array | Conversation history |

---

### Response Format

**Success Response:**
```json
{
  "candidates": [
    {
      "content": {
        "role": "model",
        "parts": [
          {
            "text": "{\n  \"insights\": [..."
          }
        ]
      },
      "finishReason": "STOP"
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 150,
    "candidatesTokenCount": 200,
    "totalTokenCount": 350
  }
}
```

---

### JSON Response Handling

**Model JSON Behavior:**
- Relies on prompt engineering to output JSON.
- Prompt should explicitly ask for JSON-only output.
- May still include markdown fences if not explicitly told otherwise.

**Parsing Strategy:**
1.  Concatenate all `text_delta` parts from the stream.
2.  Trim whitespace from the final string.
3.  Extract content from ```json ... ``` markdown fences if present.
4.  Parse the extracted string as JSON.
5.  If parsing fails, log the error and the raw string for debugging.

---

### Error Responses

Handled by the standard Google Cloud client libraries, which will raise exceptions for HTTP status codes like 4xx (e.g., `PermissionDenied`) or 5xx (`InternalServerError`).

---

### Model-Specific Notes (Legacy Client)

**System Prompt:**
- Legacy client on Vertex AI responded well to the `system` parameter for setting context and persona.

**Tool Use:**
- While the legacy client supported tool use, for AFS we use structured JSON output for simplicity and reliability.

---

## 🟢 Deprecated: Legacy Client 2 (removed)

### API Configuration

Este cliente fue deprecado. El sistema usa exclusivamente modelos Gemini. Ver sección "Vertex AI Client - gemini-2.5-pro".

---

### Authentication


### API Configuration

**Provider:** Vertex AI  
**Base URL:** https://aiplatform.googleapis.com  
**Model:** gemini-2.5-pro  
**Documentation:** https://cloud.google.com/vertex-ai/docs

---

### Authentication

Use ADC (Application Default Credentials), same as otros modelos Gemini.

---

### Request Format

**Endpoint:**
```
POST https://{LOCATION}-aiplatform.googleapis.com/v1/projects/{PROJECT}/locations/{LOCATION}/publishers/google/models/gemini-2.5-pro:generateContent
```

**Body (with JSON Mode):**
```
{
  "contents": [
    {
      "parts": [
        {
          "text": "You are an elite Performance Engineer... Design an advanced optimization strategy... Respond with ONLY a valid JSON object."
        }
      ]
    }
  ],
  "generation_config": {
    "temperature": 0.0,
    "max_output_tokens": 8192,
    "response_mime_type": "application/json"
  }
}
```

**Parameter Descriptions:**

| Parameter | Value | Purpose |
|-----------|-------|---------|
| contents | array | Conversation messages |
| temperature | 0.0 | Deterministic responses |
| max_output_tokens | 8192 | Max response length |
| response_mime_type | application/json | **Enables native JSON mode** |

---

### JSON Mode

**Structured JSON Output:**
- Setting `response_mime_type` to `application/json` instructs Gemini to output a valid JSON object.
- This is more reliable than prompt engineering alone.
- The prompt must still instruct the model to produce JSON.

**Benefits:**
- Eliminates parsing errors from markdown fences or extra text.
- The `text` field in the response is a guaranteed-to-be-valid JSON string.

---

### Response Format

**Success Response (JSON Mode):**
```json
{
  "candidates": [
    {
      "content": {
        "role": "model",
        "parts": [
          {
            "text": "{\n  \"proposal_type\": \"materialized_view\",\n  \"sql_commands\": [...],\n  \"rationale\": \"...\"\n}"
          }
        ]
      },
      "finishReason": "STOP"
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 4832,
    "candidatesTokenCount": 956,
    "totalTokenCount": 5788
  }
}
```

**Extract Text:**
```
// The response is a clean JSON string, no extra parsing needed.
json_string = response.candidates[0].content.parts[0].text
parsed_object = JSON.parse(json_string)
```

---

### Error Responses

Standard Google Cloud / Vertex AI errors. A `400 Bad Request` can occur if the model is unable to produce JSON that conforms to the prompt's instructions.

---

### Model-Specific Notes (gemini-2.5-pro)

**Function Calling:**
- Gemini has robust function calling capabilities, which could be a future enhancement for AFS to allow agents to dynamically request more information.

**Streaming Responses:**
- Gemini supports streaming for real-time token generation, which is good for UIs but not essential for the batch processing in AFS.

**Seed Parameter:**
- Improves reproducibility for testing by using the same seed for generation.
- Optional enhancement for consistency.

---

## 🔗 Common Client Interface

### LLMClient Interface

**Purpose:**  
Standardize LLM interactions across Vertex AI models.

**Interface Definition (Conceptual):**

```
Interface LLMClient:
  
  Method: SendMessage(prompt string, systemPrompt string) 
          (response string, error)
  
  Purpose: Send basic text prompt
  
  Parameters:
    - prompt: User message
    - systemPrompt: System/role definition
  
  Returns:
    - response: Raw text response
    - error: If request failed
  
  ---
  
  Method: SendMessageWithJSON(prompt string, systemPrompt string) 
          (jsonResponse map, error)
  
  Purpose: Send prompt expecting JSON response
  
  Parameters:
    - prompt: User message (must mention JSON)
    - systemPrompt: System prompt
  
  Returns:
    - jsonResponse: Parsed JSON as map
    - error: If request or parsing failed
  
  Process:
    1. Send request to LLM
    2. Extract response text
    3. Parse JSON
    4. Validate structure
    5. Return parsed object
  
  ---
  
  Method: GetUsage() (inputTokens int, outputTokens int)
  
  Purpose: Get token usage for last request
  
  Returns:
    - inputTokens: Prompt tokens
    - outputTokens: Response tokens
  
  Usage: Cost tracking
```

---

### Implementation Structure

**Client Implementation:**

**VertexClient:**
- Implements: LLMClient interface
- Dependencies: Google Auth, HTTP client, config (project, location, models)
- Methods: SendMessage, SendMessageWithJSON, GetUsage
- Specifics: Vertex AI API format

**Shared Logic:**
- HTTP timeout handling
- Retry with exponential backoff
- Token counting
- Error wrapping with context

---

### Usage in Agents

**Agent Dependency:**

```
Gemini25ProAgent:
  Dependencies:
    - llmClient: LLMClient (interface)
    - mcpClient: MCPClient
    - config: Config
  
  Constructor injection:
    agent = NewGemini25ProAgent(vertexClient, mcpClient, config)
  
  Donde vertexClient es un VertexClient (model=gemini-2.5-pro) implementando LLMClient
```

**Agent Method Example:**

```
Gemini25ProAgent.AnalyzeTask(task):
  1. Get EXPLAIN plan via MCP
  2. Build prompt with context
  3. Call llmClient.SendMessageWithJSON(prompt, systemPrompt)
  4. Receive parsed JSON response
  5. Map to AnalysisResult domain entity
  6. Return result
```

**Testability:**

```
Testing agents:
  1. Create mock LLMClient
  2. Inject into agent
  3. Test agent logic without real API calls
  
Mock returns predefined JSON responses
Verify agent parses and validates correctly
```

---

## ⏱️ Rate Limiting (Vertex AI Quotas)

### Vertex AI Quotas

**Gemini (Free Tier):**
```
Requests: 60 per minute
         1,500 per day

```

---

### AFS Usage Patterns

**Typical Task:**
```
3 agents working in parallel
Each agent makes 2-3 LLM calls:
  - 1 for analysis
  - 1 for proposal generation
  - (optional) 1 for refinement

Total LLM calls per task: 6-9
```

**Concurrent Tasks:**
```
If 5 tasks running simultaneously:
  5 tasks × 3 agents × 2 calls = 30 LLM calls

Spread over ~3 minutes (agent workflow time):
  30 calls / 3 minutes = 10 calls/minute

Well within model limits (by Vertex quotas)
```

**Burst Scenarios:**
```
10 tasks submitted simultaneously:
  10 tasks × 3 agents = 30 agents starting
  30 agents × 1 initial analysis call = 30 calls

If all hit at same second:
  30 calls/minute instantaneous
  
Still within quotas (project/region specific in Vertex AI)
```

---

### Rate Limit Handling

**Detection:**

**HTTP Status Code:**
```
429 Too Many Requests
```

**Response Body:**
```
Vertex error indicating quota exceeded (RESOURCE_EXHAUSTED)
```

**Retry Strategy:**

**Exponential Backoff:**
```
Attempt 1: Immediate request
Attempt 2: Wait 1 second, retry
Attempt 3: Wait 2 seconds, retry
Attempt 4: Wait 4 seconds, retry
Attempt 5: Wait 8 seconds, retry

Max attempts: 5
Max total wait: 15 seconds

Formula: wait_time = base_delay × (2 ^ (attempt - 2))
  where base_delay = 1 second
```

**Jitter:**
```
Add randomness to prevent thundering herd:
  wait_time = base_wait × (0.5 + random(0, 1))
  
Example:
  Base wait: 4 seconds
  With jitter: 2-6 seconds random
```

---

### Request Queueing (Advanced)

**Purpose:**  
Smooth out burst requests to stay within rate limits.

**Conceptual Implementation:**

**Per-Provider Queue:**
```
VertexClient maintains internal queue:
  - Rate limits por proyecto/ubicación
  - Token bucket algorithm
  - Queue si el bucket está vacío
  - Procesa al reponerse tokens

Beneficios:
  - Previene 429
  - Backpressure automático
  - Fair queuing (FIFO)
```

**Token Bucket:**
```
Bucket capacity: 50 tokens (requests)
Refill rate: 50 tokens per minute (0.83/second)

When request arrives:
  IF bucket has tokens:
    Remove 1 token
    Execute request immediately
  ELSE:
    Queue request
    Wait for token refill
    Process from queue
```

---

## ❌ Error Handling

### Error Categories

**Transient Errors (Retry):**
- 429 Rate Limit
- 500 Internal Server Error
- 503 Service Unavailable
- 529 Overloaded
- Network timeouts

**Permanent Errors (Fail Fast):**
- 401 Unauthorized (invalid API key)
- 400 Bad Request (invalid parameters)
- 400 Context Length Exceeded
- Malformed JSON response (after retries)

**Partial Errors (Log & Continue):**
- Safety filter triggered (Gemini)
- Response truncated (max tokens reached)
- JSON parsing errors (attempt recovery)

---

### Error Response Handling

**Structured Error Information (Vertex AI):**

```
Conceptual error structure:
{
  model: "gemini25pro" | "gemini25flash" | "gemini20flash"
  endpoint: "aiplatform.googleapis.com"
  error_type: "rate_limit" | "auth" | "timeout" | ...
  status_code: 429
  message: "Quota exceeded"
  retry_after: 60
  request_id: "req_abc123"
  timestamp: "2024-01-15T10:30:00Z"
}
```

**Logging:**

```
Log all errors with context:
  - Model name
  - Agent type
  - Task ID
  - Request payload (truncated)
  - Full error response
  - Retry attempts made
  
Critical errors (non-retryable):
  - Alert via monitoring
  - May indicate configuration issue
```

---

### Timeout Configuration

**Request Timeouts:**

**Analysis Calls:**
```
Timeout: 120 seconds (2 minutes)
Reason: EXPLAIN analysis can be complex
```

**Proposal Generation:**
```
Timeout: 60 seconds (1 minute)
Reason: Simpler task, faster expected
```

**Refinement Calls:**
```
Timeout: 30 seconds
Reason: Quick follow-up, should be fast
```

**Network Timeout:**
```
Connection timeout: 10 seconds
Read timeout: As per call type above
```

---

### Fallback Strategies

**Scenario 1: Specific Model Unavailable**

**Problem:** Un modelo Gemini específico no disponible

**Strategy:**
```
One model unavailable → Continue with remaining models
Consensus with 2 proposals instead of 3
Mark model as temporarily unavailable
Log incident for review
```

---

**Scenario 2: All Providers Rate Limited**

**Problem:** Burst of tasks exceeds all rate limits

**Strategy:**
```
Queue tasks in pending state
Process sequentially as rate limits reset
User sees "waiting for capacity" message
Background job processes queue
```

---

**Scenario 3: Response Parsing Failure**

**Problem:** LLM returns malformed JSON

**Strategy:**
```
Attempt 1: Standard JSON parse
Attempt 2: Strip markdown, retry parse
Attempt 3: Extract between {...}, retry
Attempt 4: Re-prompt LLM with error feedback
Attempt 5: Mark agent as failed, continue with others
```

---

## 💰 Cost Tracking

### Token Counting

**Input Tokens:**
```
Approximate counting:
  - System prompt: ~200 tokens
  - User prompt base: ~100 tokens
  - EXPLAIN plan: ~500-2000 tokens
  - Schema info: ~200-1000 tokens
  
Total input: 1,000 - 3,500 tokens typical
```

**Output Tokens:**
```
JSON responses:
  - Analysis: ~500-800 tokens
  - Proposal: ~300-600 tokens
  
Total output: 800-1,400 tokens typical
```

**Tracking per Request:**

```
After each LLM call:
  Store in database:
    - Agent type
    - Task ID
    - Input token count
    - Output token count
    - Cost (calculated)
    - Timestamp

Aggregate by:
  - Task (total cost per task)
  - Agent (cost per agent type)
  - Day/Week/Month (usage trends)
```

---

### Cost Calculation

**Formula:**
```
cost = (input_tokens / 1,000,000 × input_price) +
       (output_tokens / 1,000,000 × output_price)
```

**Example (Gemini):**
```
Input: 2,000 tokens
Output: 1,000 tokens

cost = (2,000 / 1,000,000 × $3) + (1,000 / 1,000,000 × $15)
     = $0.006 + $0.015
     = $0.021
```

---

### Budget Monitoring

**Daily Budget:**
```
Set limit: $10/day
Track cumulative spend
If approaching limit:
  - Alert administrators
  - Queue low-priority tasks
  - Continue high-priority tasks
If exceeded:
  - Pause new tasks
  - Complete in-progress tasks
  - Resume next day
```

**Monthly Budget:**
```
Set limit: $200/month
Weekly check-ins
Forecast based on current usage
Adjust task priority if needed
```

---

### Cost Optimization

**Strategies:**

**1. Prompt Optimization:**
```
Reduce input tokens:
  - Truncate EXPLAIN plans (keep essential info)
  - Summarize schema (not full DDL)
  - Remove redundant context

Savings: 20-30% reduction in input tokens
```

**2. Model Selection:**
```
Route by priority:
  - High priority: gemini-2.5-pro (calidad)
  - Medium priority: gemini-2.5-flash
  - Low priority: gemini-2.0-flash (cost-effective)

Savings: 40-60% on low-priority tasks
```

**3. Caching:**
```
Cache schema info:
  - Same schema used across tasks
  - Don't regenerate per task
  - Share across agents

Savings: Minimal (schema tokens small %)
```

**4. Output Length Limits:**
```
Set max_tokens to reasonable values:
  - Analysis: 2,000 tokens max
  - Proposal: 1,500 tokens max
  
Prevents unnecessarily long responses
Savings: 10-15% on output costs
```

---

## ✅ Best Practices

### Prompt Engineering

**Clear Instructions:**
```
✓ "Respond ONLY with valid JSON"
✓ "Include these exact fields: insights, proposed_actions, confidence"
✓ "confidence must be a number between 0 and 1"

✗ "Give me your analysis"
✗ "Respond with structured data"
```

**Output Format Specification:**
```
Always provide example JSON structure in prompt:

Example:
{
  "insights": ["insight1", "insight2"],
  "proposed_actions": ["action1"],
  "confidence": 0.85
}

Benefits:
  - LLM understands exact format
  - Reduces parsing errors
  - Consistent responses
```

**Context Ordering:**
```
Order information by importance:
  1. Task description (what to do)
  2. Data to analyze (EXPLAIN, schema)
  3. Output format (JSON schema)
  4. Constraints (limitations)

LLMs weight earlier content higher
```

---

### Response Validation

**Schema Validation:**
```
After parsing JSON:
  1. Check required fields present
  2. Validate types (string, number, array)
  3. Validate ranges (confidence 0-1)
  4. Validate enums (proposal_type values)
  5. Check array lengths (non-empty where required)
```

**Business Logic Validation:**
```
Beyond schema:
  - SQL commands syntactically valid
  - Confidence realistic (not always 1.0)
  - Insights specific (not generic)
  - Proposed actions actionable
```

**Graceful Degradation:**
```
If optional field missing:
  - Use default value
  - Log warning
  - Continue processing

If required field missing:
  - Attempt re-prompt (once)
  - If still fails, mark agent as failed
  - Continue with other agents
```

---

### Testing Strategies

**Mock LLM Responses:**
```
Create test fixtures:
  - analysis_response_gemini25pro.json
  - proposal_response_gemini.json
  
Use in unit tests:
  - No real API calls
  - Fast execution
  - Deterministic results
  - Test parsing logic
```

**Integration Tests:**
```
Occasional real API calls:
  - Verify prompts still work
  - Catch API changes
  - Validate parsing
  
Run in CI weekly (not per commit)
Use separate API keys (test quota)
```

**Prompt Regression Testing:**
```
When changing prompts:
  1. Test with historical queries
  2. Compare response quality
  3. Ensure no degradation
  4. Version prompts (track changes)
```

---

### Security Practices

**API Key Management:**
```
✓ Store in environment variables
✓ Never commit to git
✓ Rotate periodically (quarterly)
✓ Use different keys per environment
✓ Set spending limits per key

✗ Hardcode in source
✗ Log full API keys
✗ Share keys across projects
```

**Request Logging:**
```
Log requests securely:
  ✓ Truncate sensitive data in queries
  ✓ Mask API keys in logs
  ✓ Log request IDs (for support)
  
  ✗ Log full SQL with PII
  ✗ Log API keys
  ✗ Log in plaintext (encrypt logs)
```

**Rate Limit Respect:**
```
✓ Implement backoff
✓ Queue requests if needed
✓ Monitor usage
✓ Alert before hitting limits

✗ Aggressive retry without backoff
✗ Ignore rate limit responses
```

---

## 🎯 Summary

This LLM integration provides:

**Vertex AI Client with 3 Models:**
- gemini-2.5-pro - Planner/QA: desambiguación, planificación, verificación de SQL/código
- gemini-2.5-flash - Generación/ejecución: SQL/código, transformaciones y pruebas
- gemini-2.0-flash - Bajo costo: tareas masivas, boilerplate y refactors simples

**Unified Interface:**
- LLMClient abstraction
- SendMessage and SendMessageWithJSON methods
- Provider-agnostic agent code
- Easy testing with mocks

**Robust Error Handling:**
- Transient error retry (exponential backoff)
- Permanent error fast-fail
- Partial failure tolerance
- Comprehensive logging

**Cost Management:**
- Token tracking per request
- Cost calculation per provider
- Budget monitoring and alerts
- Optimization strategies

**Rate Limiting:**
- Respects provider limits (50-500 req/min)
- Exponential backoff on 429
- Optional request queueing
- Jitter to prevent thundering herd

**Best Practices:**
- Clear prompt engineering
- Response validation
- Secure API key management
- Testing strategies (mocks + integration)

**Typical Costs:**
- $0.11 per task (3 agents)
- $11 per 100 tasks
- $112 per 1,000 tasks

---

**Related Documentation:**
- Previous: [04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md) 
  - Agent specializations and prompt templates
- Previous: [06-TIGER-CLOUD-MCP.md](06-TIGER-CLOUD-MCP.md) 
  - Database integration
- See also: [03-SYSTEM-ARCHITECTURE.md](03-SYSTEM-ARCHITECTURE.md) 
  - Where LLM clients fit in architecture

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```

---

## ✅ Documento `07-LLM-INTEGRATION.md` Creado

**Contenido incluido:**
- ✅ **Cliente unificado Vertex AI** con 3 modelos (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)
- ✅ **Autenticación** vía ADC (Google Application Default Credentials)
- ✅ **Request/Response** a Vertex con selección de modelo
- ✅ **Interfaz común LLMClient** con soporte multi-modelo
- ✅ **Rate limiting** (límites por modelo/región en Vertex + estrategias de manejo)
- ✅ **Error handling** (categorización, retry logic, fallbacks)
- ✅ **Cost tracking** (fórmulas, ejemplos, optimizaciones)
- ✅ **Best practices** (prompts, validation, testing, security)
- ✅ **Model comparison** (features, pricing, context windows)
- ✅ **JSON mode** bajo Vertex
- ✅ **Sin código** (todo conceptual con ejemplos estructurados)

**Características destacadas:**
- Ejemplos de request/response por cada modelo en Vertex
- Métricas y límites por modelo/región
- Exponential backoff con jitter
- Token bucket algorithm conceptual
- Security best practices

**Documentos restantes:**
- 09-FRONTEND-COMPONENTS.md
- 10-DEVELOPMENT-WORKFLOW.md
- 11-DEPLOYMENT-STRATEGY.md

**¿Continuamos con alguno de estos?** 🚀