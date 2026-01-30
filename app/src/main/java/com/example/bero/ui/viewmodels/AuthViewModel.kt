package com.example.bero.ui.viewmodels

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.bero.data.repository.AuthRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * Authentication states
 */
sealed class AuthState {
    object Initial : AuthState()
    object Loading : AuthState()
    object OtpSent : AuthState()
    object LoggedIn : AuthState()
    object NewUser : AuthState()
    data class Error(val message: String) : AuthState()
}

/**
 * ViewModel for authentication screens
 */
class AuthViewModel(private val authRepository: AuthRepository) : ViewModel() {
    
    private val _authState = MutableStateFlow<AuthState>(AuthState.Initial)
    val authState: StateFlow<AuthState> = _authState.asStateFlow()
    
    private val _requestId = MutableStateFlow<String?>(null)
    val requestId: StateFlow<String?> = _requestId.asStateFlow()
    
    private val _phoneNumber = MutableStateFlow("")
    val phoneNumber: StateFlow<String> = _phoneNumber.asStateFlow()
    
    private val _userType = MutableStateFlow<String?>(null)
    val userType: StateFlow<String?> = _userType.asStateFlow()
    
    init {
        // Check if already logged in
        checkLoginStatus()
    }
    
    private fun checkLoginStatus() {
        if (authRepository.isLoggedIn()) {
            _authState.value = AuthState.LoggedIn
            _userType.value = authRepository.getUserType()
        }
    }
    
    /**
     * Send OTP to phone number
     */
    fun sendOtp(phone: String) {
        viewModelScope.launch {
            _authState.value = AuthState.Loading
            _phoneNumber.value = phone
            
            authRepository.sendOtp(phone)
                .onSuccess { reqId ->
                    _requestId.value = reqId
                    _authState.value = AuthState.OtpSent
                }
                .onFailure { e ->
                    _authState.value = AuthState.Error(e.message ?: "Failed to send OTP")
                }
        }
    }
    
    /**
     * Verify OTP and login
     */
    fun verifyOtp(otp: String) {
        val phone = _phoneNumber.value
        val reqId = _requestId.value
        
        if (phone.isEmpty() || reqId == null) {
            _authState.value = AuthState.Error("Please request OTP first")
            return
        }
        
        viewModelScope.launch {
            _authState.value = AuthState.Loading
            
            authRepository.verifyOtp(phone, otp, reqId)
                .onSuccess { isNewUser ->
                    _userType.value = authRepository.getUserType()
                    _authState.value = if (isNewUser) AuthState.NewUser else AuthState.LoggedIn
                }
                .onFailure { e ->
                    _authState.value = AuthState.Error(e.message ?: "Invalid OTP")
                }
        }
    }
    
    /**
     * Resend OTP
     */
    fun resendOtp() {
        val phone = _phoneNumber.value
        if (phone.isNotEmpty()) {
            sendOtp(phone)
        }
    }
    
    /**
     * Logout
     */
    fun logout() {
        viewModelScope.launch {
            authRepository.logout()
            _authState.value = AuthState.Initial
            _requestId.value = null
            _phoneNumber.value = ""
            _userType.value = null
        }
    }
    
    /**
     * Check if user is a worker
     */
    fun isWorker(): Boolean {
        return _userType.value == "WORKER"
    }
    
    /**
     * Check if user is a client
     */
    fun isClient(): Boolean {
        return _userType.value == "CLIENT"
    }
    
    /**
     * Reset to initial state
     */
    fun reset() {
        _authState.value = AuthState.Initial
        _requestId.value = null
    }
}
