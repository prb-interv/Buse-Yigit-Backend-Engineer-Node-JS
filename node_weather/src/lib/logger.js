const pino = require('pino');

function createLogger() {
    const logLevel = process.env.LOG_LEVEL || 'info';
    
    const config = {
        level: logLevel,
        transport: {
            target: 'pino-pretty',
            options: {
                colorize: true,
                translateTime: 'HH:MM:ss',
                ignore: 'pid,hostname,level,time',
                messageFormat: '{msg}'
            }
        }
    };

    return pino(config);
}

const logger = createLogger();



function createChildLogger(context) {
    return logger.child(context);
}

function logWeatherRequest(service, location, responseTime, success, error = null) {
    const logData = {
        service,
        location,
        responseTime,
        success
    };

    if (error) {
        logData.error = error;
        logger.error(logData, `Weather API request failed: ${service}`);
    } else {
        logger.info(logData, `Weather API request completed: ${service}`);
    }
}

function logAggregation(location, userCount, waitTime, triggerReason) {
    logger.info({
        location,
        userCount,
        waitTime,
        triggerReason,
        aggregation: true
    }, `Request aggregation completed: ${triggerReason}`);
}

function logDatabaseOperation(operation, metadata, success, error = null) {
    const logData = {
        operation,
        ...metadata,
        success
    };

    if (error) {
        logData.error = error;
        logger.error(logData, `Database operation failed: ${operation}`);
    } else {
        logger.debug(logData, `Database operation completed: ${operation}`);
    }
}

module.exports = { logger, createChildLogger, logWeatherRequest, logAggregation, logDatabaseOperation };
