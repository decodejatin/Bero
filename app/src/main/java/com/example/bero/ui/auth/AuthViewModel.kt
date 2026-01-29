package com.example.bero.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.bero.auth.*
import com.bero.domain.models.User
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * ViewModel for authentication screens
 */
class AuthViewModel : ViewModel() {
    
    private val sessionManager: SessionManager = InMemorySessionManager()
    private val authRepository: AuthRepository = AuthRepositoryImpl()
    private val authUseCase = AuthUseCase(authRepository, sessionManager)
    
    val authState: StateFlow<AuthState> = authUseCase.authState
    
    private val _uiState = MutableStateFlow(AuthUiState())
    val uiState: StateFlow<AuthUiState> = _uiState.asStateFlow()
    
    init {
        viewModelScope.launch {
            authUseCase.initialize()
        }
    }
    
    /**
     * Request OTP for phone number
     */
    fun requestOtp(phoneNumber: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                isLoading = true,
                error = null,
                phoneNumber = phoneNumber
            )
            
            authUseCase.requestOtp(phoneNumber).fold(
                onSuccess = {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        otpSent = true
                    )
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = getErrorMessage(error)
                    )
                }
            )
        }
    }
    
    /**
     * Verify OTP
     */
    fun verifyOtp(otp: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                isLoading = true,
                error = null
            )
            
            authUseCase.verifyOtp(otp).fold(
                onSuccess = { user ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        otpVerified = true
                    )
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = getErrorMessage(error)
                    )
                }
            )
        }
    }
    
    /**
     * Resend OTP
     */
    fun resendOtp() {
        _uiState.value.phoneNumber?.let { phone ->
            requestOtp(phone)
        }
    }
    
    /**
     * Go back from OTP screen
     */
    fun goBackToLogin() {
        _uiState.value = AuthUiState()
    }
    
    /**
     * Authenticate with Truecaller
     */
    fun authenticateWithTruecaller(token: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                isLoading = true,
                error = null
            )
            
            authUseCase.authenticateWithTruecaller(token).fold(
                onSuccess = { user ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        otpVerified = true
                    )
                },
                onFailure = { error ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = getErrorMessage(error)
                    )
                }
            )
        }
    }
    
    /**
     * Logout
     */
    fun logout() {
        viewModelScope.launch {
            authUseCase.logout()
            _uiState.value = AuthUiState()
        }
    }
    
    private fun getErrorMessage(error: Throwable): String {
        return when (error) {
            is AuthError.InvalidPhoneNumber -> "Please enter a valid 10-digit phone number"
            is AuthError.InvalidOtp -> "Invalid OTP. Please try again."
            is AuthError.OtpExpired -> "OTP has expired. Please request a new one."
            is AuthError.NetworkError -> "Network error. Please check your connection."
            is AuthError.UserBlocked -> "Your account has been blocked. Contact support."
            is AuthError.Unknown -> error.message ?: "An error occurred"
            else -> "An unexpected error occurred"
        }
    }
}

/**
 * UI state for auth screens
 */
data class AuthUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val phoneNumber: String? = null,
    val otpSent: Boolean = false,
    val otpVerified: Boolean = false
)
