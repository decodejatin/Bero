package com.example.bero.data.payment

import com.example.bero.data.models.WorkerProfile

/**
 * Transaction breakdown for a job payment
 */
data class TransactionBreakdown(
    val grossAmountMicros: Long,
    val commissionMicros: Long,
    val commissionRate: Double,
    val gstOnCommissionMicros: Long,
    val tdsDeductionMicros: Long,
    val workerPayoutMicros: Long
)

/**
 * Tax Calculator for Indian e-commerce compliance
 */
object TaxCalculator {

    private const val GST_RATE = 0.18
    private const val TDS_RATE = 0.01
    private const val MICROS_PER_RUPEE = 100000L

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

    fun calculateCashJobDebit(
        grossAmountRupees: Double,
        workerProfile: WorkerProfile
    ): Long {
        val breakdown = calculateBreakdown(grossAmountRupees, workerProfile)
        return breakdown.commissionMicros + 
               breakdown.gstOnCommissionMicros + 
               breakdown.tdsDeductionMicros
    }

    fun microsToRupees(micros: Long): Double = micros.toDouble() / MICROS_PER_RUPEE

    fun formatRupees(rupees: Double): String = "₹${String.format("%.2f", rupees)}"
}
