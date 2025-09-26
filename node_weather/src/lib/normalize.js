function normalizeLocation(location) {
    if (!location || typeof location !== 'string') {
        throw new Error('Location must be a text');
    }
    
    return location.trim().toLowerCase();
}


function isValidTemperature(temperature) {
    return typeof temperature === 'number' && 
           !isNaN(temperature) && 
           temperature >= -100 && 
           temperature <= 100;
}

function normalizeWeatherApiResponse(response) {
    try {
        if (!response || !response.current) {
            throw new Error('WeatherAPI response is broken');
        }

        const temperature = response.current.temp_c;
        
        if (!isValidTemperature(temperature)) {
            throw new Error('WeatherAPI gave us bad temperature');
        }

        return {
            temperature,
            service: 'weatherapi',
            location: response.location?.name || 'unknown'
        };
    } catch (error) {
        throw new Error(`Cannot read WeatherAPI data: ${error.message}`);
    }
}

function normalizeWeatherStackResponse(response) {
    try {
        if (!response || !response.current) {
            throw new Error('WeatherStack response is broken');
        }

        const temperature = response.current.temperature;
        
        if (!isValidTemperature(temperature)) {
            throw new Error('WeatherStack gave us bad temperature');
        }

        return {
            temperature,
            service: 'weatherstack',
            location: response.location?.name || 'unknown'
        };
    } catch (error) {
        throw new Error(`Cannot read WeatherStack data: ${error.message}`);
    }
}

function calculateAverageTemperature(temp1, temp2) {
    if (!isValidTemperature(temp1) || !isValidTemperature(temp2)) {
        throw new Error('Both temperatures must be good numbers');
    }
    
    return Math.round((temp1 + temp2) / 2 * 10) / 10; // Round to 1 decimal
}

module.exports = {
    normalizeLocation,
    isValidTemperature,
    normalizeWeatherApiResponse,
    normalizeWeatherStackResponse,
    calculateAverageTemperature
};
