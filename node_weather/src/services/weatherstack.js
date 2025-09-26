let got;
const dotenv = require('dotenv');
const { logWeatherRequest } = require('../lib/logger.js');
const { normalizeWeatherStackResponse } = require('../lib/normalize.js');

dotenv.config();
class WeatherStackService {
    constructor() {
        this.apiKey = process.env.WEATHERSTACK_KEY;
        this.baseUrl = process.env.WEATHERSTACK_BASE_URL || 'http://api.weatherstack.com';
        this.timeout = 5000; // Wait 5 s
        
        if (!this.apiKey) throw new Error('WEATHERSTACK_KEY is missing in environment');
    }

    async getWeather(location) {
        const startTime = Date.now();
        
        try {
            if (!got) {
                got = (await import('got')).default;
            }
            const url = `${this.baseUrl}/current`;
            const searchParams = {
                access_key: this.apiKey,
                query: location
            };

            const response = await got(url, {
                searchParams,
                timeout: { request: this.timeout },
                retry: {
                    limit: 2,
                    methods: ['GET']
                }
            }).json();

            const responseTime = Date.now() - startTime;
            
            if (response.error) throw new Error(`WeatherStack error: ${response.error.info}`);
            
            
            const normalizedData = normalizeWeatherStackResponse(response);
            logWeatherRequest('weatherstack', location, responseTime, true);
            
            return normalizedData;
            
        } catch (error) {
            const responseTime = Date.now() - startTime;
            
            logWeatherRequest('weatherstack', location, responseTime, false, error.message);            
            throw new Error(`WeatherStack failed: ${error.message}`);
        }
    }
    async healthCheck() {
        try {
            await this.getWeather('London');
            return true;
        } catch (error) {
            return false;
        }
    }
}

const weatherStackService = new WeatherStackService();
module.exports = { weatherStackService };
