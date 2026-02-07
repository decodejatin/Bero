package com.example.bero.data.network

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

/**
 * Secure token storage using EncryptedSharedPreferences
 */
class TokenManager(context: Context) {
    
    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()
    
    private val sharedPreferences: SharedPreferences = EncryptedSharedPreferences.create(
        context,
        "bero_secure_prefs",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )
    
    companion object {
        private const val KEY_ACCESS_TOKEN = "access_token"
        private const val KEY_REFRESH_TOKEN = "refresh_token"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_USER_TYPE = "user_type"
        private const val KEY_PHONE_NUMBER = "phone_number"
        private const val KEY_TOKEN_EXPIRY = "token_expiry"
        private const val KEY_KYC_STATUS = "kyc_status"
        private const val KEY_HAS_VIDEO_BIO = "has_video_bio"
        private const val KEY_IS_PROFILE_COMPLETE = "is_profile_complete"
    }
    
    fun saveTokens(tokens: AuthTokens) {
        sharedPreferences.edit().apply {
            putString(KEY_ACCESS_TOKEN, tokens.access_token)
            putString(KEY_REFRESH_TOKEN, tokens.refresh_token)
            putLong(KEY_TOKEN_EXPIRY, System.currentTimeMillis() + (tokens.expires_in * 1000))
            apply()
        }
    }
    
    fun saveUser(user: UserDto) {
        sharedPreferences.edit().apply {
            putString(KEY_USER_ID, user.id)
            putString(KEY_USER_TYPE, user.user_type)
            putString(KEY_PHONE_NUMBER, user.phone_number)
            apply()
        }
    }
    
    fun getAccessToken(): String? {
        return sharedPreferences.getString(KEY_ACCESS_TOKEN, null)
    }
    
    fun getRefreshToken(): String? {
        return sharedPreferences.getString(KEY_REFRESH_TOKEN, null)
    }
    
    fun getUserId(): String? {
        return sharedPreferences.getString(KEY_USER_ID, null)
    }
    
    fun getUserType(): String? {
        return sharedPreferences.getString(KEY_USER_TYPE, null)
    }
    
    fun getPhoneNumber(): String? {
        return sharedPreferences.getString(KEY_PHONE_NUMBER, null)
    }
    
    fun isTokenExpired(): Boolean {
        val expiry = sharedPreferences.getLong(KEY_TOKEN_EXPIRY, 0)
        return System.currentTimeMillis() >= expiry
    }
    
    fun isLoggedIn(): Boolean {
        // User is logged in if they have valid access token OR a refresh token
        val hasAccessToken = getAccessToken() != null
        val hasRefreshToken = getRefreshToken() != null
        val tokenNotExpired = !isTokenExpired()
        
        // Allow login if: (valid access token) OR (has refresh token for auto-refresh)
        return (hasAccessToken && tokenNotExpired) || hasRefreshToken
    }
    
    fun clearAll() {
        sharedPreferences.edit().clear().apply()
    }
    
    // Session persistence methods
    fun saveSession(session: com.example.bero.data.auth.Session) {
        sharedPreferences.edit().apply {
            putString(KEY_USER_ID, session.userId)
            putString(KEY_PHONE_NUMBER, session.phoneNumber)
            putString(KEY_USER_TYPE, session.userType.name)
            putString(KEY_KYC_STATUS, session.kycStatus.name)
            putBoolean(KEY_HAS_VIDEO_BIO, session.hasVideoBio)
            putBoolean(KEY_IS_PROFILE_COMPLETE, session.isProfileComplete)
            // Token is already saved separately
            apply()
        }
    }
    
    fun getStoredSession(): com.example.bero.data.auth.Session? {
        val userId = getUserId() ?: return null
        val phoneNumber = getPhoneNumber() ?: return null
        val token = getAccessToken() ?: return null
        
        // Don't check token expiry here - let the API interceptor handle refresh
        // This allows the session to persist and auto-refresh when needed
        
        val userTypeStr = getUserType() ?: "NONE"
        val kycStatusStr = sharedPreferences.getString(KEY_KYC_STATUS, "NONE") ?: "NONE"
        val hasVideoBio = sharedPreferences.getBoolean(KEY_HAS_VIDEO_BIO, false)
        val isProfileComplete = sharedPreferences.getBoolean(KEY_IS_PROFILE_COMPLETE, false)
        
        return com.example.bero.data.auth.Session(
            userId = userId,
            phoneNumber = phoneNumber,
            token = token,
            userType = try {
                com.example.bero.data.models.UserType.valueOf(userTypeStr)
            } catch (e: Exception) {
                com.example.bero.data.models.UserType.NONE
            },
            kycStatus = try {
                com.example.bero.data.models.KycStatus.valueOf(kycStatusStr)
            } catch (e: Exception) {
                com.example.bero.data.models.KycStatus.NONE
            },
            hasVideoBio = hasVideoBio,
            isProfileComplete = isProfileComplete
        )
    }
}
