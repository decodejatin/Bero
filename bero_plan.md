# **Project Bero: Technical Specification & Master Plan**

**Version:** 1.0

**Target Market:** India (Tier 1, 2, & 3 Cities)

**Core Philosophy:** "Zero Friction for Workers, Maximum Trust for Clients."

## ---

**1\. Executive Summary**

**Bero** is a hyper-local, on-demand marketplace for blue-collar services (plumbers, electricians, cleaners, helpers). Unlike western counterparts, Bero is engineered for the unique constraints of the Indian market:

* **Low Literacy:** Interfaces designed for voice and visuals, not heavy text.  
* **Trust Deficit:** Solved via Aadhaar/Digilocker verification and video bios.  
* **Financial Compliance:** Automated handling of TDS u/s 194-O and GST.  
* **Network Conditions:** "Offline-First" architecture for spotty 4G/5G coverage.

## ---

**2\. Technology Stack (The "No Regret" Choice)**

We are choosing a stack optimized for high concurrency (millions of users) and long-term maintainability.

| Component | Technology | Reasoning |
| :---- | :---- | :---- |
| **Mobile (Android/iOS)** | **Kotlin Multiplatform (KMP)** | **Critical.** Allows sharing 100% of business logic (Auth, Networking, Location Sync, Payment Calcs) between Android & iOS while keeping the UI Native (Compose/SwiftUI) for maximum performance on low-end devices. |
| **Backend API** | **Go (Golang)** | Best-in-class for handling thousands of concurrent location streams/websockets. Compiles to a single binary. |
| **Database (OLTP)** | **CockroachDB** | Distributed SQL. Resilient to node failures. Ensures "Ledger" accuracy (money never disappears). |
| **Caching / Hot State** | **Redis Cluster** | Stores active driver locations and session data for sub-millisecond access. |
| **Geospatial Index** | **H3 (Uber)** | Hexagonal grid system. Superior to GPS radius for calculating surge pricing and supply/demand in dense Indian neighborhoods. |
| **Payments** | **Cashfree / Razorpay Route** | specifically for **Split Payments** (Client pays ₹500 → System auto-splits ₹450 to Worker, ₹50 to Platform). |
| **Auth** | **Truecaller SDK** | One-tap login. Reduces OTP friction. |
| **Maps** | **Mapbox SDK** | Cheaper and more customizable than Google Maps for custom styling (e.g., highlighting landmarks). |

## ---

**3\. Financial Architecture (India Compliance)**

### **3.1 The Commission Model & TDS**

**Regulation:** Under Section 194-O of the Income Tax Act, e-commerce operators must deduct **1% TDS** on the *gross* amount of sale if the worker earns \> ₹5 Lakh/year (or generically to stay safe).

**Transaction Flow (Example: ₹1,000 Job)**

1. **Job Fee:** ₹1,000  
2. **Platform Commission (15%):** ₹150  
3. **GST on Commission (18%):** ₹27 (18% of ₹150)  
4. **TDS Deduction (1% of Gross):** ₹10 (1% of ₹1,000)  
5. **Worker Payout:** ₹1,000 \- ₹150 \- ₹27 \- ₹10 \= **₹813**

### **3.2 The "Negative Wallet" (Cash Jobs)**

Since many Indian customers prefer Cash (UPI QR on spot):

1. Worker collects ₹1,000 Cash.  
2. Bero Wallet records a **Debit** of ₹187 (Commission \+ Taxes).  
3. **Blocking Logic:** If Wallet Balance \< \-₹500, Worker cannot accept new jobs until they "Recharge" the app via UPI.

## ---

**4\. Feature Specification: The "Bero" Difference**

### **4.1 The "Vernacular" Interface (Worker App)**

* **Audio-First:** Job offers are read aloud: *"New Paint Job. Sector 4\. 800 Rupees."*  
* **Video Bio:** Workers record a 15-second "Selfie Video" introducing themselves. This builds massive trust with clients compared to a static photo.  
* **Visual Navigation:** Instead of just street names, use photos of landmarks ("Turn Left at Blue Water Tank").

### **4.2 The "Streak" System (Retention)**

**Goal:** Prevent workers from switching to other apps.

* **Logic:** Work 7 days consecutively \= **Silver Tier** (Commision drops to 12%).  
* **Logic:** Work 30 days consecutively \= **Gold Tier** (Commission drops to 10% \+ Free Accident Insurance).  
* **Streak Freeze:** Workers can buy a "Freeze" power-up using earned coins to save their streak if they take a holiday.

## ---

**5\. Database Schema (PostgreSQL / CockroachDB)**

### **5.1 users**

Core identity table.

SQL

CREATE TABLE users (  
    id UUID PRIMARY KEY DEFAULT gen\_random\_uuid(),  
    phone\_number VARCHAR(15) UNIQUE NOT NULL,  
    full\_name VARCHAR(100),  
    aadhaar\_kyc\_status VARCHAR(20) DEFAULT 'PENDING', \-- NONE, PENDING, VERIFIED, REJECTED  
    user\_type VARCHAR(10) NOT NULL, \-- 'WORKER' or 'CLIENT'  
    created\_at TIMESTAMPTZ DEFAULT NOW()  
);

### **5.2 worker\_profiles**

The supply-side attributes.

SQL

CREATE TABLE worker\_profiles (  
    user\_id UUID REFERENCES users(id),  
    skills TEXT, \-- e.g.  
    h3\_index\_res9 VARCHAR(15), \-- Current hex location  
    is\_online BOOLEAN DEFAULT FALSE,  
    wallet\_balance\_micros BIGINT DEFAULT 0, \-- Stored in smallest currency unit  
    rating\_avg DECIMAL(3, 2),  
    streak\_count INT DEFAULT 0,  
    last\_active\_date DATE  
);

### **5.3 ledger\_transactions**

**Double-Entry Accounting System.** Never delete money, only move it.

SQL

CREATE TABLE ledger\_transactions (  
    id UUID PRIMARY KEY DEFAULT gen\_random\_uuid(),  
    job\_id UUID, \-- Nullable (if it's a wallet recharge)  
    debit\_account\_id UUID NOT NULL, \-- Who loses money  
    credit\_account\_id UUID NOT NULL, \-- Who gains money  
    amount\_micros BIGINT NOT NULL,  
    transaction\_type VARCHAR(50), \-- 'COMMISSION', 'TDS\_DEDUCTION', 'JOB\_PAYOUT', 'GST\_COLLECTION'  
    created\_at TIMESTAMPTZ DEFAULT NOW()  
);

## ---

**6\. Project Roadmap**

### **Phase 0: Setup (Weeks 1-2)**

* \[ \] Initialize Git Repo (bero-monorepo).  
* \[ \] Setup KMP (Kotlin Multiplatform) Shared Module.  
* \[ \] Configure CI/CD (GitHub Actions).

### **Phase 1: The "Supply" (Weeks 3-8)**

* \[ \] Build Android Worker App.  
* \[ \] Implement Truecaller Auth.  
* \[ \] Implement Aadhaar KYC (using a vendor API like HyperVerge).  
* \[ \] **Milestone:** Onboard 50 workers manually in 1 neighborhood.

### **Phase 2: The "Demand" (Weeks 9-14)**

* \[ \] Build Client App (Android).  
* \[ \] Implement Mapbox "Ghost Cars" (Show fake/real workers on map).  
* \[ \] Implement Razorpay Split Payments.  
* \[ \] **Milestone:** First Real Cash Transaction.

### **Phase 3: The "Brain" (Weeks 15+)**

* \[ \] Implement Go Backend "Matchmaker" (H3 Logic).  
* \[ \] Turn on "Streak" Gamification.  
* \[ \] Launch iOS Client App (using Shared KMP logic).

## ---

**7\. Folder Structure (Monorepo)**

bero/

├── androidApp/ \# Native Android UI (Jetpack Compose)

├── iosApp/ \# Native iOS UI (SwiftUI)

├── shared/ \# KMP Module (The "Brain")

│ ├── src/commonMain/kotlin/com/bero/

│ │ ├── auth/ \# Login Logic

│ │ ├── domain/ \# UseCases & Models

│ │ ├── network/ \# Ktor / gRPC clients

│ │ └── payment/ \# Tax & Commission Math

├── backend/ \# Go (Golang) Services

│ ├── cmd/server/

│ ├── internal/matchmaker/

│ └── internal/ledger/

├── infrastructure/ \# Terraform / Docker Compose

└── README.md