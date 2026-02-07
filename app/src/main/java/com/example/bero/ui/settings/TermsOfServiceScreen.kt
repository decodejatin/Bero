@file:OptIn(ExperimentalMaterial3Api::class)

package com.example.bero.ui.settings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun TermsOfServiceScreen(
    onBackClick: () -> Unit = {}
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Terms of Service", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState())
                .padding(16.dp)
        ) {
            Text(
                text = "Terms of Service",
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            
            Text(
                text = "Last updated: February 2026",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                modifier = Modifier.padding(bottom = 24.dp)
            )
            
            TermsSection(
                title = "1. Acceptance of Terms",
                content = """
By downloading, installing, or using the Bero mobile application ("App"), you agree to be bound by these Terms of Service ("Terms"). If you do not agree to these Terms, do not use the App.

These Terms constitute a legally binding agreement between you and Bero ("Company," "we," "us," or "our").
                """.trimIndent()
            )
            
            TermsSection(
                title = "2. Description of Service",
                content = """
Bero is a platform that connects service workers (plumbers, electricians, carpenters, cleaners, etc.) with clients seeking their services. The App facilitates:

• Job posting and discovery
• Worker-client communication
• Booking and scheduling
• Payment processing
• Reviews and ratings

We act solely as an intermediary platform and are not responsible for the quality of services provided by workers.
                """.trimIndent()
            )
            
            TermsSection(
                title = "3. User Accounts",
                content = """
To use Bero, you must:

• Be at least 18 years old
• Register with a valid phone number
• Provide accurate and complete information
• Maintain the security of your account credentials
• Notify us immediately of any unauthorized use

You are responsible for all activities that occur under your account. We reserve the right to suspend or terminate accounts that violate these Terms.
                """.trimIndent()
            )
            
            TermsSection(
                title = "4. User Types",
                content = """
Workers: Individuals or businesses offering professional services through the App. Workers must:
• Provide accurate information about their skills and qualifications
• Complete identity verification (KYC) when required
• Maintain appropriate licenses and insurance as required by law
• Deliver services professionally and as described

Clients: Individuals seeking to hire workers for services. Clients must:
• Provide accurate job descriptions and requirements
• Treat workers with respect
• Pay for services as agreed
• Provide a safe working environment
                """.trimIndent()
            )
            
            TermsSection(
                title = "5. Payments and Fees",
                content = """
• Payment for services is processed through the App
• Workers agree to pay applicable platform fees
• Clients agree to pay the quoted price plus any applicable taxes
• Refunds are subject to our Refund Policy
• We reserve the right to modify fees with notice

All prices are in Indian Rupees (INR) unless otherwise specified.
                """.trimIndent()
            )
            
            TermsSection(
                title = "6. Prohibited Conduct",
                content = """
Users may not:

• Violate any applicable laws or regulations
• Provide false or misleading information
• Harass, abuse, or harm other users
• Circumvent the platform for payments
• Use the App for fraudulent purposes
• Share account credentials with others
• Engage in any discriminatory behavior
• Post inappropriate or offensive content
• Spam or send unsolicited communications
• Attempt to reverse engineer or hack the App
                """.trimIndent()
            )
            
            TermsSection(
                title = "7. Intellectual Property",
                content = """
All content, features, and functionality of the App are owned by Bero and are protected by intellectual property laws. You may not:

• Copy, modify, or distribute our content
• Use our trademarks without permission
• Create derivative works based on the App
• Remove any copyright or proprietary notices
                """.trimIndent()
            )
            
            TermsSection(
                title = "8. Disclaimers",
                content = """
THE APP IS PROVIDED "AS IS" WITHOUT WARRANTIES OF ANY KIND. WE DISCLAIM ALL WARRANTIES, EXPRESS OR IMPLIED, INCLUDING:

• Merchantability and fitness for a particular purpose
• Accuracy or reliability of information
• Uninterrupted or error-free service
• Quality of services provided by workers

We are not liable for disputes between workers and clients.
                """.trimIndent()
            )
            
            TermsSection(
                title = "9. Limitation of Liability",
                content = """
TO THE MAXIMUM EXTENT PERMITTED BY LAW, BERO SHALL NOT BE LIABLE FOR:

• Indirect, incidental, or consequential damages
• Loss of profits, data, or goodwill
• Personal injury or property damage
• Actions or omissions of third parties
• Unauthorized access to your account

Our total liability shall not exceed the fees paid by you in the preceding 12 months.
                """.trimIndent()
            )
            
            TermsSection(
                title = "10. Indemnification",
                content = """
You agree to indemnify and hold harmless Bero, its affiliates, and their respective officers, directors, employees, and agents from any claims, damages, losses, or expenses arising from:

• Your use of the App
• Your violation of these Terms
• Your violation of any third-party rights
• Services provided or received through the App
                """.trimIndent()
            )
            
            TermsSection(
                title = "11. Modifications to Terms",
                content = """
We reserve the right to modify these Terms at any time. Changes will be effective upon posting to the App. Your continued use of the App after changes constitutes acceptance of the modified Terms.

We will notify users of material changes through the App or via email.
                """.trimIndent()
            )
            
            TermsSection(
                title = "12. Termination",
                content = """
We may terminate or suspend your access to the App at any time, without prior notice, for:

• Violation of these Terms
• Fraudulent or illegal activity
• Harm to other users or the platform
• Any other reason at our sole discretion

Upon termination, your right to use the App ceases immediately.
                """.trimIndent()
            )
            
            TermsSection(
                title = "13. Governing Law",
                content = """
These Terms shall be governed by and construed in accordance with the laws of India. Any disputes arising from these Terms shall be subject to the exclusive jurisdiction of the courts in [Your City], India.
                """.trimIndent()
            )
            
            TermsSection(
                title = "14. Contact Information",
                content = """
For questions about these Terms of Service, please contact us at:

Email: legal@beroapp.com
Address: [Your Company Address]

You may also reach us through the Help & Support section in the App.
                """.trimIndent()
            )
            
            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
private fun TermsSection(
    title: String,
    content: String
) {
    Column(modifier = Modifier.padding(bottom = 24.dp)) {
        Text(
            text = title,
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        Text(
            text = content,
            style = MaterialTheme.typography.bodyMedium,
            lineHeight = 24.sp,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
        )
    }
}
