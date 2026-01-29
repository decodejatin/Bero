package com.bero.auth

import com.bero.domain.models.User
import com.bero.domain.models.KycStatus
import com.bero.domain.models.UserType
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * Use case for authentication flow management.
 * Coordinates between AuthRepository and SessionManager.
 */
class AuthUseCase(
    private val authRepository: AuthRepository,
    private val sessionManager: SessionManager
) {
    private val _authState = MutableStateFlow<AuthState>(AuthState.NotAuthenticated)
    val authState: StateFlow<AuthState> = _authState.asStateFlow()
    
    private var currentOtpRequest: OtpRequest? = null
    
    /**
     * Initialize auth state from stored session
     */
    suspend fun initialize() {
        val session = sessionManager.getSession()
        if (session != null) {
            _authState.value = determineAuthState(session)
        } else {
            _authState.value = AuthState.NotAuthenticated
        }
    }
    
    /**
     * Request OTP for phone number
     */
    suspend fun requestOtp(phoneNumber: String): Result<Unit> {
        _authState.value = AuthState.Authenticating(phoneNumber)
        
        return authRepository.sendOtp(phoneNumber).fold(
            onSuccess = { otpRequest ->
                currentOtpRequest = otpRequest
                Result.success(Unit)
            },
            onFailure = { error ->
                _authState.value = AuthState.NotAuthenticated
                Result.failure(error)
            }
        )
    }
    
    /**
     * Verify OTP and complete authentication
     */
    suspend fun verifyOtp(otp: String): Result<User> {
        val otpRequest = currentOtpRequest ?: return Result.failure(AuthError.OtpExpired)
        
        return authRepository.verifyOtp(
            phoneNumber = otpRequest.phoneNumber,
            otp = otp,
            requestId = otpRequest.requestId
        ).fold(
            onSuccess = { user ->
                val token = "mock-token-${currentTimeMillis()}" // TODO: Get from server
                val session = Session(
                    userId = user.id,
                    phoneNumber = user.phoneNumber,
                    token = token,
                    userType = user.userType,
                    kycStatus = user.aadhaarKycStatus
                )
                sessionManager.saveSession(session)
                _authState.value = determineAuthState(session)
                currentOtpRequest = null
                Result.success(user)
            },
            onFailure = { error ->
                Result.failure(error)
            }
        )
    }
    
    /**
     * Authenticate with Truecaller
     */
    suspend fun authenticateWithTruecaller(token: String): Result<User> {
        return authRepository.authenticateWithTruecaller(token).fold(
            onSuccess = { user ->
                val authToken = "mock-token-${currentTimeMillis()}"
                val session = Session(
                    userId = user.id,
                    phoneNumber = user.phoneNumber,
                    token = authToken,
                    userType = user.userType,
                    kycStatus = user.aadhaarKycStatus
                )
                sessionManager.saveSession(session)
                _authState.value = determineAuthState(session)
                Result.success(user)
            },
            onFailure = { error ->
                Result.failure(error)
            }
        )
    }
    
    /**
     * Logout current user
     */
    suspend fun logout(): Result<Unit> {
        authRepository.logout()
        sessionManager.clearSession()
        currentOtpRequest = null
        _authState.value = AuthState.NotAuthenticated
        return Result.success(Unit)
    }
    
    /**
     * Update KYC status after verification
     */
    suspend fun updateKycStatus(status: KycStatus) {
        sessionManager.updateKycStatus(status)
        sessionManager.getSession()?.let { session ->
            _authState.value = determineAuthState(session.copy(kycStatus = status))
        }
    }
    
    /**
     * Update video bio status
     */
    suspend fun updateVideoBioStatus(hasVideoBio: Boolean) {
        sessionManager.updateVideoBioStatus(hasVideoBio)
    }
    
    /**
     * Determine auth state based on session
     */
    private suspend fun determineAuthState(session: Session): AuthState {
        // Get current user from server if possible
        val user = authRepository.getCurrentUser().getOrNull() ?: User(
            id = session.userId,
            phoneNumber = session.phoneNumber,
            userType = session.userType,
            aadhaarKycStatus = session.kycStatus
        )
        
        return when {
            session.kycStatus != KycStatus.VERIFIED -> AuthState.RequiresKyc(user, session.token)
            !session.hasVideoBio -> AuthState.RequiresVideoBio(user, session.token)
            else -> AuthState.Authenticated(user, session.token)
        }
    }
}
