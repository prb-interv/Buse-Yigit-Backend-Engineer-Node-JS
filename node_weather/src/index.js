const { createApp } = require('./app.js');
const { logger } = require('./lib/logger.js');

const PORT = process.env.PORT || 3000;
const HOST = process.env.HOST || '0.0.0.0';

async function startServer() {
    try {
        const app = await createApp();
        const server = app.listen(PORT, HOST, () => {
            logger.info({
                port: PORT,
                host: HOST,
                environment: process.env.NODE_ENV || 'production'
            }, 'Weather server started successfully');
            
            logger.info({
                endpoints: {
                    weather: `http://${HOST}:${PORT}/weather?q=<location>`,
                    root: `http://${HOST}:${PORT}/`,
                    health: process.env.NODE_ENV === 'development' ? 
                        `http://${HOST}:${PORT}/weather/health` : 'disabled',
                    stats: process.env.NODE_ENV === 'development' ? 
                        `http://${HOST}:${PORT}/weather/stats` : 'disabled'
                }
            }, 'Available endpoints');
        });
        server.on('error', (error) => {
            if (error.code === 'EADDRINUSE') {
                logger.error({
                    port: PORT,
                    error: error.message
                }, `Port ${PORT} is already in use`);
            } else {
                logger.error({
                    error: error.message
                }, 'Server error occurred');
            }
            process.exit(1);
        });
        
        return server;
        
    } catch (error) {
        logger.error({
            error: error.message,
            stack: error.stack
        }, 'Failed to start server');
        process.exit(1);
    }
}
process.on('uncaughtException', (error) => {
    logger.error({
        error: error.message,
        stack: error.stack
    }, 'Uncaught exception occurred');
    process.exit(1);
});
process.on('unhandledRejection', (reason, promise) => {
    logger.error({
        reason: reason?.message || reason,
        promise: promise.toString()
    }, 'Unhandled promise rejection');
    process.exit(1);
});

startServer();
