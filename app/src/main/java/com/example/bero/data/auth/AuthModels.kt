package com.example.bero.data.auth

import com.example.bero.data.models.User
import com.example.bero.data.models.KycStatus
import com.example.bero.data.models.UserType

/**
 * Represents the authentication state of the application.
 */
sealed class AuthState {
    data object NotAuthenticated : AuthState()
    data class Authenticating(val phoneNumber: String) : AuthState()
    data class Authenticated(val user: User, val token: String) : AuthState()
    data class RequiresRoleSelection(val user: User, val token: String) : AuthState()
    data class RequiresProfileCreation(val user: User, val token: String) : AuthState()
    data class RequiresKyc(val user: User, val token: String) : AuthState()
    data class RequiresVideoBio(val user: User, val token: String) : AuthState()
}

/**
 * Session data stored locally
 */
data class Session(
    val userId: String,
    val phoneNumber: String,
    val token: String,
    val userType: UserType,
    val kycStatus: KycStatus = KycStatus.NONE,
    val hasVideoBio: Boolean = false,
    val isProfileComplete: Boolean = false
)

/**
 * OTP request result
 */
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

/**
 * Repository for authentication operations.
 */
interface AuthRepository {
    suspend fun sendOtp(phoneNumber: String): Result<OtpRequest>
    suspend fun verifyOtp(phoneNumber: String, otp: String, requestId: String): Result<User>
    suspend fun authenticateWithTruecaller(truecallerToken: String): Result<User>
    suspend fun refreshToken(refreshToken: String): Result<Session>
    suspend fun logout(): Result<Unit>
    suspend fun getCurrentUser(): Result<User>
    suspend fun setUserType(userType: UserType): Result<Unit>
}

/**
 * Default implementation
 */
/**
 * Default implementation using Real API via AppContainer
 */
class AuthRepositoryImpl : AuthRepository {
    
    // Use the real API client from AppContainer
    private val apiClient by lazy { com.example.bero.di.AppContainer.instance.apiClient }
    private val tokenManager by lazy { com.example.bero.di.AppContainer.instance.tokenManager }
    
    override suspend fun sendOtp(phoneNumber: String): Result<OtpRequest> {
        // Validation handled by backend or basic check here
        if (!isValidIndianPhoneNumber(phoneNumber)) {
            return Result.failure(AuthError.InvalidPhoneNumber)
        }
        
        return apiClient.sendOtp(phoneNumber).map { response ->
            OtpRequest(
                phoneNumber = phoneNumber,
                requestId = response.request_id,
                expiresInSeconds = response.expires_in_seconds
            )
        }
    }
    
    override suspend fun verifyOtp(phoneNumber: String, otp: String, requestId: String): Result<User> {
        return apiClient.verifyOtp(phoneNumber, otp, requestId).map { authResponse ->
            // Map UserDto to Domain User
            mapUserDtoToUser(authResponse.user)
        }
    }
    
    override suspend fun authenticateWithTruecaller(truecallerToken: String): Result<User> {
        return Result.failure(AuthError.Unknown("Truecaller auth not implemented"))
    }
    
    override suspend fun refreshToken(refreshToken: String): Result<Session> {
         return apiClient.refreshToken().map { tokens ->
             // We need to fetch user to create full session, likely handled by UseCase or just return minimal session
             // For now, failure as auto-refresh happens in interceptor usually
             throw NotImplementedError("Refresh flow handled by TokenManager")
         }
    }
    
    override suspend fun logout(): Result<Unit> {
        return apiClient.logout()
    }
    
    override suspend fun getCurrentUser(): Result<User> {
        return apiClient.getCurrentUser().map { userDto ->
            mapUserDtoToUser(userDto)
        }
    }

    override suspend fun setUserType(userType: UserType): Result<Unit> {
        return apiClient.setUserType(userType.name).map { 
            // Success response map to Unit
        }
    }
    
    private fun mapUserDtoToUser(dto: com.example.bero.data.network.UserDto): User {
        return User(
            id = dto.id,
            phoneNumber = dto.phone_number,
            fullName = dto.full_name,
            aadhaarKycStatus = try {
                KycStatus.valueOf(dto.aadhaar_kyc_status)
            } catch (e: Exception) { KycStatus.NONE },
            userType = try {
                UserType.valueOf(dto.user_type)
            } catch (e: Exception) { UserType.NONE },
            createdAt = 0 // Timestamp parsing if needed
        )
    }

    private fun isValidIndianPhoneNumber(phone: String): Boolean {
        val cleaned = phone.replace(" ", "").replace("-", "")
        return when {
            cleaned.startsWith("+91") && cleaned.length == 13 -> true
            cleaned.startsWith("91") && cleaned.length == 12 -> true
            cleaned.length == 10 && cleaned.all { it.isDigit() } -> true
            else -> false
        }
    }
}

/**
 * Interface for managing user session persistence.
 */
interface SessionManager {
    suspend fun saveSession(session: Session)
    suspend fun getSession(): Session?
    suspend fun clearSession()
    suspend fun updateKycStatus(status: KycStatus)
    suspend fun updateVideoBioStatus(hasVideoBio: Boolean)
    suspend fun isLoggedIn(): Boolean = getSession() != null
}

/**
 * In-memory session manager for testing and development
 * Note: Now synchronized with TokenManager via AuthRepositoryImpl flow
 */
class InMemorySessionManager : SessionManager {
    private var currentSession: Session? = null
    
    override suspend fun saveSession(session: Session) {
        currentSession = session
    }
    
    override suspend fun getSession(): Session? = currentSession
    
    override suspend fun clearSession() {
        currentSession = null
    }
    
    override suspend fun updateKycStatus(status: KycStatus) {
        currentSession = currentSession?.copy(kycStatus = status)
    }
    
    override suspend fun updateVideoBioStatus(hasVideoBio: Boolean) {
        currentSession = currentSession?.copy(hasVideoBio = hasVideoBio)
    }
}

/**
 * Use case for authentication flow management.
 */
class AuthUseCase(
    private val authRepository: AuthRepository,
    private val sessionManager: SessionManager
) {
    private val _authState = kotlinx.coroutines.flow.MutableStateFlow<AuthState>(AuthState.NotAuthenticated)
    val authState: kotlinx.coroutines.flow.StateFlow<AuthState> = _authState
    
    private var currentOtpRequest: OtpRequest? = null
    
    suspend fun initialize() {
        val session = sessionManager.getSession()
        if (session != null) {
            _authState.value = determineAuthState(session)
        } else {
            _authState.value = AuthState.NotAuthenticated
        }
    }
    
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
    
    suspend fun verifyOtp(otp: String): Result<User> {
        val otpRequest = currentOtpRequest ?: return Result.failure(AuthError.OtpExpired)
        
        return authRepository.verifyOtp(
            phoneNumber = otpRequest.phoneNumber,
            otp = otp,
            requestId = otpRequest.requestId
        ).fold(
            onSuccess = { user ->
                val token = "mock-token-${System.currentTimeMillis()}"
                val session = Session(
                    userId = user.id,
                    phoneNumber = user.phoneNumber,
                    token = token,
                    userType = user.userType,
                    kycStatus = user.aadhaarKycStatus,
                    isProfileComplete = !user.fullName.isNullOrBlank()
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
    
    suspend fun authenticateWithTruecaller(token: String): Result<User> {
        return authRepository.authenticateWithTruecaller(token).fold(
            onSuccess = { user ->
                val authToken = "mock-token-${System.currentTimeMillis()}"
                val session = Session(
                    userId = user.id,
                    phoneNumber = user.phoneNumber,
                    token = authToken,
                    userType = user.userType,
                    kycStatus = user.aadhaarKycStatus,
                    isProfileComplete = !user.fullName.isNullOrBlank()
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
    
    suspend fun logout(): Result<Unit> {
        authRepository.logout()
        sessionManager.clearSession()
        currentOtpRequest = null
        _authState.value = AuthState.NotAuthenticated
        return Result.success(Unit)
    }
    
    suspend fun updateKycStatus(status: KycStatus) {
        sessionManager.updateKycStatus(status)
        sessionManager.getSession()?.let { session ->
            _authState.value = determineAuthState(session.copy(kycStatus = status))
        }
    }
    
    suspend fun updateVideoBioStatus(hasVideoBio: Boolean) {
        sessionManager.updateVideoBioStatus(hasVideoBio)
        sessionManager.getSession()?.let { session ->
            _authState.value = determineAuthState(session.copy(hasVideoBio = hasVideoBio))
        }
    }
    
    suspend fun selectRole(role: UserType): Result<Unit> {
        return authRepository.setUserType(role).fold(
            onSuccess = {
                sessionManager.getSession()?.let { session ->
                    val updatedSession = session.copy(userType = role)
                    sessionManager.saveSession(updatedSession)
                    _authState.value = determineAuthState(updatedSession)
                }
                Result.success(Unit)
            },
            onFailure = { error ->
                Result.failure(error)
            }
        )
    }

    suspend fun completeProfileCreation() {
        sessionManager.getSession()?.let { session ->
            val updatedSession = session.copy(isProfileComplete = true)
            sessionManager.saveSession(updatedSession)
            _authState.value = determineAuthState(updatedSession)
        }
    }


    
    private suspend fun determineAuthState(session: Session): AuthState {
        val user = authRepository.getCurrentUser().getOrNull() ?: User(
            id = session.userId,
            phoneNumber = session.phoneNumber,
            userType = session.userType,
            aadhaarKycStatus = session.kycStatus
        )
        // If fullName is needed for check, we might rely on session flag instead of user object if user object is constructed from session
        
        return when {
            session.userType == UserType.NONE -> AuthState.RequiresRoleSelection(user, session.token)
            !session.isProfileComplete -> AuthState.RequiresProfileCreation(user, session.token)
            session.userType == UserType.WORKER && session.kycStatus != KycStatus.VERIFIED -> AuthState.RequiresKyc(user, session.token)
            session.userType == UserType.WORKER && !session.hasVideoBio -> AuthState.RequiresVideoBio(user, session.token)
            else -> AuthState.Authenticated(user, session.token)
        }
    }
}
