package com.example.bero.data.network

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.logging.HttpLoggingInterceptor
import java.util.concurrent.TimeUnit

/**
 * HTTP Client for Bero API
 * Handles all network requests to the Go backend
 */
class BeroApiClient(private val tokenManager: TokenManager) {
    
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        encodeDefaults = true
    }
    
    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = HttpLoggingInterceptor.Level.BODY
    }
    
    private val client = OkHttpClient.Builder()
        .connectTimeout(ApiConfig.CONNECT_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .readTimeout(ApiConfig.READ_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .writeTimeout(ApiConfig.WRITE_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .addInterceptor(loggingInterceptor)
        .addInterceptor { chain ->
            val originalRequest = chain.request()
            val token = tokenManager.getAccessToken()
            
            val newRequest = if (token != null && !isPublicEndpoint(originalRequest.url.encodedPath)) {
                originalRequest.newBuilder()
                    .addHeader("Authorization", "Bearer $token")
                    .build()
            } else {
                originalRequest
            }
            chain.proceed(newRequest)
        }
        .build()
    
    private fun isPublicEndpoint(path: String): Boolean {
        return path.contains("send-otp") || 
               path.contains("verify-otp") || 
               path.contains("refresh")
    }
    
    // ==================== AUTH API ====================
    
    suspend fun sendOtp(phoneNumber: String): Result<OtpRequestResponse> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(SendOtpRequest(phoneNumber))
            val response = post(ApiConfig.Endpoints.SEND_OTP, requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<OtpRequestResponse>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Request failed: ${response.code}")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun verifyOtp(phoneNumber: String, otp: String, requestId: String): Result<AuthResponse> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(VerifyOtpRequest(phoneNumber, otp, requestId))
            val response = post(ApiConfig.Endpoints.VERIFY_OTP, requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                val authResponse = json.decodeFromString<AuthResponse>(body)
                
                // Save tokens and user info
                tokenManager.saveTokens(authResponse.tokens)
                tokenManager.saveUser(authResponse.user)
                
                Result.success(authResponse)
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Verification failed: ${response.code}")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun refreshToken(): Result<AuthTokens> = withContext(Dispatchers.IO) {
        try {
            val refreshToken = tokenManager.getRefreshToken() 
                ?: return@withContext Result.failure(Exception("No refresh token"))
            
            val requestBody = json.encodeToString(RefreshTokenRequest(refreshToken))
            val response = post(ApiConfig.Endpoints.REFRESH_TOKEN, requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                val tokens = json.decodeFromString<AuthTokens>(body)
                tokenManager.saveTokens(tokens)
                Result.success(tokens)
            } else {
                // Token refresh failed, user needs to re-login
                tokenManager.clearAll()
                Result.failure(Exception("Session expired"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun logout(): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            post(ApiConfig.Endpoints.LOGOUT, "")
            tokenManager.clearAll()
            Result.success(Unit)
        } catch (e: Exception) {
            tokenManager.clearAll() // Clear anyway
            Result.success(Unit)
        }
    }
    
    suspend fun getCurrentUser(): Result<UserDto> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.ME)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<UserDto>(body))
            } else {
                Result.failure(Exception("Failed to get user: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== JOBS API ====================
    
    suspend fun getAvailableJobs(
        locality: String? = null,
        category: String? = null,
        limit: Int = 20,
        offset: Int = 0
    ): Result<List<JobDto>> = withContext(Dispatchers.IO) {
        try {
            val params = mutableListOf<String>()
            locality?.let { params.add("locality=$it") }
            category?.let { params.add("category=$it") }
            params.add("limit=$limit")
            params.add("offset=$offset")
            
            val queryString = if (params.isNotEmpty()) "?" + params.joinToString("&") else ""
            val response = get(ApiConfig.Endpoints.JOBS + queryString)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<List<JobDto>>(body))
            } else {
                Result.failure(Exception("Failed to get jobs: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getMyJobs(
        status: String? = null,
        limit: Int = 20,
        offset: Int = 0
    ): Result<List<JobDto>> = withContext(Dispatchers.IO) {
        try {
            val params = mutableListOf<String>()
            status?.let { params.add("status=$it") }
            params.add("limit=$limit")
            params.add("offset=$offset")
            
            val queryString = if (params.isNotEmpty()) "?" + params.joinToString("&") else ""
            val response = get(ApiConfig.Endpoints.MY_JOBS + queryString)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<List<JobDto>>(body))
            } else {
                Result.failure(Exception("Failed to get my jobs: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getJobDetails(jobId: String): Result<JobDto> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.jobDetails(jobId))
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<JobDto>(body))
            } else {
                Result.failure(Exception("Job not found"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun createJob(request: CreateJobRequest): Result<JobDto> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(request)
            val response = post(ApiConfig.Endpoints.JOBS, requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<JobDto>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to create job")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun acceptJob(jobId: String, estimatedArrivalMinutes: Int = 30): Result<SuccessResponse> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(AcceptJobRequest(estimatedArrivalMinutes))
            val response = post(ApiConfig.Endpoints.acceptJob(jobId), requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<SuccessResponse>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to accept job")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun startJob(jobId: String): Result<SuccessResponse> = withContext(Dispatchers.IO) {
        try {
            val response = post(ApiConfig.Endpoints.startJob(jobId), "")
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<SuccessResponse>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to start job")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun completeJob(jobId: String, request: CompleteJobRequest): Result<SuccessResponse> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(request)
            val response = post(ApiConfig.Endpoints.completeJob(jobId), requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<SuccessResponse>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to complete job")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun cancelJob(jobId: String): Result<SuccessResponse> = withContext(Dispatchers.IO) {
        try {
            val response = post(ApiConfig.Endpoints.cancelJob(jobId), "")
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<SuccessResponse>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to cancel job")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== PROFILE API ====================
    
    suspend fun getProfile(): Result<ProfileDto> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.PROFILE)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<ProfileDto>(body))
            } else {
                Result.failure(Exception("Failed to get profile: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun updateProfile(fullName: String, email: String?, address: String?): Result<ProfileDto> = withContext(Dispatchers.IO) {
        try {
            val request = UpdateProfileRequest(fullName, email, address)
            val requestBody = json.encodeToString(request)
            val response = put(ApiConfig.Endpoints.PROFILE, requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<ProfileDto>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to update profile")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun setUserType(userType: String): Result<SuccessResponse> = withContext(Dispatchers.IO) {
        try {
            val request = SetUserTypeRequest(userType)
            val requestBody = json.encodeToString(request)
            val response = post(ApiConfig.Endpoints.SET_USER_TYPE, requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<SuccessResponse>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to set user type")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== HTTP Helpers ====================
    
    private fun get(endpoint: String): okhttp3.Response {
        val request = Request.Builder()
            .url(ApiConfig.baseUrl + endpoint)
            .get()
            .addHeader("Content-Type", "application/json")
            .build()
        return client.newCall(request).execute()
    }
    
    private fun post(endpoint: String, body: String): okhttp3.Response {
        val mediaType = "application/json; charset=utf-8".toMediaType()
        val requestBody = body.toRequestBody(mediaType)
        
        val request = Request.Builder()
            .url(ApiConfig.baseUrl + endpoint)
            .post(requestBody)
            .addHeader("Content-Type", "application/json")
            .build()
        return client.newCall(request).execute()
    }
    
    private fun put(endpoint: String, body: String): okhttp3.Response {
        val mediaType = "application/json; charset=utf-8".toMediaType()
        val requestBody = body.toRequestBody(mediaType)
        
        val request = Request.Builder()
            .url(ApiConfig.baseUrl + endpoint)
            .put(requestBody)
            .addHeader("Content-Type", "application/json")
            .build()
        return client.newCall(request).execute()
    }
}
