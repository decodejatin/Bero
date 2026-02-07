package com.example.bero.data.models

import java.util.UUID

/**
 * Represents a job posting in the Bero platform
 */
data class Job(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val description: String,
    val category: ServiceCategory,
    val location: String,
    val landmark: String? = null,
    val amountRupees: Double,
    val status: JobStatus = JobStatus.OPEN,
    val clientId: String,
    val clientName: String,
    val clientRating: Double = 4.5,
    val workerId: String? = null,
    val workerName: String? = null,
    val workerPhone: String? = null,
    val workerRating: Double? = null,
    val postedAt: Long = System.currentTimeMillis(),
    val scheduledDate: String = "", // Keep as String for display
    val scheduledTime: String? = null,
    val scheduledTimeSlot: String = "9 AM - 12 PM",
    val distance: Double = 0.0, // in km
    val urgency: JobUrgency = JobUrgency.NORMAL,
    // Additional fields for UI
    val address: String = "",
    val locality: String = "",
    val city: String = "",
    val pincode: String = "",
    val estimatedDurationMinutes: Int = 60,
    val requiredSkills: List<String> = emptyList(),
    val workerConfirmed: Boolean = false,
    val clientConfirmed: Boolean = false
) {
    // Computed properties for compatibility
    val paymentAmountRupees: Double get() = amountRupees
    val isUrgent: Boolean get() = urgency == JobUrgency.URGENT
}

enum class JobStatus {
    OPEN,
    ACCEPTED,
    ASSIGNED, // Alias for ACCEPTED in some screens
    IN_PROGRESS,
    AWAITING_CONFIRMATION,
    COMPLETED,
    CANCELLED,
    DISPUTED
}

enum class JobUrgency {
    LOW,
    NORMAL,
    URGENT
}

/**
 * Service categories available in Bero
 */
enum class ServiceCategory(
    val displayName: String,
    val emoji: String,
    val color: Long // ARGB color
) {
    PLUMBER("Plumber", "🔧", 0xFF2196F3),
    PLUMBING("Plumbing", "🔧", 0xFF2196F3), // Alias
    ELECTRICIAN("Electrician", "⚡", 0xFFFFC107),
    ELECTRICAL("Electrical", "⚡", 0xFFFFC107), // Alias
    CARPENTER("Carpenter", "🪚", 0xFF8D6E63),
    CARPENTRY("Carpentry", "🪚", 0xFF795548), // Alias
    PAINTER("Painter", "🎨", 0xFFE91E63),
    PAINTING("Painting", "🎨", 0xFF9C27B0), // Alias
    CLEANER("Cleaner", "🧹", 0xFF4CAF50),
    CLEANING("Cleaning", "🧹", 0xFF00BCD4), // Alias
    AC_REPAIR("AC Repair", "❄️", 0xFF00BCD4),
    APPLIANCE_REPAIR("Appliance Repair", "🔌", 0xFF9C27B0),
    PEST_CONTROL("Pest Control", "🐛", 0xFFFF5722),
    GARDENER("Gardener", "🌱", 0xFF8BC34A),
    GARDENING("Gardening", "🌱", 0xFF8BC34A), // Alias
    DRIVER("Driver", "🚗", 0xFF607D8B),
    COOK("Cook", "👨‍🍳", 0xFFFF9800),
    HELPER("Helper", "🤝", 0xFF3F51B5),
    OTHER("Other", "🔨", 0xFF9E9E9E)
}

/**
 * Transaction in the wallet
 */
data class Transaction(
    val id: String = UUID.randomUUID().toString(),
    val type: TransactionType,
    val amountRupees: Double,
    val description: String,
    val jobId: String? = null,
    val timestamp: Long = System.currentTimeMillis(),
    val status: TransactionStatus = TransactionStatus.COMPLETED
)

enum class TransactionType {
    JOB_EARNING,
    COMMISSION_DEDUCTION,
    TDS_DEDUCTION,
    GST_DEDUCTION,
    WALLET_RECHARGE,
    WITHDRAWAL,
    BONUS,
    STREAK_REWARD
}

enum class TransactionStatus {
    PENDING,
    COMPLETED,
    FAILED
}

/**
 * Chat message between client and worker
 */
data class ChatMessage(
    val id: String = UUID.randomUUID().toString(),
    val senderId: String,
    val receiverId: String,
    val message: String,
    val timestamp: Long = System.currentTimeMillis(),
    val isRead: Boolean = false,
    val messageType: MessageType = MessageType.TEXT
)

enum class MessageType {
    TEXT,
    IMAGE,
    LOCATION,
    SYSTEM
}

/**
 * Conversation summary for chat list
 */
data class Conversation(
    val id: String,
    val participantId: String,
    val participantName: String,
    val participantImage: String? = null,
    val lastMessage: String,
    val lastMessageTime: Long,
    val unreadCount: Int = 0,
    val jobTitle: String? = null
)

/**
 * Notification item
 */
data class NotificationItem(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val message: String,
    val type: NotificationType,
    val timestamp: Long = System.currentTimeMillis(),
    val isRead: Boolean = false,
    val actionData: String? = null // e.g., jobId for navigation
)

enum class NotificationType {
    NEW_JOB,
    JOB_ACCEPTED,
    JOB_COMPLETED,
    PAYMENT_RECEIVED,
    STREAK_MILESTONE,
    RATING_RECEIVED,
    SYSTEM_UPDATE
}

/**
 * Review/Rating given by users
 */
data class Review(
    val id: String = UUID.randomUUID().toString(),
    val jobId: String,
    val reviewerId: String,
    val reviewerName: String,
    val revieweeId: String,
    val rating: Float, // 1.0 to 5.0
    val comment: String? = null,
    val timestamp: Long = System.currentTimeMillis(),
    val tags: List<String> = emptyList() // e.g., "On Time", "Professional"
)

/**
 * Enhanced Worker Profile for display
 */
data class WorkerDisplayProfile(
    val userId: String,
    val name: String,
    val phoneNumber: String,
    val skills: List<ServiceCategory>,
    val rating: Double,
    val totalJobs: Int,
    val isOnline: Boolean,
    val isKycVerified: Boolean,
    val hasVideoBio: Boolean,
    val tier: WorkerTier,
    val streakCount: Int,
    val memberSince: String,
    val location: String,
    val distance: Double? = null // in km
)



/**
 * Booking for client view
 */
data class Booking(
    val id: String = UUID.randomUUID().toString(),
    val job: Job,
    val worker: WorkerDisplayProfile? = null,
    val createdAt: Long = System.currentTimeMillis(),
    val completedAt: Long? = null,
    val rating: Float? = null,
    val review: String? = null
)

/**
 * Worker skill/service with pricing
 */
data class Skill(
    val id: String = UUID.randomUUID().toString(),
    val category: ServiceCategory,
    val pricePerHour: Double,
    val minimumCharge: Double,
    val isActive: Boolean = true,
    val isVerified: Boolean = false,
    val experienceYears: Int = 0,
    val description: String? = null
)

/**
 * Payment method for clients
 */
data class PaymentMethod(
    val id: String = UUID.randomUUID().toString(),
    val type: PaymentMethodType,
    val displayName: String,
    val lastFourDigits: String? = null,
    val upiId: String? = null,
    val isDefault: Boolean = false,
    val addedAt: Long = System.currentTimeMillis()
)

enum class PaymentMethodType {
    UPI,
    DEBIT_CARD,
    CREDIT_CARD,
    WALLET,
    NET_BANKING
}

/**
 * Earnings summary for analytics
 */
data class EarningsSummary(
    val period: String, // "Week", "Month", "Year"
    val totalEarnings: Double,
    val jobsCompleted: Int,
    val averagePerJob: Double,
    val commissionPaid: Double,
    val tdsPaid: Double,
    val gstPaid: Double,
    val netEarnings: Double,
    val dailyEarnings: List<DailyEarning> = emptyList()
)

data class DailyEarning(
    val date: String,
    val amount: Double,
    val jobCount: Int
)

/**
 * FAQ for help screen
 */
data class FAQ(
    val id: String = UUID.randomUUID().toString(),
    val question: String,
    val answer: String,
    val category: FAQCategory,
    val isPopular: Boolean = false
)

enum class FAQCategory {
    GETTING_STARTED,
    PAYMENTS,
    BOOKINGS,
    ACCOUNT,
    SAFETY,
    OTHER
}

/**
 * Support ticket
 */
data class SupportTicket(
    val id: String = UUID.randomUUID().toString(),
    val subject: String,
    val description: String,
    val category: String,
    val status: TicketStatus = TicketStatus.OPEN,
    val createdAt: Long = System.currentTimeMillis(),
    val updatedAt: Long = System.currentTimeMillis()
)

enum class TicketStatus {
    OPEN,
    IN_PROGRESS,
    RESOLVED,
    CLOSED
}

/**
 * Notification preference settings
 */
data class NotificationPreference(
    val type: NotificationPreferenceType,
    val enabled: Boolean = true,
    val sound: Boolean = true
)

enum class NotificationPreferenceType {
    JOB_ALERTS,
    BOOKING_UPDATES,
    PAYMENT_NOTIFICATIONS,
    CHAT_MESSAGES,
    PROMOTIONAL_OFFERS,
    APP_UPDATES
}
