/*
Weather API Test Script
İstanbul: 11 eşzamanlı istek
Ankara: 3 istek, 1 saniye aralıklarla

Test kapsamında İstanbul için ilk 10 isteğin anında ve eşzamanlı, diğer isteğin ise 5 saniye boyunca diğer istekleri dinledikten sonra cevap dönmesi beklenir.
Ankara için ise 3 istek 1 saniye aralıklarla beklenir. İlk isteğin 5, 2. isteğin 4, 3. isteğin 3 saniye sonra cevap dönmesi beklenir.
*/

const BASE_URL = "http://localhost:3000";
const ISTANBUL_REQUESTS = 11;
const ANKARA_REQUESTS = 3;
const ANKARA_INTERVAL = 1000; // 1 second in milliseconds

async function makeRequest(city, requestNum) {
    const startTime = new Date();
    
    try {
        const response = await fetch(`${BASE_URL}/weather?q=${city}`, {
            signal: AbortSignal.timeout(30000) // 30 second timeout
        });
        
        const endTime = new Date();
        const duration = (endTime - startTime) / 1000;
        
        if (!response.ok) {
            return {
                requestNumber: requestNum,
                city: city,
                startTime: formatTime(startTime),
                endTime: formatTime(endTime),
                duration: round(duration, 3),
                success: false,
                error: `HTTP ${response.status}`
            };
        }
        
        const weatherResp = await response.json();
        
        return {
            requestNumber: requestNum,
            city: city,
            startTime: formatTime(startTime),
            endTime: formatTime(endTime),
            duration: round(duration, 3),
            temperature: weatherResp.temperature,
            success: true
        };
        
    } catch (error) {
        const endTime = new Date();
        const duration = (endTime - startTime) / 1000;
        
        return {
            requestNumber: requestNum,
            city: city,
            startTime: formatTime(startTime),
            endTime: formatTime(endTime),
            duration: round(duration, 3),
            success: false,
            error: error.message
        };
    }
}

async function makeRequestWithDelay(city, requestNum, delay) {
    await new Promise(resolve => setTimeout(resolve, delay));
    return makeRequest(city, requestNum);
}

async function testIstanbul() {
    console.log("TEST 1: İstanbul");
    console.log("🎯 Beklenen: 10 istek HEMEN, 1 istek 5 saniye sonra");
    console.log(`⏰ Başlangıç zamanı: ${formatTime(new Date())}\n`);
    
    const t0 = new Date();
    const promises = [];
    
    for (let i = 1; i <= ISTANBUL_REQUESTS; i++) {
        console.log(`🚀 İstanbul İsteği #${i} başlatıldı`);
        promises.push(makeRequest("Istanbul", i));
    }
    
    console.log("\n⏳ İstanbul istekleri tamamlanmayı bekliyor...");
    const results = await Promise.all(promises);
    
    const totalDuration = (new Date() - t0) / 1000;
    
    console.log("\nISTANBUL RESULTS:");
    console.log(`Total duration: ${totalDuration.toFixed(3)} seconds\n`);
    
    // Sort by request number
    results.sort((a, b) => a.requestNumber - b.requestNumber);
    
    for (const r of results) {
        if (r.success) {
            console.log(`✅ Request #${r.requestNumber}: ${r.startTime} → ${r.endTime} | Duration: ${r.duration}s | Temperature: ${r.temperature.toFixed(1)}°C`);
        } else {
            console.log(`❌ Request #${r.requestNumber}: ${r.startTime} → ${r.endTime} | Duration: ${r.duration}s | Error: ${r.error}`);
        }
    }
    
    return results;
}

async function testAnkara() {
    console.log("\n" + "================================================================================\n");
    console.log("TEST 2: Ankara ");
    console.log(`Start time: ${formatTime(new Date())}\n`);
    
    const t0 = new Date();
    const promises = [];
    
    for (let i = 1; i <= ANKARA_REQUESTS; i++) {
        const delay = (i - 1) * ANKARA_INTERVAL; // 0ms, 1000ms, 2000ms
        promises.push(makeRequestWithDelay("Ankara", i, delay));
    }
    
    const results = await Promise.all(promises);
    
    const totalDuration = (new Date() - t0) / 1000;
    
    console.log("\nANKARA RESULTS:");
    console.log(`Total duration: ${totalDuration.toFixed(3)} seconds\n`);
    
    // Sort by request number
    results.sort((a, b) => a.requestNumber - b.requestNumber);
    
    for (const r of results) {
        if (r.success) {
            console.log(`✅ Request #${r.requestNumber}: ${r.startTime} → ${r.endTime} | Duration: ${r.duration}s | Temperature: ${r.temperature.toFixed(1)}°C`);
        } else {
            console.log(`❌ Request #${r.requestNumber}: ${r.startTime} → ${r.endTime} | Duration: ${r.duration}s | Error: ${r.error}`);
        }
    }
    
    return results;
}

function formatTime(date) {
    return date.toTimeString().split(' ')[0] + '.' + date.getMilliseconds().toString().padStart(3, '0');
}

function round(val, precision) {
    const factor = Math.pow(10, precision);
    return Math.round(val * factor) / factor;
}

async function main() {
    console.log(`Base URL: ${BASE_URL}\n`);
    
    await testIstanbul();
    await testAnkara();
    
    console.log("est completed ");
}

main().catch(console.error);
