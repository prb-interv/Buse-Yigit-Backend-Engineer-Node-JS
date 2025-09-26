const { Router } = require('express');
const { requestAggregator } = require('../core/aggregator.js');
const { logger } = require('../lib/logger.js');

const router = Router();

router.get('/', async (req, res) => {
    const startTime = Date.now();
    
    try {
        // Check if location exists
        const { q: location } = req.query;
        
        if (!location) {
            return res.status(400).json({
                error: 'Missing location parameter',
                message: 'Please provide location using ?q=<location>'
            });
        }

        if (typeof location !== 'string' || location.trim().length === 0) {
            return res.status(400).json({
                error: 'Invalid location parameter',
                message: 'Location must be a non-empty text'
            });
        }

        const result = await new Promise((resolve, reject) => {
            requestAggregator.addRequest(location, (error, data) => {
                if (error) {
                    reject(error);
                } else {
                    resolve(data);
                }
            });
        });

        const responseTime = Date.now() - startTime;

        logger.info({
            location,
            temperature: result.temperature,
            responseTime
        }, 'Weather request completed successfully');

        res.status(200).json(result);

    } catch (error) {
        const responseTime = Date.now() - startTime;
        
        logger.error({
            location: req.query.q,
            error: error.message,
            responseTime
        }, 'Weather request failed');

        res.status(500).json({
            error: 'Weather service error',
            message: error.message
        });
    }
});

//available in development 
if (process.env.NODE_ENV === 'development') {

    router.get('/health', async (req, res) => {
        try {
            const stats = requestAggregator.getStats();            
            res.status(200).json({
                status: 'healthy',
                timestamp: new Date().toISOString(),
                aggregator: stats,
                services: {
                    weatherapi: 'available',
                    weatherstack: 'available'
                }
            });

        } catch (error) {
            logger.error({
                error: error.message
            }, 'Health check failed');

            res.status(500).json({
                status: 'unhealthy',
                timestamp: new Date().toISOString(),
                error: error.message
            });
        }
    });

    router.get('/stats', (req, res) => {
        try {
            const stats = requestAggregator.getStats();
            
            res.status(200).json({
                timestamp: new Date().toISOString(),
                mode: 'development',
                ...stats
            });

        } catch (error) {
            logger.error({
                error: error.message
            }, 'Stats request failed');

            res.status(500).json({
                error: 'Could not get statistics',
                message: error.message
            });
        }
    });
}

module.exports = router;
