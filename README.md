# Weather Query Application

This repository contains two implementations of a weather query application - one built with Node.js and another with Go. Both applications provide identical functionality for querying weather data from multiple services.

### Overview

The weather query application aggregates temperature data from two different weather services (weatherapi.com and weatherstack.com) and returns the average temperature for a requested location. The application implements request aggregation to minimize API costs by grouping requests for the same location.

## Project Structure

### Node.js Implementation (`node_weather/`)
- Express.js web server
- Modular architecture with separate services, routes, and database layers
- Request queue management for aggregation using Map-based caching
- SQLite integration with prepared statements and connection pooling
- Event-driven architecture with async/await patterns
- Built-in logging system with configurable levels

### Go Implementation (`go_weather/`)
- Standard library HTTP server with custom routing
- Clean architecture with handlers, services, and clients
- Goroutine-based concurrency for request handling
- SQLite database with GORM-like patterns and connection management
- Mutex synchronization for thread-safe request aggregation
- Structured logging with context propagation
