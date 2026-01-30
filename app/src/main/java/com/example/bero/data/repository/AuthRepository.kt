package com.example.bero.data.repository

import com.example.bero.data.network.AuthResponse
import com.example.bero.data.network.BeroApiClient
import com.example.bero.data.network.TokenManager
import com.example.bero.data.network.UserDto

/**
 * Repository for authentication operations
 */
class AuthRepository(
    private val apiClient: BeroApiClient,
    private val tokenManager: TokenManager
) {
    
    /**
     * Send OTP to phone number
     * @return requestId for verification
     */
    suspend fun sendOtp(phoneNumber: String): Result<String> {
        return apiClient.sendOtp(phoneNumber).map { it.request_id }
    }
    
    /**
     * Verify OTP and login
     * @return true if user is new, false if existing
     */
    suspend fun verifyOtp(phoneNumber: String, otp: String, requestId: String): Result<Boolean> {
        return apiClient.verifyOtp(phoneNumber, otp, requestId).map { it.is_new }
    }
    
    /**
     * Refresh access token
     */
    suspend fun refreshToken(): Result<Unit> {
        return apiClient.refreshToken().map { }
    }
    
    /**
     * Logout and clear session
     */
    suspend fun logout(): Result<Unit> {
        return apiClient.logout()
    }
    
    /**
     * Get current user from server
     */
    suspend fun getCurrentUser(): Result<UserDto> {
        return apiClient.getCurrentUser()
    }
    
    /**
     * Check if user is logged in
     */
    fun isLoggedIn(): Boolean {
        return tokenManager.isLoggedIn()
    }
    
    /**
     * Get user type from local storage
     */
    fun getUserType(): String? {
        return tokenManager.getUserType()
    }
    
    /**
     * Get user ID from local storage
     */
    fun getUserId(): String? {
        return tokenManager.getUserId()
    }
    
    /**
     * Get phone number from local storage
     */
    fun getPhoneNumber(): String? {
        return tokenManager.getPhoneNumber()
    }
}
