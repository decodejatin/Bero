package com.example.bero.data

import com.example.bero.data.models.*

/**
 * Provides dummy data for the app during development
 */
object DummyDataProvider {

    // Sample worker profiles
    val sampleWorkers = listOf(
        WorkerDisplayProfile(
            userId = "w1",
            name = "Rajesh Kumar",
            phoneNumber = "+91 98765 43210",
            skills = listOf(ServiceCategory.PLUMBER, ServiceCategory.AC_REPAIR),
            rating = 4.8,
            totalJobs = 156,
            isOnline = true,
            isKycVerified = true,
            hasVideoBio = true,
            tier = WorkerTier.GOLD,
            streakCount = 45,
            memberSince = "Jan 2024",
            location = "Sector 15, Noida",
            distance = 1.2
        ),
        WorkerDisplayProfile(
            userId = "w2",
            name = "Suresh Sharma",
            phoneNumber = "+91 98765 43211",
            skills = listOf(ServiceCategory.ELECTRICIAN),
            rating = 4.6,
            totalJobs = 89,
            isOnline = true,
            isKycVerified = true,
            hasVideoBio = false,
            tier = WorkerTier.SILVER,
            streakCount = 12,
            memberSince = "Mar 2024",
            location = "Sector 18, Noida",
            distance = 2.5
        ),
        WorkerDisplayProfile(
            userId = "w3",
            name = "Amit Patel",
            phoneNumber = "+91 98765 43212",
            skills = listOf(ServiceCategory.CARPENTER, ServiceCategory.PAINTER),
            rating = 4.9,
            totalJobs = 234,
            isOnline = false,
            isKycVerified = true,
            hasVideoBio = true,
            tier = WorkerTier.GOLD,
            streakCount = 0,
            memberSince = "Nov 2023",
            location = "Sector 22, Noida",
            distance = 3.8
        ),
        WorkerDisplayProfile(
            userId = "w4",
            name = "Vikram Singh",
            phoneNumber = "+91 98765 43213",
            skills = listOf(ServiceCategory.CLEANER, ServiceCategory.GARDENER),
            rating = 4.3,
            totalJobs = 45,
            isOnline = true,
            isKycVerified = true,
            hasVideoBio = false,
            tier = WorkerTier.BRONZE,
            streakCount = 3,
            memberSince = "Aug 2024",
            location = "Sector 10, Noida",
            distance = 0.8
        ),
        WorkerDisplayProfile(
            userId = "w5",
            name = "Pradeep Yadav",
            phoneNumber = "+91 98765 43214",
            skills = listOf(ServiceCategory.AC_REPAIR, ServiceCategory.APPLIANCE_REPAIR),
            rating = 4.7,
            totalJobs = 112,
            isOnline = true,
            isKycVerified = true,
            hasVideoBio = true,
            tier = WorkerTier.SILVER,
            streakCount = 8,
            memberSince = "Feb 2024",
            location = "Sector 12, Noida",
            distance = 1.5
        )
    )

    // Sample available jobs for workers
    val sampleJobs = listOf(
        Job(
            id = "j1",
            title = "Kitchen Sink Repair",
            description = "Water leaking from kitchen sink pipe. Need urgent repair.",
            category = ServiceCategory.PLUMBER,
            location = "Sector 15, Noida",
            landmark = "Near Big Bazaar",
            amountRupees = 500.0,
            clientId = "c1",
            clientName = "Priya Gupta",
            clientRating = 4.8,
            scheduledDate = "Today",
            scheduledTime = "2:00 PM",
            distance = 1.2,
            urgency = JobUrgency.URGENT
        ),
        Job(
            id = "j2",
            title = "AC Not Cooling",
            description = "Split AC in bedroom not cooling properly. Gas may need refilling.",
            category = ServiceCategory.AC_REPAIR,
            location = "Sector 18, Noida",
            landmark = "Opposite Metro Station",
            amountRupees = 800.0,
            clientId = "c2",
            clientName = "Rahul Mehta",
            clientRating = 4.5,
            scheduledDate = "Tomorrow",
            scheduledTime = "10:00 AM",
            distance = 2.5,
            urgency = JobUrgency.NORMAL
        ),
        Job(
            id = "j3",
            title = "Full House Deep Cleaning",
            description = "3BHK apartment needs complete deep cleaning including kitchen and bathrooms.",
            category = ServiceCategory.CLEANER,
            location = "Sector 22, Noida",
            landmark = "Near City Mall",
            amountRupees = 2500.0,
            clientId = "c3",
            clientName = "Sneha Kapoor",
            clientRating = 4.9,
            scheduledDate = "30 Jan",
            scheduledTime = "9:00 AM",
            distance = 3.8,
            urgency = JobUrgency.LOW
        ),
        Job(
            id = "j4",
            title = "Ceiling Fan Installation",
            description = "Install 2 new ceiling fans in living room and bedroom.",
            category = ServiceCategory.ELECTRICIAN,
            location = "Sector 10, Noida",
            landmark = "Near Community Park",
            amountRupees = 600.0,
            clientId = "c4",
            clientName = "Arun Sharma",
            clientRating = 4.6,
            scheduledDate = "Today",
            scheduledTime = "5:00 PM",
            distance = 0.8,
            urgency = JobUrgency.NORMAL
        ),
        Job(
            id = "j5",
            title = "Wardrobe Repair",
            description = "Wardrobe door hinges broken. Need replacement and fixing.",
            category = ServiceCategory.CARPENTER,
            location = "Sector 12, Noida",
            landmark = "Near ICICI Bank",
            amountRupees = 450.0,
            clientId = "c5",
            clientName = "Kavita Joshi",
            clientRating = 4.7,
            scheduledDate = "31 Jan",
            scheduledTime = "11:00 AM",
            distance = 1.5,
            urgency = JobUrgency.LOW
        )
    )

    // Sample transactions for wallet
    val sampleTransactions = listOf(
        Transaction(
            type = TransactionType.JOB_EARNING,
            amountRupees = 800.0,
            description = "AC Repair - Sector 18",
            jobId = "j101",
            timestamp = System.currentTimeMillis() - 3600000
        ),
        Transaction(
            type = TransactionType.COMMISSION_DEDUCTION,
            amountRupees = -120.0,
            description = "Platform Commission (15%)",
            jobId = "j101",
            timestamp = System.currentTimeMillis() - 3600000
        ),
        Transaction(
            type = TransactionType.STREAK_REWARD,
            amountRupees = 100.0,
            description = "7-Day Streak Bonus 🔥",
            timestamp = System.currentTimeMillis() - 86400000
        ),
        Transaction(
            type = TransactionType.JOB_EARNING,
            amountRupees = 500.0,
            description = "Plumbing Work - Sector 15",
            jobId = "j100",
            timestamp = System.currentTimeMillis() - 172800000
        ),
        Transaction(
            type = TransactionType.COMMISSION_DEDUCTION,
            amountRupees = -75.0,
            description = "Platform Commission (15%)",
            jobId = "j100",
            timestamp = System.currentTimeMillis() - 172800000
        ),
        Transaction(
            type = TransactionType.WITHDRAWAL,
            amountRupees = -1000.0,
            description = "Bank Transfer to **4521",
            timestamp = System.currentTimeMillis() - 259200000
        ),
        Transaction(
            type = TransactionType.BONUS,
            amountRupees = 200.0,
            description = "Welcome Bonus 🎉",
            timestamp = System.currentTimeMillis() - 604800000
        )
    )

    // Sample conversations
    val sampleConversations = listOf(
        Conversation(
            id = "conv1",
            participantId = "c1",
            participantName = "Priya Gupta",
            lastMessage = "Okay, I'll be waiting. Thank you!",
            lastMessageTime = System.currentTimeMillis() - 1800000,
            unreadCount = 2,
            jobTitle = "Kitchen Sink Repair"
        ),
        Conversation(
            id = "conv2",
            participantId = "c2",
            participantName = "Rahul Mehta",
            lastMessage = "Can you come by 10 AM tomorrow?",
            lastMessageTime = System.currentTimeMillis() - 7200000,
            unreadCount = 0,
            jobTitle = "AC Not Cooling"
        ),
        Conversation(
            id = "conv3",
            participantId = "c3",
            participantName = "Sneha Kapoor",
            lastMessage = "Great work! Will definitely recommend you.",
            lastMessageTime = System.currentTimeMillis() - 86400000,
            unreadCount = 0,
            jobTitle = "Deep Cleaning"
        )
    )

    // Sample chat messages
    val sampleChatMessages = listOf(
        ChatMessage(
            senderId = "c1",
            receiverId = "w1",
            message = "Hi, are you available for sink repair today?",
            timestamp = System.currentTimeMillis() - 3600000
        ),
        ChatMessage(
            senderId = "w1",
            receiverId = "c1",
            message = "Yes, I can come by 2 PM. Will that work?",
            timestamp = System.currentTimeMillis() - 3300000
        ),
        ChatMessage(
            senderId = "c1",
            receiverId = "w1",
            message = "Perfect! Please bring necessary tools.",
            timestamp = System.currentTimeMillis() - 3000000
        ),
        ChatMessage(
            senderId = "w1",
            receiverId = "c1",
            message = "Sure, I'll bring everything needed. See you soon!",
            timestamp = System.currentTimeMillis() - 2700000
        ),
        ChatMessage(
            senderId = "c1",
            receiverId = "w1",
            message = "Okay, I'll be waiting. Thank you!",
            timestamp = System.currentTimeMillis() - 1800000
        )
    )

    // Sample notifications
    val sampleNotifications = listOf(
        NotificationItem(
            title = "New Job Available! 🔧",
            message = "Kitchen Sink Repair in Sector 15 - ₹500",
            type = NotificationType.NEW_JOB,
            timestamp = System.currentTimeMillis() - 1800000,
            actionData = "j1"
        ),
        NotificationItem(
            title = "Payment Received 💰",
            message = "₹680 credited for AC Repair job",
            type = NotificationType.PAYMENT_RECEIVED,
            timestamp = System.currentTimeMillis() - 3600000,
            isRead = true
        ),
        NotificationItem(
            title = "Streak Milestone! 🔥",
            message = "Congratulations! You've completed 7 days streak. Silver Tier unlocked!",
            type = NotificationType.STREAK_MILESTONE,
            timestamp = System.currentTimeMillis() - 86400000,
            isRead = true
        ),
        NotificationItem(
            title = "5-Star Rating ⭐",
            message = "Priya Gupta rated you 5 stars for Kitchen Sink Repair",
            type = NotificationType.RATING_RECEIVED,
            timestamp = System.currentTimeMillis() - 172800000,
            isRead = true
        ),
        NotificationItem(
            title = "Job Completed ✅",
            message = "Your AC Repair job has been marked complete",
            type = NotificationType.JOB_COMPLETED,
            timestamp = System.currentTimeMillis() - 259200000,
            isRead = true
        )
    )

    // Sample reviews
    val sampleReviews = listOf(
        Review(
            jobId = "j100",
            reviewerId = "c1",
            reviewerName = "Priya Gupta",
            revieweeId = "w1",
            rating = 5.0f,
            comment = "Excellent work! Very professional and quick. Highly recommended!",
            tags = listOf("On Time", "Professional", "Clean Work"),
            timestamp = System.currentTimeMillis() - 86400000
        ),
        Review(
            jobId = "j99",
            reviewerId = "c2",
            reviewerName = "Rahul Mehta",
            revieweeId = "w1",
            rating = 4.5f,
            comment = "Good job, fixed the AC properly. Slightly delayed but work quality was great.",
            tags = listOf("Quality Work", "Experienced"),
            timestamp = System.currentTimeMillis() - 172800000
        ),
        Review(
            jobId = "j98",
            reviewerId = "c3",
            reviewerName = "Sneha Kapoor",
            revieweeId = "w1",
            rating = 5.0f,
            comment = "Best plumber in the area! Will definitely call again.",
            tags = listOf("On Time", "Friendly", "Fair Price"),
            timestamp = System.currentTimeMillis() - 259200000
        ),
        Review(
            jobId = "j97",
            reviewerId = "c4",
            reviewerName = "Arun Sharma",
            revieweeId = "w1",
            rating = 4.0f,
            comment = "Good work overall.",
            tags = listOf("Professional"),
            timestamp = System.currentTimeMillis() - 345600000
        )
    )

    // Wallet balance
    val currentWalletBalance = 2850.0
    val pendingBalance = 500.0

    // Streak data
    val currentStreak = 12
    val longestStreak = 45
    val streakFreezeAvailable = 2

    // Statistics
    val totalEarningsThisMonth = 12500.0
    val totalJobsThisMonth = 18
    val averageRating = 4.8

    // Client bookings
    val sampleBookings = sampleJobs.mapIndexed { index, job ->
        Booking(
            job = job.copy(
                status = when (index) {
                    0 -> JobStatus.ACCEPTED
                    1 -> JobStatus.OPEN
                    2 -> JobStatus.IN_PROGRESS
                    else -> JobStatus.COMPLETED
                }
            ),
            worker = if (index != 1) sampleWorkers.getOrNull(index) else null,
            rating = if (index >= 3) 4.5f else null
        )
    }

    // Heper methods
    fun getWorkers(): List<WorkerDisplayProfile> = sampleWorkers
    
    fun getWorkerById(id: String): WorkerDisplayProfile? = sampleWorkers.find { it.userId == id }
    
    fun getReviews(): List<Review> = sampleReviews
    
    fun getJobs(): List<Job> = sampleJobs
    
    fun getJobById(id: String): Job? = sampleJobs.find { it.id == id }
}
