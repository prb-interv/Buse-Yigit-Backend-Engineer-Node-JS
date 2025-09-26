const { logAggregation } = require('../lib/logger.js');
const { normalizeLocation, calculateAverageTemperature } = require('../lib/normalize.js');
const { weatherApiService } = require('../services/weatherapi.js');
const { weatherStackService } = require('../services/weatherstack.js');
const { getQueue } = require('../lib/queue.js');


class RequestAggregator {
    constructor() {
        this.groups = new Map(); 
        this.windowMs = parseInt(process.env.AGGREGATION_WINDOW_MS) || 5000; 
        this.maxUsers = parseInt(process.env.MAX_USERS_PER_LOCATION) || 10;
    }

    /**
     * Add a new weather request to the group
     * @param {string} location - Place name from user
     * @param {Function} callback - Function to call when we get the result
     */
    async addRequest(location, callback) {
        const normalizedLocation = normalizeLocation(location);
        
        // Get or create group for this location
        let group = this.groups.get(normalizedLocation);
        
        if (!group) {
            // Create new group for this location
            group = {
                location: normalizedLocation,
                originalLocation: location,
                callbacks: [],
                startTime: Date.now(),
                timer: null,
                processing: false
            };
            
            this.groups.set(normalizedLocation, group);
            
            // Start timer for this group
            group.timer = setTimeout(() => {
                this.processGroup(normalizedLocation, 'timeout');
            }, this.windowMs);
        }

        // Add user to this group
        group.callbacks.push(callback);
        
        // Check if we reached max users limit
        if (group.callbacks.length >= this.maxUsers) {
            // Stop timer and process immediately
            clearTimeout(group.timer);
            this.processGroup(normalizedLocation, 'max_users');
        }
    }

    /**
     * Process a group of requests
     * Get weather data and send to all users in the group
     * @param {string} location - Normalized location name
     * @param {string} reason - Why we are processing now
     */
    async processGroup(location, reason) {
        const group = this.groups.get(location);
        
        if (!group || group.processing) {
            return; // Group already being processed or doesn't exist
        }
        
        group.processing = true;
        const userCount = group.callbacks.length;
        const waitTime = Date.now() - group.startTime;
        
        this.groups.delete(location);
        
        try {
            const [weatherApiData, weatherStackData] = await Promise.all([
                weatherApiService.getWeather(group.originalLocation),
                weatherStackService.getWeather(group.originalLocation)
            ]);
            
            const averageTemp = calculateAverageTemperature(
                weatherApiData.temperature,
                weatherStackData.temperature
            );
            
            const response = {
                location: group.originalLocation,
                temperature: averageTemp
            };
            
            logAggregation(location, userCount, waitTime, reason);
            
            const queue = getQueue();
            queue.logWeatherQuery({
                location: group.originalLocation,
                service1Temperature: weatherApiData.temperature,
                service2Temperature: weatherStackData.temperature,
                requestCount: userCount
            });
            
            // Send success 
            group.callbacks.forEach(callback => {
                try {
                    callback(null, response);
                } catch (error) {
                    console.error('Error calling user callback:', error);
                }
            });
            
        } catch (error) {
            // Log error
            logAggregation(location, userCount, waitTime, `error_${reason}`);
            
            group.callbacks.forEach(callback => {
                try {
                    callback(error, null);
                } catch (callbackError) {
                    console.error('Error calling user callback:', callbackError);
                }
            });
        }
    }

    /**
     * Get information about current groups
     * @returns {Object} Statistics about active groups
     */
    getStats() {
        const activeGroups = [];
        
        for (const [location, group] of this.groups) {
            activeGroups.push({
                location,
                userCount: group.callbacks.length,
                waitTime: Date.now() - group.startTime,
                processing: group.processing
            });
        }
        
        return {
            activeGroupCount: this.groups.size,
            activeGroups,
            windowMs: this.windowMs,
            maxUsers: this.maxUsers
        };
    }

    /**
     * Clean up when shutting down
     * Process all remaining groups immediately
     */
    async shutdown() {
        const activeLocations = Array.from(this.groups.keys());
        
        for (const location of activeLocations) {
            const group = this.groups.get(location);
            if (group && group.timer) {
                clearTimeout(group.timer);
            }
            await this.processGroup(location, 'shutdown');
        }
    }
}

// Create one instance to use everywhere
const requestAggregator = new RequestAggregator();
module.exports = { requestAggregator };
