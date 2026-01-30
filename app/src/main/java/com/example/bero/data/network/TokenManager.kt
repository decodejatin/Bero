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
        return getAccessToken() != null && !isTokenExpired()
    }
    
    fun clearAll() {
        sharedPreferences.edit().clear().apply()
    }
}
