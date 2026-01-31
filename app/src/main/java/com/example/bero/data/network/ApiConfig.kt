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
        
        // Profile
        const val PROFILE = "/profile"
        const val SET_USER_TYPE = "/profile/user-type"
        
        // Workers (future)
        const val WORKERS = "/workers"
        fun workerProfile(id: String) = "/workers/$id"
        
        // Earnings (future)
        const val EARNINGS = "/earnings"
        const val EARNINGS_SUMMARY = "/earnings/summary"
    }
    
    // HTTP Timeouts
    const val CONNECT_TIMEOUT_SECONDS = 30L
    const val READ_TIMEOUT_SECONDS = 30L
    const val WRITE_TIMEOUT_SECONDS = 30L
}
