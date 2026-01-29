package com.bero.auth

import com.bero.domain.models.User
import com.bero.domain.models.KycStatus
import com.bero.domain.models.UserType

/**
 * Repository for authentication operations.
 * Handles OTP-based phone authentication.
 */
interface AuthRepository {
    /**
     * Request OTP for phone number
     * @param phoneNumber The phone number with country code (e.g., +91XXXXXXXXXX)
     * @return OtpRequest with requestId for verification
     */
    suspend fun sendOtp(phoneNumber: String): Result<OtpRequest>
    
    /**
     * Verify OTP and authenticate user
     * @param phoneNumber The phone number
     * @param otp The 6-digit OTP
     * @param requestId The requestId from sendOtp
     * @return User on successful verification
     */
    suspend fun verifyOtp(phoneNumber: String, otp: String, requestId: String): Result<User>
    
    /**
     * Authenticate with Truecaller token (Android only)
     * @param truecallerToken Token from Truecaller SDK
     * @return User on successful verification
     */
    suspend fun authenticateWithTruecaller(truecallerToken: String): Result<User>
    
    /**
     * Refresh authentication token
     * @param refreshToken The refresh token
     * @return New session with fresh tokens
     */
    suspend fun refreshToken(refreshToken: String): Result<Session>
    
    /**
     * Logout and invalidate session
     */
    suspend fun logout(): Result<Unit>
    
    /**
     * Get current user profile from server
     */
    suspend fun getCurrentUser(): Result<User>
}

/**
 * Default implementation using Ktor HTTP client
 */
class AuthRepositoryImpl(
    private val baseUrl: String = "https://api.bero.app"
) : AuthRepository {
    
    override suspend fun sendOtp(phoneNumber: String): Result<OtpRequest> {
        // Validate phone number format
        if (!isValidIndianPhoneNumber(phoneNumber)) {
            return Result.failure(AuthError.InvalidPhoneNumber)
        }
        
        // TODO: Implement actual API call
        // For now, return mock data for development
        return Result.success(
            OtpRequest(
                phoneNumber = phoneNumber,
                requestId = "mock-request-${currentTimeMillis()}",
                expiresInSeconds = 120
            )
        )
    }
    
    override suspend fun verifyOtp(phoneNumber: String, otp: String, requestId: String): Result<User> {
        // Validate OTP format
        if (otp.length != 6 || !otp.all { it.isDigit() }) {
            return Result.failure(AuthError.InvalidOtp)
        }
        
        // TODO: Implement actual API call
        // For development, accept any 6-digit OTP
        return Result.success(
            User(
                id = "user-${phoneNumber.takeLast(4)}",
                phoneNumber = phoneNumber,
                fullName = null,
                aadhaarKycStatus = KycStatus.NONE,
                userType = UserType.WORKER
            )
        )
    }
    
    override suspend fun authenticateWithTruecaller(truecallerToken: String): Result<User> {
        // TODO: Implement Truecaller verification via backend
        return Result.failure(AuthError.Unknown("Truecaller auth not implemented"))
    }
    
    override suspend fun refreshToken(refreshToken: String): Result<Session> {
        // TODO: Implement token refresh
        return Result.failure(AuthError.Unknown("Token refresh not implemented"))
    }
    
    override suspend fun logout(): Result<Unit> {
        // TODO: Implement logout API call
        return Result.success(Unit)
    }
    
    override suspend fun getCurrentUser(): Result<User> {
        // TODO: Implement get current user
        return Result.failure(AuthError.Unknown("Not implemented"))
    }
    
    private fun isValidIndianPhoneNumber(phone: String): Boolean {
        // Remove spaces and dashes
        val cleaned = phone.replace(" ", "").replace("-", "")
        
        // Check for valid Indian phone formats
        return when {
            // +91XXXXXXXXXX format
            cleaned.startsWith("+91") && cleaned.length == 13 -> true
            // 91XXXXXXXXXX format
            cleaned.startsWith("91") && cleaned.length == 12 -> true
            // XXXXXXXXXX format (10 digits)
            cleaned.length == 10 && cleaned.all { it.isDigit() } -> true
            else -> false
        }
    }
}

// Expect/actual for platform-specific time
expect fun currentTimeMillis(): Long

