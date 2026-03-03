package com.example.bero.data.network

/**
 * API Configuration for connecting to the Bero Backend
 */
object ApiConfig {
    // Development - use your local IP when testing on physical device
    // For emulator, use 10.0.2.2. For physical device on hotspot, use 192.168.137.1
    // const val DEV_BASE_URL = "http://10.0.2.2:8080/api/v1"
    const val DEV_BASE_URL = "http://192.168.137.1:8080/api/v1"
    
    // Production
    const val PROD_BASE_URL = "https://api.bero.app/api/v1"
    
    // Current environment
    var isProduction = false
    
    val baseUrl: String
        get() = if (isProduction) PROD_BASE_URL else DEV_BASE_URL
    
    // Endpoints
    object Endpoints {
        // Auth
        const val SEND_OTP = "/auth/send-otp"
        const val VERIFY_OTP = "/auth/verify-otp"
        const val REFRESH_TOKEN = "/auth/refresh"
        const val LOGOUT = "/auth/logout"
        const val ME = "/auth/me"
        
        // Jobs
        const val JOBS = "/jobs"
        const val MY_JOBS = "/jobs/my"
        fun jobDetails(id: String) = "/jobs/$id"
        fun acceptJob(id: String) = "/jobs/$id/accept"
        fun startJob(id: String) = "/jobs/$id/start"
        fun completeJob(id: String) = "/jobs/$id/complete"
        fun cancelJob(id: String) = "/jobs/$id/cancel"
        fun confirmJob(id: String) = "/jobs/$id/confirm"
        
        // Profile
        const val PROFILE = "/profile"
        const val SET_USER_TYPE = "/profile/user-type"
        
        // Workers (future)
        const val WORKERS = "/workers"
        fun workerProfile(id: String) = "/profile/$id"
        
        // Chat
        const val CONVERSATIONS = "/chat/conversations"
        fun conversationMessages(id: String) = "/chat/conversations/$id/messages"
        fun markConversationRead(id: String) = "/chat/conversations/$id/read"
        const val UNREAD_COUNT = "/chat/unread"
        
        // Earnings (future)
        const val EARNINGS = "/earnings"
        const val EARNINGS_SUMMARY = "/earnings/summary"
        
        // Stats
        const val PROFILE_STATS = "/profile/stats"
        
        // Addresses
        const val ADDRESSES = "/addresses"
        fun addressById(id: String) = "/addresses/$id"
        
        // Ratings
        fun submitRating(jobId: String) = "/jobs/$jobId/rate"
        fun jobRatings(jobId: String) = "/jobs/$jobId/ratings"

        // Location & Geospatial
        const val WORKER_LOCATION = "/workers/location"
        const val NEARBY_WORKERS = "/workers/nearby"
        const val WORKER_AVAILABILITY = "/workers/availability"
    }
    
    // WebSocket URL (replace http with ws)
    val wsBaseUrl: String
        get() = baseUrl.replace("http://", "ws://").replace("https://", "wss://")
                .replace("/api/v1", "/api/v1/chat/ws")
    
    // HTTP Timeouts
    const val CONNECT_TIMEOUT_SECONDS = 30L
    const val READ_TIMEOUT_SECONDS = 30L
    const val WRITE_TIMEOUT_SECONDS = 30L
}
