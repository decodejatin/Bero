package com.example.bero.ui.job

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostJobScreen() {
    var jobTitle by remember { mutableStateOf("") }
    var jobDescription by remember { mutableStateOf("") }
    var budget by remember { mutableStateOf("") }
    var isPosted by remember { mutableStateOf(false) }

    if (isPosted) {
        Column(
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text("✅", fontSize = 64.sp)
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                "Job Posted Successfully!",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text("Workers nearby will see your job soon.")
            Spacer(modifier = Modifier.height(24.dp))
            Button(onClick = { 
                isPosted = false 
                jobTitle = ""
                jobDescription = ""
                budget = ""
            }) {
                Text("Post Another Job")
            }
        }
        return
    }

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(onClick = { isPosted = true }) {
                Icon(Icons.Default.Add, "Post Job")
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Text(
                text = "Post a New Job",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )
            
            OutlinedTextField(
                value = jobTitle,
                onValueChange = { jobTitle = it },
                label = { Text("Job Title (e.g. Broken Pipe)") },
                modifier = Modifier.fillMaxWidth()
            )
            
            OutlinedTextField(
                value = jobDescription,
                onValueChange = { jobDescription = it },
                label = { Text("Description") },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(120.dp),
                maxLines = 5
            )
            
            OutlinedTextField(
                value = budget,
                onValueChange = { budget = it },
                label = { Text("Budget (₹)") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                modifier = Modifier.fillMaxWidth()
            )
            
            Spacer(modifier = Modifier.weight(1f))
            
            Button(
                onClick = { isPosted = true },
                modifier = Modifier.fillMaxWidth().height(50.dp)
            ) {
                Text("Post Job")
            }
        }
    }
}
