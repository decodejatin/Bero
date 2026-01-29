package com.bero.auth

import com.bero.domain.models.User
import com.bero.domain.models.KycStatus
import com.bero.domain.models.UserType
import kotlinx.serialization.Serializable

/**
 * Represents the authentication state of the application.
 * This is the single source of truth for auth status.
 */
sealed class AuthState {
    /**
     * User is not authenticated - show login screen
     */
    data object NotAuthenticated : AuthState()
    
    /**
     * Authentication is in progress (OTP sent, waiting verification)
     */
    data class Authenticating(val phoneNumber: String) : AuthState()
    
    /**
     * User is authenticated and KYC is complete
     */
    data class Authenticated(val user: User, val token: String) : AuthState()
    
    /**
     * User is authenticated but needs to complete KYC
     */
    data class RequiresKyc(val user: User, val token: String) : AuthState()
    
    /**
     * User is authenticated but needs to record video bio
     */
    data class RequiresVideoBio(val user: User, val token: String) : AuthState()
}

/**
 * Session data stored locally
 */
@Serializable
data class Session(
    val userId: String,
    val phoneNumber: String,
    val token: String,
    val userType: UserType,
    val kycStatus: KycStatus = KycStatus.NONE,
    val hasVideoBio: Boolean = false
)

/**
 * OTP request result
 */
@Serializable
data class OtpRequest(
    val phoneNumber: String,
    val requestId: String,
    val expiresInSeconds: Int = 120
)

/**
 * Auth error types
 */
sealed class AuthError : Exception() {
    data object InvalidPhoneNumber : AuthError()
    data object InvalidOtp : AuthError()
    data object OtpExpired : AuthError()
    data object NetworkError : AuthError()
    data object UserBlocked : AuthError()
    data class Unknown(override val message: String) : AuthError()
}

