const express = require('express');
const dotenv = require('dotenv');
const { logger } = require('./lib/logger.js');
const { getDatabase } = require('./db/sqlite.js');
const { initializeQueue } = require('./lib/queue.js');
const weatherRoutes = require('./routes/weather.js');


dotenv.config();
const app = express();


async function initializeServices() {
    try {
        
        const dbPath = process.env.DB_PATH || './weather.sqlite';
        const database = getDatabase(dbPath);
        await database.initialize();
        
        
        initializeQueue(database);
        
        logger.info('All services initialized successfully');
        return true;
        
    } catch (error) {
        logger.error({
            error: error.message
        }, 'Failed to initialize services');
        throw error;
    }
}

/**
 
 */
function configureMiddleware() {
    
    app.use(express.json({ limit: '10mb' }));   
    app.use(express.urlencoded({ extended: true }));    
    app.set('trust proxy', true);
        
    app.use((req, res, next) => {
        res.setHeader('X-Content-Type-Options', 'nosniff');
        res.setHeader('X-Frame-Options', 'DENY');
        res.setHeader('X-XSS-Protection', '1; mode=block');
        next();
    });
}


function configureRoutes() {

    app.use('/weather', weatherRoutes);    
    app.get('/', (req, res) => {
        res.json({
            name: 'Weather Query Application',
            version: '1.0.0',
            endpoints: {
                weather: '/weather?q=<location>',
                health: process.env.NODE_ENV === 'development' ? '/weather/health' : 'disabled',
                stats: process.env.NODE_ENV === 'development' ? '/weather/stats' : 'disabled'
            }
        });
    });

    app.use((req, res) => {
        res.status(404).json({
            error: 'Not found',
            message: `Endpoint ${req.method} ${req.path} does not exist`
        });
    });
}

function configureErrorHandling() {

    app.use((error, req, res, next) => {
        logger.error({
            error: error.message,
            stack: error.stack,
            method: req.method,
            url: req.url
        }, 'Unhandled error in Express app');
        
        res.status(500).json({
            error: 'Internal server error',
            message: 'Something unexpected happened'
        });
    });
}

function setupGracefulShutdown() {
    const signals = ['SIGTERM', 'SIGINT'];
    
    signals.forEach(signal => {
        process.on(signal, async () => {
            logger.info(`Received ${signal}, starting graceful shutdown...`);
            
            try {
                const { requestAggregator } = await import('./core/aggregator.js');
                const { getQueue } = await import('./lib/queue.js');
                
                await requestAggregator.shutdown();
                const queue = getQueue();
                await queue.flush();
                const { getDatabase } = await import('./db/sqlite.js');
                const database = getDatabase();
                database.close();
                
                logger.info('Graceful shutdown completed');
                process.exit(0);
                
            } catch (error) {
                logger.error({
                    error: error.message
                }, 'Error during graceful shutdown');
                process.exit(1);
            }
        });
    });
}

async function createApp() {
    try {
        await initializeServices();
        configureMiddleware();
        configureRoutes();
        configureErrorHandling();
        setupGracefulShutdown();
        
        logger.info('Express application configured successfully');
        return app;
        
    } catch (error) {
        logger.error({
            error: error.message
        }, 'Failed to create Express application');
        throw error;
    }
}

module.exports = { createApp };
