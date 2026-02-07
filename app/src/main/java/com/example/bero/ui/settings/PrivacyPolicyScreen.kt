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
fun PrivacyPolicyScreen(
    onBackClick: () -> Unit = {}
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Privacy Policy", fontWeight = FontWeight.Bold) },
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
                text = "Privacy Policy",
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
            
            PolicySection(
                title = "1. Introduction",
                content = """
Welcome to Bero ("we," "our," or "us"). We are committed to protecting your personal information and your right to privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our mobile application.

By using Bero, you agree to the collection and use of information in accordance with this policy.
                """.trimIndent()
            )
            
            PolicySection(
                title = "2. Information We Collect",
                content = """
We collect information you provide directly to us, including:

• Personal Information: Name, phone number, email address, profile photo, and address.
• Identity Verification: Government-issued ID documents for KYC verification.
• Payment Information: Transaction history and payment method details.
• Location Data: Your device's location when you use our services.
• Device Information: Device type, operating system, and unique device identifiers.
• Usage Data: How you interact with our app, including pages visited and features used.
                """.trimIndent()
            )
            
            PolicySection(
                title = "3. How We Use Your Information",
                content = """
We use the information we collect to:

• Provide, maintain, and improve our services
• Process transactions and send related information
• Connect workers with clients seeking services
• Verify user identity and prevent fraud
• Send notifications about bookings, payments, and updates
• Respond to your comments, questions, and requests
• Monitor and analyze trends, usage, and activities
• Personalize and improve your experience
                """.trimIndent()
            )
            
            PolicySection(
                title = "4. Information Sharing",
                content = """
We may share your information in the following situations:

• With Service Providers: Third-party vendors who assist in providing our services.
• Between Users: Workers and clients can view each other's profiles and contact information for service coordination.
• For Legal Purposes: When required by law or to protect our rights.
• Business Transfers: In connection with any merger, sale, or acquisition.
• With Your Consent: When you give us permission to share your information.
                """.trimIndent()
            )
            
            PolicySection(
                title = "5. Data Security",
                content = """
We implement appropriate technical and organizational security measures to protect your personal information, including:

• Encryption of data in transit and at rest
• Regular security assessments and updates
• Access controls and authentication mechanisms
• Secure data storage practices

However, no method of transmission over the Internet is 100% secure, and we cannot guarantee absolute security.
                """.trimIndent()
            )
            
            PolicySection(
                title = "6. Your Rights",
                content = """
You have the right to:

• Access your personal information
• Correct inaccurate or incomplete information
• Delete your account and personal data
• Opt-out of marketing communications
• Withdraw consent for data processing
• Request data portability

To exercise these rights, please contact us through the app's Help & Support section.
                """.trimIndent()
            )
            
            PolicySection(
                title = "7. Third-Party Services",
                content = """
Our app uses third-party services that may collect information, including:

• Phone number verification services (Truecaller)
• Payment processing services
• Analytics and crash reporting services
• Cloud storage providers

These services have their own privacy policies governing the use of your information.
                """.trimIndent()
            )
            
            PolicySection(
                title = "8. Children's Privacy",
                content = """
Our services are not intended for users under 18 years of age. We do not knowingly collect personal information from children. If we discover that a child has provided us with personal information, we will delete it immediately.
                """.trimIndent()
            )
            
            PolicySection(
                title = "9. Changes to This Policy",
                content = """
We may update this Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the "Last updated" date. Your continued use of the app after changes constitutes acceptance of the updated policy.
                """.trimIndent()
            )
            
            PolicySection(
                title = "10. Contact Us",
                content = """
If you have questions about this Privacy Policy or our privacy practices, please contact us at:

Email: privacy@beroapp.com
Address: [Your Company Address]

You may also reach us through the Help & Support section in the app.
                """.trimIndent()
            )
            
            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
private fun PolicySection(
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
