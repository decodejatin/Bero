package com.example.bero.data.network

import kotlinx.serialization.Serializable

/**
 * API Response wrapper models matching backend responses
 */

// Auth Responses
@Serializable
data class OtpRequestResponse(
    val request_id: String,
    val expires_in_seconds: Int
)

@Serializable
data class AuthResponse(
    val user: UserDto,
    val tokens: AuthTokens,
    val is_new: Boolean
)

@Serializable
data class AuthTokens(
    val access_token: String,
    val refresh_token: String,
    val expires_in: Long
)

@Serializable
data class UserDto(
    val id: String,
    val phone_number: String,
    val full_name: String? = null,
    val email: String? = null,
    val aadhaar_kyc_status: String = "NONE",
    val user_type: String,
    val created_at: String? = null,
    val updated_at: String? = null
)

// Job Responses
@Serializable
data class JobDto(
    val id: String,
    val title: String,
    val description: String,
    val category: String,
    val status: String = "OPEN",
    val client_id: String,
    val client_name: String,
    val client_phone: String? = null,
    val address: String,
    val locality: String,
    val city: String = "Delhi",
    val pincode: String,
    val latitude: Double? = null,
    val longitude: Double? = null,
    val estimated_duration_minutes: Int = 60,
    val payment_amount_rupees: Double,
    val scheduled_date: String,
    val scheduled_time_slot: String,
    val is_urgent: Boolean = false,
    val required_skills: List<String> = emptyList(),
    val assigned_worker_id: String? = null,
    val assigned_worker_name: String? = null,
    val assigned_worker_phone: String? = null,
    val assigned_worker_rating: Double? = null,
    val created_at: String? = null,
    val updated_at: String? = null
)

// Worker Profile Response
@Serializable
data class WorkerProfileDto(
    val user_id: String,
    val skills: List<String> = emptyList(),
    val h3_index_res9: String? = null,
    val is_online: Boolean = false,
    val wallet_balance_micros: Long = 0,
    val rating_avg: Double = 0.0,
    val rating_count: Int = 0,
    val streak_count: Int = 0,
    val last_active_date: String? = null,
    val tier: String = "BRONZE",
    val video_bio_url: String? = null
)

// Earnings Response
@Serializable
data class EarningsSummaryDto(
    val period: String,
    val total_earnings: Double,
    val jobs_completed: Int,
    val average_per_job: Double,
    val commission_paid: Double,
    val tds_paid: Double,
    val gst_paid: Double,
    val net_earnings: Double,
    val daily_earnings: List<DailyEarningDto> = emptyList()
)

@Serializable
data class DailyEarningDto(
    val date: String,
    val amount: Double,
    val job_count: Int
)

// Request Bodies
@Serializable
data class SendOtpRequest(
    val phone_number: String
)

@Serializable
data class VerifyOtpRequest(
    val phone_number: String,
    val otp: String,
    val request_id: String
)

@Serializable
data class RefreshTokenRequest(
    val refresh_token: String
)

@Serializable
data class CreateJobRequest(
    val title: String,
    val description: String,
    val category: String,
    val client_name: String,
    val address: String,
    val locality: String,
    val city: String = "Delhi",
    val pincode: String,
    val latitude: Double? = null,
    val longitude: Double? = null,
    val estimated_duration_minutes: Int = 60,
    val payment_amount_rupees: Double,
    val scheduled_date: String,
    val scheduled_time_slot: String,
    val is_urgent: Boolean = false,
    val required_skills: List<String> = emptyList()
)

@Serializable
data class AcceptJobRequest(
    val estimated_arrival_minutes: Int = 30
)

@Serializable
data class CompleteJobRequest(
    val worker_notes: String? = null,
    val photo_proof_urls: List<String> = emptyList()
)

// Generic Responses
@Serializable
data class SuccessResponse(
    val message: String
)

@Serializable
data class ErrorResponse(
    val error: String
)

// Profile
@Serializable
data class ProfileDto(
    val id: String,
    val phone_number: String,
    val full_name: String? = null,
    val email: String? = null,
    val user_type: String,
    val address: String? = null,
    val kyc_status: String = "NONE"
)

@Serializable
data class UpdateProfileRequest(
    val full_name: String,
    val email: String? = null,
    val address: String? = null
)

@Serializable
data class SetUserTypeRequest(
    val user_type: String
)

// Chat DTOs
@Serializable
data class ConversationDto(
    val id: String,
    val participant_id: String,
    val participant_name: String = "User",
    val job_id: String? = null,
    val last_message: String = "",
    val last_message_at: String = "",
    val unread_count: Long = 0,
    val created_at: String = ""
)

@Serializable
data class ChatMessageDto(
    val id: String,
    val conversation_id: String,
    val sender_id: String,
    val content: String,
    val message_type: String = "text",
    val is_read: Boolean = false,
    val created_at: String = ""
)

@Serializable
data class CreateConversationRequest(
    val participant_id: String,
    val job_id: String? = null
)

@Serializable
data class SendChatMessageRequest(
    val content: String,
    val message_type: String = "text"
)

@Serializable
data class UnreadCountResponse(
    val unread_count: Long = 0
)

// Stats DTOs
@Serializable
data class UserStatsDto(
    val jobs_posted: Long = 0,
    val jobs_completed: Long = 0,
    val total_spent: Double = 0.0,
    val total_earned: Double = 0.0,
    val avg_rating: Double = 0.0
)

// Address DTOs
@Serializable
data class SavedAddressDto(
    val id: String = "",
    val user_id: String = "",
    val label: String = "",
    val full_address: String = "",
    val latitude: Double? = null,
    val longitude: Double? = null,
    val is_default: Boolean = false,
    val created_at: String = "",
    val updated_at: String = ""
)

@Serializable
data class CreateAddressRequest(
    val label: String,
    val full_address: String,
    val latitude: Double? = null,
    val longitude: Double? = null,
    val is_default: Boolean = false
)

// Rating DTOs
@Serializable
data class SubmitRatingRequest(
    val rating: Int,
    val review: String = "",
    val tags: List<String> = emptyList()
)

// Location DTOs — powers the geospatial infrastructure layer
@Serializable
data class UpdateLocationRequest(
    val latitude: Double,
    val longitude: Double
)

@Serializable
data class UpdateLocationResponse(
    val status: String = "",
    val h3_index: String = ""
)

@Serializable
data class SetAvailabilityRequest(
    val is_available: Boolean
)

@Serializable
data class NearbyWorkerDto(
    val worker_id: String,
    val name: String = "",
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
    val distance_meters: Double = 0.0,
    val h3_index: String = "",
    val rating_avg: Double = 0.0,
    val skills: List<String> = emptyList(),
    val tier: String = "BRONZE"
)

@Serializable
data class NearbyWorkersResponse(
    val workers: List<NearbyWorkerDto> = emptyList(),
    val query_lat: Double = 0.0,
    val query_lon: Double = 0.0,
    val radius_meters: Double = 0.0,
    val count: Int = 0
)
