const { logger, logDatabaseOperation } = require('./logger.js');

class DatabaseQueue {
    constructor(database, options = {}) {
        this.database = database;
        this.queue = [];
        this.processing = false;
        this.batchSize = options.batchSize || 10;
        this.maxRetries = options.maxRetries || 3;
        this.retryDelay = options.retryDelay || 1000;
        this.flushInterval = options.flushInterval || 500; // ms
        
        this.startProcessing();
    }

    enqueue(operation) {
        operation.retries = 0;
        operation.timestamp = Date.now();
        
        this.queue.push(operation);
        
        logger.debug({
            queueSize: this.queue.length,
            operation: operation.type
        }, 'Operation added to database queue');

        if (this.queue.length >= this.batchSize) {
            this.processQueue();
        }
    }

    logWeatherQuery(data) {
        this.enqueue({
            type: 'insertWeatherQuery',
            data: data
        });
    }

    startProcessing() {
        setInterval(() => {
            if (this.queue.length > 0 && !this.processing) {
                this.processQueue();
            }
        }, this.flushInterval);
    }

    async processQueue() {
        if (this.processing || this.queue.length === 0) {
            return;
        }

        this.processing = true;
        const batch = this.queue.splice(0, this.batchSize);
        
        logger.debug({
            batchSize: batch.length,
            remainingInQueue: this.queue.length
        }, 'Processing database queue batch');

        for (const operation of batch) {
            try {
                await this.executeOperation(operation);
                
                logDatabaseOperation(
                    operation.type,
                    { queueTime: Date.now() - operation.timestamp },
                    true
                );
                
            } catch (error) {
                await this.handleFailedOperation(operation, error);
            }
        }
        logger.debug({
            processedCount: batch.length,
            remainingInQueue: this.queue.length
        }, 'Database queue batch processed and cleared from memory');

        this.processing = false;
    }
    async executeOperation(operation) {
        switch (operation.type) {
            case 'insertWeatherQuery':
                return this.database.insertWeatherQuery(operation.data);
            
            default:
                throw new Error(`Unknown operation type: ${operation.type}`);
        }
    }
    async handleFailedOperation(operation, error) {
        operation.retries++;
        
        logDatabaseOperation(
            operation.type,
            { 
                retries: operation.retries,
                maxRetries: this.maxRetries,
                error: error.message 
            },
            false,
            error.message
        );

        if (operation.retries < this.maxRetries) {
            setTimeout(() => {
                this.queue.unshift(operation);
            }, this.retryDelay * operation.retries);
            
        } else {
            logger.error({
                operation: operation.type,
                data: operation.data,
                error: error.message,
                retries: operation.retries
            }, 'Database operation failed permanently after max retries');
        }
    }
    async flush() {
        logger.info('Flushing database queue...');
        
        while (this.queue.length > 0) {
            await this.processQueue();
            
            // Small delay to prevent busy waiting
            if (this.queue.length > 0) {
                await new Promise(resolve => setTimeout(resolve, 10));
            }
        }
        
        logger.info('Database queue flushed');
    }
    getStats() {
        return {
            queueSize: this.queue.length,
            processing: this.processing,
            batchSize: this.batchSize
        };
    }
}
let queueInstance = null;

function initializeQueue(database, options = {}) {
    if (!queueInstance) {
        queueInstance = new DatabaseQueue(database, options);
    }
    return queueInstance;
}

function getQueue() {
    if (!queueInstance) {
        throw new Error('Database queue not initialized. Call initializeQueue() first.');
    }
    return queueInstance;
}

module.exports = { DatabaseQueue, initializeQueue, getQueue };
