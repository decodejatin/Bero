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
        // Auto-refresh token on 401 responses
        .authenticator { _, response ->
            // Don't retry if we already tried refreshing or this is a refresh endpoint
            if (response.request.url.encodedPath.contains("refresh") ||
                response.request.header("X-Token-Refreshed") != null) {
                return@authenticator null
            }
            
            synchronized(this) {
                val refreshToken = tokenManager.getRefreshToken() ?: return@authenticator null
                
                // Call refresh endpoint synchronously
                val refreshBody = json.encodeToString(RefreshTokenRequest(refreshToken))
                    .toRequestBody("application/json; charset=utf-8".toMediaType())
                val refreshRequest = Request.Builder()
                    .url(ApiConfig.baseUrl + ApiConfig.Endpoints.REFRESH_TOKEN)
                    .post(refreshBody)
                    .addHeader("Content-Type", "application/json")
                    .build()
                
                // Use a separate client without the authenticator to avoid loops
                val refreshClient = OkHttpClient.Builder()
                    .connectTimeout(ApiConfig.CONNECT_TIMEOUT_SECONDS, TimeUnit.SECONDS)
                    .readTimeout(ApiConfig.READ_TIMEOUT_SECONDS, TimeUnit.SECONDS)
                    .build()
                
                try {
                    val refreshResponse = refreshClient.newCall(refreshRequest).execute()
                    if (refreshResponse.isSuccessful) {
                        val body = refreshResponse.body?.string()
                        if (body != null) {
                            val tokens = json.decodeFromString<AuthTokens>(body)
                            tokenManager.saveTokens(tokens)
                            
                            // Retry the original request with the new token
                            response.request.newBuilder()
                                .header("Authorization", "Bearer ${tokens.access_token}")
                                .header("X-Token-Refreshed", "true")
                                .build()
                        } else null
                    } else {
                        // Refresh failed — session truly expired
                        tokenManager.clearAll()
                        null
                    }
                } catch (e: Exception) {
                    android.util.Log.e("BeroApiClient", "Token refresh failed: ${e.message}")
                    null
                }
            }
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
    
    suspend fun acceptJob(jobId: String, estimatedArrivalMinutes: Int = 30): Result<JobDto> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(AcceptJobRequest(estimatedArrivalMinutes))
            val response = post(ApiConfig.Endpoints.acceptJob(jobId), requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<JobDto>(body))
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
    
    suspend fun startJob(jobId: String): Result<JobDto> = withContext(Dispatchers.IO) {
        try {
            val response = post(ApiConfig.Endpoints.startJob(jobId), "{}")
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<JobDto>(body))
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
    
    suspend fun completeJob(jobId: String, request: CompleteJobRequest): Result<JobDto> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(request)
            val response = post(ApiConfig.Endpoints.completeJob(jobId), requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<JobDto>(body))
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
    
    suspend fun cancelJob(jobId: String): Result<JobDto> = withContext(Dispatchers.IO) {
        try {
            val response = post(ApiConfig.Endpoints.cancelJob(jobId), "{}")
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<JobDto>(body))
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
    
    suspend fun confirmJobCompletion(jobId: String): Result<JobDto> = withContext(Dispatchers.IO) {
        try {
            val response = post(ApiConfig.Endpoints.confirmJob(jobId), "{}")
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<JobDto>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to confirm job completion")
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
    
    suspend fun getWorkerProfile(workerId: String): Result<ProfileDto> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.workerProfile(workerId))
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<ProfileDto>(body))
            } else {
                Result.failure(Exception("Worker profile not found"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== CHAT API ====================
    
    suspend fun getConversations(): Result<List<ConversationDto>> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.CONVERSATIONS)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<List<ConversationDto>>(body))
            } else {
                Result.failure(Exception("Failed to get conversations: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getOrCreateConversation(participantId: String, jobId: String? = null): Result<ConversationDto> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(CreateConversationRequest(participantId, jobId))
            val response = post(ApiConfig.Endpoints.CONVERSATIONS, requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<ConversationDto>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to create conversation")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getMessages(conversationId: String, limit: Int = 50, offset: Int = 0): Result<List<ChatMessageDto>> = withContext(Dispatchers.IO) {
        try {
            val endpoint = ApiConfig.Endpoints.conversationMessages(conversationId) + "?limit=$limit&offset=$offset"
            val response = get(endpoint)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<List<ChatMessageDto>>(body))
            } else {
                Result.failure(Exception("Failed to get messages: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun sendChatMessage(conversationId: String, content: String, messageType: String = "text"): Result<ChatMessageDto> = withContext(Dispatchers.IO) {
        try {
            val requestBody = json.encodeToString(SendChatMessageRequest(content, messageType))
            val response = post(ApiConfig.Endpoints.conversationMessages(conversationId), requestBody)
            
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                Result.success(json.decodeFromString<ChatMessageDto>(body))
            } else {
                val errorBody = response.body?.string()
                val error = try {
                    json.decodeFromString<ErrorResponse>(errorBody ?: "")
                } catch (e: Exception) {
                    ErrorResponse("Failed to send message")
                }
                Result.failure(Exception(error.error))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun markConversationAsRead(conversationId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val response = put(ApiConfig.Endpoints.markConversationRead(conversationId), "")
            if (response.isSuccessful) Result.success(Unit)
            else Result.failure(Exception("Failed to mark as read"))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getUnreadCount(): Result<Int> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.UNREAD_COUNT)
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return@withContext Result.failure(Exception("Empty response"))
                val map = json.decodeFromString<Map<String, Int>>(body)
                Result.success(map["unread_count"] ?: 0)
            } else {
                Result.failure(Exception("Failed to get unread count"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== Stats ====================
    
    suspend fun getUserStats(): Result<UserStatsDto> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.PROFILE_STATS)
            val body = response.body?.string() ?: ""
            if (response.isSuccessful) {
                Result.success(json.decodeFromString(body))
            } else {
                Result.failure(Exception("Failed to get stats: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== Addresses ====================
    
    suspend fun getAddresses(): Result<List<SavedAddressDto>> = withContext(Dispatchers.IO) {
        try {
            val response = get(ApiConfig.Endpoints.ADDRESSES)
            val body = response.body?.string() ?: "[]"
            if (response.isSuccessful) {
                Result.success(json.decodeFromString(body))
            } else {
                Result.failure(Exception("Failed to get addresses: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun createAddress(request: CreateAddressRequest): Result<SavedAddressDto> = withContext(Dispatchers.IO) {
        try {
            val response = post(ApiConfig.Endpoints.ADDRESSES, json.encodeToString(request))
            val body = response.body?.string() ?: ""
            if (response.isSuccessful) {
                Result.success(json.decodeFromString(body))
            } else {
                Result.failure(Exception("Failed to create address: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun updateAddress(id: String, request: CreateAddressRequest): Result<SavedAddressDto> = withContext(Dispatchers.IO) {
        try {
            val response = put(ApiConfig.Endpoints.addressById(id), json.encodeToString(request))
            val body = response.body?.string() ?: ""
            if (response.isSuccessful) {
                Result.success(json.decodeFromString(body))
            } else {
                Result.failure(Exception("Failed to update address: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun deleteAddress(id: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val response = delete(ApiConfig.Endpoints.addressById(id))
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to delete address: ${response.code}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== Ratings ====================
    
    suspend fun submitRating(jobId: String, rating: Int, review: String = "", tags: List<String> = emptyList()): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val request = SubmitRatingRequest(rating = rating, review = review, tags = tags)
            val response = post(ApiConfig.Endpoints.submitRating(jobId), json.encodeToString(request))
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                val body = response.body?.string() ?: ""
                Result.failure(Exception("Failed to submit rating: $body"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // ==================== LOCATION API ====================
    
    /**
     * Update worker's current location.
     * Called every 10 seconds by LocationTrackingService.
     * Returns raw JSON string for LocationRepository to deserialize.
     */
    suspend fun updateWorkerLocation(latitude: Double, longitude: Double): String = withContext(Dispatchers.IO) {
        val request = UpdateLocationRequest(latitude, longitude)
        val response = put(ApiConfig.Endpoints.WORKER_LOCATION, json.encodeToString(request))
        response.body?.string() ?: throw Exception("Empty response from location update")
    }
    
    /**
     * Query nearby available workers around a point.
     * Returns raw JSON string for LocationRepository to deserialize.
     */
    suspend fun getNearbyWorkers(latitude: Double, longitude: Double, radius: Int = 2000): String = withContext(Dispatchers.IO) {
        val endpoint = "${ApiConfig.Endpoints.NEARBY_WORKERS}?lat=$latitude&lon=$longitude&radius=$radius"
        val response = get(endpoint)
        response.body?.string() ?: throw Exception("Empty response from nearby workers query")
    }
    
    /**
     * Toggle worker availability (online/offline for job matching).
     */
    suspend fun setWorkerAvailability(available: Boolean): String = withContext(Dispatchers.IO) {
        val request = SetAvailabilityRequest(available)
        val response = put(ApiConfig.Endpoints.WORKER_AVAILABILITY, json.encodeToString(request))
        response.body?.string() ?: throw Exception("Empty response from availability update")
    }
    
    // ==================== Completion Flow ====================
    
    /**
     * Worker marks a job as completed.
     */
    suspend fun workerMarkComplete(jobId: String): String = withContext(Dispatchers.IO) {
        val body = """{"worker_id":"${tokenManager.getUserId()}"}"""
        val response = post("/api/v1/jobs/$jobId/complete-by-worker", body)
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Client confirms job completion.
     */
    suspend fun clientConfirmComplete(jobId: String): String = withContext(Dispatchers.IO) {
        val body = """{"client_id":"${tokenManager.getUserId()}"}"""
        val response = post("/api/v1/jobs/$jobId/confirm-by-client", body)
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Submit mandatory mutual rating after job is fully completed.
     */
    suspend fun submitMutualRating(jobId: String, rating: Int, review: String = ""): String = withContext(Dispatchers.IO) {
        val body = """{"rater_id":"${tokenManager.getUserId()}","rating":$rating,"review":"$review"}"""
        val response = post("/api/v1/jobs/$jobId/rate", body)
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Get dual-sided completion status for a job.
     */
    suspend fun getCompletionStatus(jobId: String): String = withContext(Dispatchers.IO) {
        val response = get("/api/v1/jobs/$jobId/completion-status")
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Check if worker is blocked by pending rating.
     */
    suspend fun checkWorkerBlocked(workerId: String): String = withContext(Dispatchers.IO) {
        val response = get("/api/v1/completion/blocked/worker/$workerId")
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Check if client is blocked by pending rating.
     */
    suspend fun checkClientBlocked(clientId: String): String = withContext(Dispatchers.IO) {
        val response = get("/api/v1/completion/blocked/client/$clientId")
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    // ==================== Dynamic Pricing ====================
    
    /**
     * Get current surge multiplier for an H3 hexagon.
     */
    suspend fun getSurgePrice(h3Index: String): String = withContext(Dispatchers.IO) {
        val response = get("/api/v1/pricing/surge?h3=$h3Index")
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Get full price quote with surge for a job.
     */
    suspend fun getPriceQuote(jobId: String): String = withContext(Dispatchers.IO) {
        val response = get("/api/v1/pricing/quote/$jobId")
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    // ==================== Pipeline ====================
    
    /**
     * Submit job to the unified dispatch pipeline (pricing + enqueue).
     */
    suspend fun submitJobToPipeline(jobId: String): String = withContext(Dispatchers.IO) {
        val body = """{"job_id":"$jobId"}"""
        val response = post("/api/v1/pipeline/submit", body)
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Get pipeline status (queue depth, last batch, lifetime stats).
     */
    suspend fun getPipelineStatus(): String = withContext(Dispatchers.IO) {
        val response = get("/api/v1/pipeline/status")
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    // ==================== Stability ====================
    
    /**
     * Get cancellation/stability status for a user.
     */
    suspend fun getStabilityStatus(userId: String): String = withContext(Dispatchers.IO) {
        val response = get("/api/v1/stability/user/$userId/status")
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    // ==================== Matching ====================
    
    /**
     * Decline an assigned job (triggers requeue).
     */
    suspend fun declineJob(jobId: String): String = withContext(Dispatchers.IO) {
        val body = """{"worker_id":"${tokenManager.getUserId()}"}"""
        val response = post("/api/v1/matching/decline/$jobId", body)
        response.body?.string() ?: throw Exception("Empty response")
    }
    
    /**
     * Go offline (remove worker location, mark unavailable).
     */
    suspend fun goOffline(): String = withContext(Dispatchers.IO) {
        val response = delete("/api/v1/workers/location")
        response.body?.string() ?: throw Exception("Empty response")
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
    
    private fun delete(endpoint: String): okhttp3.Response {
        val request = Request.Builder()
            .url(ApiConfig.baseUrl + endpoint)
            .delete()
            .addHeader("Content-Type", "application/json")
            .build()
        return client.newCall(request).execute()
    }
}
