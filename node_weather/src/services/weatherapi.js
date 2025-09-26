let got;
const dotenv = require('dotenv');
const { logWeatherRequest } = require('../lib/logger.js');
const { normalizeWeatherApiResponse } = require('../lib/normalize.js');

dotenv.config();

class WeatherApiService {
    constructor() {
        this.apiKey = process.env.WEATHERAPI_KEY;
        this.baseUrl = process.env.WEATHERAPI_BASE_URL || 'http://api.weatherapi.com/v1';
        this.timeout = 5000; //max 5 s
        if (!this.apiKey) throw new Error('WEATHERAPI_KEY is missing in environment');
    }
    async getWeather(location) {
        const startTime = Date.now();
        
        try {
            if (!got) {
                got = (await import('got')).default;
            }
            const url = `${this.baseUrl}/forecast.json`;
            const searchParams = {
                key: this.apiKey,
                q: location,
                days: 1,
                aqi: 'no',   
                alerts: 'no' 
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
            const normalizedData = normalizeWeatherApiResponse(response);
            
            logWeatherRequest('weatherapi', location, responseTime, true);
            
            return normalizedData;
            
        } catch (error) {
            const responseTime = Date.now() - startTime;
            
            logWeatherRequest('weatherapi', location, responseTime, false, error.message);
            
            throw new Error(`WeatherAPI failed: ${error.message}`);
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

const weatherApiService = new WeatherApiService();
module.exports = { weatherApiService };
