# ExpWeather - Weather API Aggregator

Weather query application with request aggregation that combines data from multiple weather APIs.

## Features

- **Multi-API Support**: Integrates with WeatherAPI and WeatherStack APIs
- **Request Aggregation**: Efficiently handles multiple requests
- **SQLite Database**: Stores weather data locally
- **Express.js API**: RESTful endpoints for weather queries
- **Request Queue**: Manages API requests with queuing system
- **Logging**: Comprehensive logging with Pino

## Prerequisites

- Node.js v20 or higher
- npm or yarn package manager

## Installation

1. Clone the repository:
```bash
git clone https://github.com/bygt/expweather.git
cd expweather
```

2. Install dependencies:
```bash
npm install
```

3. Create environment file:
```bash
cp .env.example .env
```

4. Configure your API keys in `.env`:
```env
WEATHERAPI_KEY=your_weatherapi_key_here
WEATHERSTACK_KEY=your_weatherstack_key_here

# Optional configurations
WEATHERAPI_BASE_URL=http://api.weatherapi.com/v1
WEATHERSTACK_BASE_URL=http://api.weatherstack.com
NODE_ENV=development
```

## Usage

### Development Mode
```bash
npm run dev
```

### Production Mode
```bash
npm start
```

### Testing
```bash
npm test
```

## API Endpoints

### Get Weather Data
```
GET /weather?q=<location>
```

**Example:**
```bash
curl "http://localhost:3000/weather?q=Istanbul"
```

**Response:**
```json
{
  "location": "Istanbul",
  "temperature": 22,
  "humidity": 65,
  "description": "Partly cloudy",
  "source": "weatherapi",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Health Check (Development only)
```
GET /weather/health
```

### Statistics (Development only)
```
GET /weather/stats
```

## Project Structure

```
expweather/
├── src/
│   ├── app.js              # Express app configuration
│   ├── index.js            # Application entry point
│   ├── core/
│   │   └── aggregator.js   # Weather data aggregation logic
│   ├── db/
│   │   └── sqlite.js       # SQLite database operations
│   ├── lib/
│   │   ├── logger.js       # Logging configuration
│   │   ├── normalize.js    # Data normalization utilities
│   │   └── queue.js        # Request queue management
│   ├── routes/
│   │   └── weather.js      # Weather API routes
│   └── services/
│       ├── weatherapi.js   # WeatherAPI service
│       └── weatherstack.js  # WeatherStack service
├── test.js                 # Test file
├── weather.sqlite          # SQLite database file
├── package.json
└── README.md
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `WEATHERAPI_KEY` | WeatherAPI service key | Required |
| `WEATHERSTACK_KEY` | WeatherStack service key | Required |
| `WEATHERAPI_BASE_URL` | WeatherAPI base URL | `http://api.weatherapi.com/v1` |
| `WEATHERSTACK_BASE_URL` | WeatherStack base URL | `http://api.weatherstack.com` |
| `NODE_ENV` | Environment mode | `production` |
| `PORT` | Server port | `3000` |
| `HOST` | Server host | `0.0.0.0` |
| `DB_PATH` | SQLite database path | `./weather.sqlite` |

### API Keys

You need to obtain API keys from:

1. **WeatherAPI**: [https://www.weatherapi.com/](https://www.weatherapi.com/)
2. **WeatherStack**: [https://weatherstack.com/](https://weatherstack.com/)

## Dependencies

### Production Dependencies
- `express`: Web framework
- `better-sqlite3`: SQLite database driver
- `got`: HTTP client for API requests
- `pino`: Logging library
- `dotenv`: Environment variable loader

### Development Dependencies
- `nodemon`: Development server with auto-restart
- `pino-pretty`: Pretty logging formatter

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin feature-name`
5. Submit a pull request

## License

ISC License - see LICENSE file for details

## Troubleshooting

### Common Issues

1. **API Key Errors**: Make sure your API keys are correctly set in `.env` file
2. **Port Already in Use**: Change the PORT in `.env` or kill the process using port 3000
3. **Database Issues**: Delete `weather.sqlite` file to reset the database

### Logs

The application uses structured logging. In development mode, logs are pretty-printed. In production, logs are in JSON format.

## Support

For issues and questions, please open an issue on GitHub.
