# Go Weather API

A professional weather query application that aggregates requests and fetches data from two different weather APIs, returning the average temperature with optimized cost management.

## Features

- **Request Aggregation**: Groups requests by location for up to 5 seconds to minimize API costs
- **Smart Batching**: Maximum 10 requests per location trigger immediate processing
- **Parallel API Calls**: Simultaneously fetches data from WeatherAPI.com and WeatherStack.com
- **SQLite Database**: Async logging of all weather queries
- **Clean Architecture**: Separation of concerns with handlers, services, and data layers
- **Environment Configuration**: Secure configuration management
- **Error Handling**: Standardized error responses with proper HTTP status codes

## Project Structure

```
goweather/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── config/config.go           # Configuration management
│   ├── database/sqlite.go         # Database operations
│   ├── handlers/weather.go        # HTTP handlers (HTTP layer)
│   ├── services/weather.go        # Business logic (Service layer)
│   └── clients/                   # External API clients
│       ├── weatherapi.go          # WeatherAPI.com client
│       └── weatherstack.go        # WeatherStack.com client
├── pkg/types/weather.go            # Data types and structures
├── test.go                         # Go integration test script
├── test.py                         # Python integration test script (legacy)
├── requirements.txt                # Python dependencies for testing (legacy)
├── .env.example                    # Environment configuration template
├── go.mod                          # Go module dependencies
├── go.sum                          # Dependency checksums
├── .gitignore                      # Git ignore rules
└── weather.sqlite                  # SQLite database file (auto-created)
```

## Installation & Setup

### Prerequisites

- **Go 1.21+** installed on your system
- Internet connection for API calls

### 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd goweather

# Install dependencies
go mod tidy
```

### 2. Environment Configuration

Copy the example environment file and configure your API keys:

```bash
# Copy the example file
cp .env.example .env

# Edit .env file and add your API keys
```

Update the `.env` file with your actual API keys:

```env
# Weather API Configuration
WEATHER_API_KEY=your_actual_weatherapi_key_here
WEATHER_STACK_KEY=your_actual_weatherstack_key_here

# Database Configuration
DATABASE_PATH=weather.sqlite

# Server Configuration
SERVER_PORT=8000
DEBUG_MODE=false (to view SQLite data, set DEBUG_MODE=true and request GET /queries endpoint)

# Aggregation Settings
MAX_REQUESTS=10
WAIT_TIME=5s

# API Timeout
API_TIMEOUT=10s
```

**Get your API keys from:**
- WeatherAPI.com: https://www.weatherapi.com/signup.aspx
- WeatherStack.com: https://weatherstack.com/signup/free

### 3. Run the Application

#### Option A: Docker (Recommended - Production Ready)

**Prerequisites:**
- Docker Desktop installed and running

**Steps:**
```bash
# 1. Create environment file from example
cp .env.example .env
# Edit .env file and add your actual API keys

# 2. Build and run with docker-compose
docker-compose up --build

# Or run in background
docker-compose up --build -d
```

**Docker Management:**
```bash
# View logs
docker-compose logs -f

# Stop containers
docker-compose down

# Stop and remove volumes (reset database)
docker-compose down -v
```

**Advantages:**
- ✅ **Isolated environment** - No Go installation required
- ✅ **Consistent deployment** - Works everywhere Docker runs
- ✅ **Production ready** - Easy scaling and deployment
- ✅ **Clean setup** - No local dependencies

#### Option B: Direct Go Run

**Option B1: Direct Run (Slower)**
```bash
# Compiles and runs (takes 10+ seconds due to dependencies)
go run cmd/server/main.go
```

**Option B2: Build and Run (Faster)**
```bash
# Build once (first build may take 10-15 seconds due to SQLite and logging dependencies)
go build -o goweather cmd/server/main.go

```

The server will start on port 8000 by default.

## API Usage

### Weather Endpoint

```bash
GET /weather?q=<location>
```

**Example Request:**
```bash
curl "http://localhost:8000/weather?q=Istanbul"
```

**Success Response (200 OK):**
```json
{
  "location": "Istanbul",
  "temperature": 25.5
}
```

**Error Response:**
```json
{
  "error": "MISSING_LOCATION",
  "code": 400,
  "message": "Location parameter 'q' is required"
}
```

### Debug Endpoint (DEBUG_MODE=true only)

```bash
GET /queries
```

**To view SQLite database data:**
1. Set `DEBUG_MODE=true` in your `.env` file
2. Restart the server
3. Make a request to `/queries` endpoint

**Example:**
```bash
curl "http://localhost:8000/queries"
```

### Health Check

```bash
GET /
```

## Request Aggregation Logic

The application implements smart request aggregation to minimize API costs:

### Aggregation Rules
1. **5-second Window**: Requests for the same location are held for up to 5 seconds
2. **Maximum 10 Requests**: Once 10 requests are queued for a location, processing triggers immediately
3. **Single API Call**: All aggregated requests share the result from one API call
4. **Parallel Processing**: Multiple locations can be processed simultaneously

### Example Scenarios

**Scenario 1: Single Request**
- Request for "Istanbul" arrives
- Timer starts (5 seconds)
- After 5 seconds: API call made, response returned
- Total time: ~6 seconds (5s wait + 1s API call)

**Scenario 2: Multiple Requests**
- Request 1 for "Istanbul" arrives → Timer starts
- Request 2 for "Istanbul" arrives after 2 seconds → Joins group
- Request 3 for "Istanbul" arrives after 3 seconds → Joins group
- After 5 seconds: Single API call, all 3 requests get the same response at the same time
- Total time: ~6 seconds after the first request arrived

**Scenario 3: Max Requests Reached**
- 10 requests for "Istanbul" arrive rapidly
- Processing triggers immediately (no wait)
- Single API call serves all 10 requests
- Total time: ~1 second

## Database Schema

SQLite database with `weather_queries` table:

```sql
CREATE TABLE weather_queries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location TEXT NOT NULL,
    service_1_temperature REAL NOT NULL,
    service_2_temperature REAL NOT NULL,
    request_count INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Testing

### Go Test Script

A comprehensive test script is included to verify the aggregation logic:

```bash
# No additional dependencies needed (uses Go standard library)
go run test.go

# Or build and run
go build test.go
./test.exe  # Windows
./test      # Linux/Mac
```

The test script verifies:
- **Istanbul Test**: 11 concurrent requests (10 immediate + 1 delayed)
- **Ankara Test**: 3 requests with 1-second intervals (aggregation test)

### Expected Test Behavior

- **Istanbul**: First 10 requests return immediately, 11th request waits ~5 seconds
- **Ankara**: All 3 requests aggregate and return together after ~5 seconds

## Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `WEATHER_API_KEY` | provided | WeatherAPI.com API key |
| `WEATHER_STACK_KEY` | provided | WeatherStack.com API key |
| `DATABASE_PATH` | `weather.sqlite` | SQLite database file path |
| `SERVER_PORT` | `8000` | HTTP server port |
| `DEBUG_MODE` | `false` | Enable debug endpoints |
| `MAX_REQUESTS` | `10` | Maximum requests per aggregation group |
| `WAIT_TIME` | `5s` | Aggregation wait time |
| `API_TIMEOUT` | `10s` | External API timeout |

## Architecture Details

### Clean Architecture Layers

1. **HTTP Layer** (`handlers/`): Request/response handling, validation
2. **Service Layer** (`services/`): Business logic, aggregation, orchestration
3. **Data Layer** (`database/`, `clients/`): External integrations

### Aggregation Implementation

- **Thread-Safe**: Uses mutexes for concurrent access
- **Memory Efficient**: Groups are cleaned up after processing
- **Scalable**: Can handle multiple locations simultaneously
- **Fault Tolerant**: Error handling for API failures

### External APIs

- **WeatherAPI.com**: Primary weather service (HTTPS)
- **WeatherStack.com**: Secondary weather service (HTTP only for free tier)

## Troubleshooting

### Common Issues

1. **Port Already in Use**
   ```bash
   # Change port in .env file
   SERVER_PORT=8001
   ```

2. **API Key Issues**
   - Verify API keys in `.env` file
   - Check API rate limits
   - Ensure internet connectivity

3. **Database Permissions**
   ```bash
   # Ensure write permissions for database file
   chmod 666 weather.sqlite
   ```

### Build and Performance

**Build Advantages:**
- ✅ **Speed**: Instant startup (1-2 seconds vs 10+ seconds)
- ✅ **Deployment**: Single executable file, no Go runtime needed
- ✅ **Production**: Better for production environments
- ✅ **Development**: Faster iteration during development

**Build Considerations:**
- 📦 **Storage**: Creates ~15-20MB executable file (SQLite + dependencies)
- 💾 **Memory**: No additional memory overhead vs `go run`
- 🔄 **Rebuild**: Need to rebuild after code changes
- 🗂️ **Artifacts**: Remember to add `*.exe` / binary to `.gitignore`

**Recommended Development Workflow:**
```bash
# Initial setup
go build -o server.exe cmd/server/main.go

# Development loop
# 1. Edit code
# 2. go build -o server.exe cmd/server/main.go
# 3. ./server.exe
# 4. Test
# 5. Repeat
```

### Logging

The application provides structured logging powered by Zerolog (similar to Pino.js):
- 🏗️ **Component-based**: server, weather, api_client, aggregation, database
- 📊 **Action-based**: startup, request, processing, completed, error
- ⏱️ **Performance**: Response times, request counts
- 🔍 **Context**: Location, user_id, temperature, error details
- 📋 **Structured**: JSON format for easy parsing and monitoring

## Production Considerations

- Set `DEBUG_MODE=false` in production
- Use proper API rate limiting
- Monitor database growth
- Implement log rotation
- Add health checks
- Consider using connection pooling for high load

## Reach me

If you have any questions, suggestions, or need support regarding this project:

- **📧 Email**: [buseyigit01@gmail.com](mailto:buseyigit01@gmail.com)
- **🐛 Issues**: Report bugs or issues via GitHub Issues
- **💬 Discussions**: Use GitHub Discussions for general questions
- **📱 LinkedIn**: [buse-yigit](https://linkedin.com/in/buse-yigit)
