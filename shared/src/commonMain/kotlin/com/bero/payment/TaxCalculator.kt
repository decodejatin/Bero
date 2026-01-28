package com.bero.payment

import com.bero.domain.models.WorkerProfile
import com.bero.domain.models.WorkerTier

/**
 * Transaction breakdown for a job payment
 * All amounts in micros (1 rupee = 100000 micros for precision)
 */
data class TransactionBreakdown(
    val grossAmountMicros: Long,
    val commissionMicros: Long,
    val commissionRate: Double,
    val gstOnCommissionMicros: Long, // 18% GST on commission
    val tdsDeductionMicros: Long,    // 1% TDS under Section 194-O
    val workerPayoutMicros: Long
)

/**
 * Tax Calculator for Indian e-commerce compliance
 * 
 * Implements:
 * - Section 194-O TDS (1% on gross amount)
 * - GST on commission (18%)
 * - Tiered commission rates based on worker streak
 */
object TaxCalculator {

    private const val GST_RATE = 0.18      // 18% GST on commission
    private const val TDS_RATE = 0.01      // 1% TDS under Section 194-O
    private const val MICROS_PER_RUPEE = 100000L

    /**
     * Calculate complete transaction breakdown for a job
     * 
     * Example: ₹1,000 Job (Bronze tier worker)
     * - Job Fee: ₹1,000
     * - Platform Commission (15%): ₹150
     * - GST on Commission (18%): ₹27
     * - TDS Deduction (1% of Gross): ₹10
     * - Worker Payout: ₹813
     * 
     * @param grossAmountRupees The total job amount in rupees
     * @param workerProfile The worker's profile for tier-based commission
     * @return TransactionBreakdown with all amounts in micros
     */
    fun calculateBreakdown(
        grossAmountRupees: Double,
        workerProfile: WorkerProfile
    ): TransactionBreakdown {
        val grossMicros = (grossAmountRupees * MICROS_PER_RUPEE).toLong()
        val commissionRate = workerProfile.getCommissionRate()

        val commissionMicros = (grossMicros * commissionRate).toLong()
        val gstMicros = (commissionMicros * GST_RATE).toLong()
        val tdsMicros = (grossMicros * TDS_RATE).toLong()

        val payoutMicros = grossMicros - commissionMicros - gstMicros - tdsMicros

        return TransactionBreakdown(
            grossAmountMicros = grossMicros,
            commissionMicros = commissionMicros,
            commissionRate = commissionRate,
            gstOnCommissionMicros = gstMicros,
            tdsDeductionMicros = tdsMicros,
            workerPayoutMicros = payoutMicros
        )
    }

    /**
     * Calculate the "negative wallet" debit for cash jobs
     * When worker collects cash, platform records a debit for its share
     * 
     * @param grossAmountRupees The total job amount collected as cash
     * @param workerProfile The worker's profile
     * @return Debit amount in micros (commission + GST + TDS)
     */
    fun calculateCashJobDebit(
        grossAmountRupees: Double,
        workerProfile: WorkerProfile
    ): Long {
        val breakdown = calculateBreakdown(grossAmountRupees, workerProfile)
        return breakdown.commissionMicros + 
               breakdown.gstOnCommissionMicros + 
               breakdown.tdsDeductionMicros
    }

    /**
     * Convert micros to rupees for display
     */
    fun microsToRupees(micros: Long): Double = micros.toDouble() / MICROS_PER_RUPEE

    /**
     * Format amount as Indian Rupee string
     */
    fun formatRupees(rupees: Double): String = "₹${String.format("%.2f", rupees)}"
}
