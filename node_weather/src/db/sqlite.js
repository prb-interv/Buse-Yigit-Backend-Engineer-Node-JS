const Database = require('better-sqlite3');

class DatabaseManager {
    constructor(dbPath = './weather.sqlite') {
        this.dbPath = dbPath;
        this.db = null;
    }

    async initialize() {
        try {
            this.db = new Database(this.dbPath);
            
            this.db.pragma('journal_mode = DELETE');            
            this.createTables();
            
            console.log('Database initialized successfully');
        } catch (error) {
            console.error('Database initialization failed:', error);
            throw error;
        }
    }

 
    createTables() {
        this.db.exec(`
            CREATE TABLE IF NOT EXISTS weather_queries (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                location TEXT NOT NULL,
                service_1_temperature REAL NOT NULL,
                service_2_temperature REAL NOT NULL,
                request_count INTEGER NOT NULL DEFAULT 1,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        `);

        this.db.exec(`CREATE INDEX IF NOT EXISTS idx_weather_queries_location ON weather_queries(location)`);
        this.db.exec(`CREATE INDEX IF NOT EXISTS idx_weather_queries_created_at ON weather_queries(created_at)`);
    }

    insertWeatherQuery(data) {
        const stmt = this.db.prepare(`
            INSERT INTO weather_queries (location, service_1_temperature, service_2_temperature, request_count)
            VALUES (?, ?, ?, ?)
        `);
        
        return stmt.run(
            data.location,
            data.service1Temperature,
            data.service2Temperature,
            data.requestCount
        );
    }

    getRecentWeatherQueries(location, limit = 10) {
        const stmt = this.db.prepare(`
            SELECT * FROM weather_queries 
            WHERE location = ? 
            ORDER BY created_at DESC 
            LIMIT ?
        `);
        
        return stmt.all(location, limit);
    }
    getStats() {
        const totalQueries = this.db.prepare('SELECT COUNT(*) as count FROM weather_queries').get();
        const uniqueLocations = this.db.prepare('SELECT COUNT(DISTINCT location) as count FROM weather_queries').get();
        
        return {
            totalQueries: totalQueries.count,
            uniqueLocations: uniqueLocations.count
        };
    }
    close() {
        if (this.db) {
            this.db.close();
            this.db = null;
        }
    }
}
let dbInstance = null;

function getDatabase(dbPath) {
    if (!dbInstance) {
        dbInstance = new DatabaseManager(dbPath);
    }
    return dbInstance;
}

module.exports = { getDatabase, DatabaseManager };
