package com.bero.auth

/**
 * JVM implementation of currentTimeMillis for tests
 */
actual fun currentTimeMillis(): Long = System.currentTimeMillis()
