# Customer Management (TMF629) Analysis

## Overview
Customer Management (TMF629) provides a standardized mechanism for managing the lifecycle of a **Customer**. While Party Management deals with identity (who someone is), Customer Management deals with the **business relationship** (what they do with the enterprise).

## Core Entities
1.  **Customer**: A party (Individual or Organization) that has a business relationship with the enterprise. 
    *   **Crucial Mapping**: Every Customer record MUST reference a `Party` from the TMF632 service.
2.  **Customer Account**: Represents the financial/billing relationship. A customer can have multiple accounts.
3.  **Customer Credit Profile**: Information regarding the customer's creditworthiness.
4.  **Contact Medium**: Customer-specific contact information.

## Key Use Cases

### 1. Customer Lifecycle Management
*   **Onboard Customer**: Create a new customer record linked to an existing Party.
*   **Update Profile**: Manage customer-specific attributes (preferences, description, status).
*   **Customer Suspension/Termination**: Handle lifecycle states (Active, Suspended, Closed).
*   **Retrieve Customer**: Fetch a unified view of the customer, including references to their Party details.

### 2. Credit Profile Management
*   Manage credit ratings and scores.
*   Automate status changes based on credit profile updates (e.g., suspending customers with poor credit).

### 3. Customer Account Management
*   Create and link accounts for billing and financial management.
*   Manage account status and validity periods.

### 4. Relationship Management
*   Manage `RelatedParty` associations within the customer context (e.g., who is the interlocutor for this customer).

### 5. Event Notifications
*   Broadcast events for customer creation, status changes, and profile updates to downstream services (Billing, Ordering, Assurance).

### 6. Party Deletion Validation (Saga Participant)
The Customer Management service acts as a validator for Party Deletion requests.

**Validation Logic:**
1.  **Listen**: Subscribes to `evt.party.deletion_initiated` from Party Management.
2.  **Validate**: Checks if any *Active* linked customers exist for the given `partyId`.
3.  **Respond**:
    *   **If Linked**: Publishes `cmd.party.cancel_deletion` to abort the process.
    *   **If Not Linked**: Publishes `cmd.party.finalize_deletion` to approve the process.


## TMF629 vs TMF632 (Integration Points)
*   **Customer -> Party**: A Customer *is* a Party playing a specific role.
*   **Service Dependency**: The Customer service depends on the Party service for core identity data.
*   **Consistency**: When a Party is updated in TMF632, the Customer service might need to react if those changes affect the customer relationship.

## Implementation Recommendations
*   **Data Isolation**: Store only customer-specific data (status, specific contact preferences, account links). Do not duplicate core party data (given name, trading name) but reference it via `partyId`.
*   **Distributed Queries**: The `query.customer.get` operation should ideally join data by querying the Party Management service asynchronously via RPC.


## 6. Gaps & Future Work (Development Specifications)

This section details the technical specifications required to close the gaps between the current implementation and the TMF629 standard. These specifications are designed for direct developer implementation, including Database Schemas (PostgreSQL), Domain Models (Go), and API Contract updates (RabbitMQ/JSON).

### 6.1 Related Parties (Hierarchies & Relationships)

**Use Case**:
Support for B2B hierarchies (Parent/Child companies) and delegated authority (e.g., "Authorized User" on an account). This allows a single "Parent" customer to view and manage billing for multiple "Child" customers.

**Status**: ✅ Backend Implemented | ⚠️ UI Partial (Onboarding Only)


**Technical Specification**:

#### 1. Database Schema
Create a new table `related_parties` to link customers to other parties.

```sql
CREATE TABLE related_parties (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    related_party_id UUID NOT NULL, -- The target Party/Customer ID
    role VARCHAR(50) NOT NULL,      -- e.g., "Parent", "Child", "BillReceiver"
    name VARCHAR(255) NOT NULL,     -- Snapshot of the related party's name
    valid_for_start TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_for_end TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_related_parties_customer_id ON related_parties(customer_id);
```

#### 2. Domain Model (`domain/customer.go`)
```go
type RelatedParty struct {
    ID             string     `gorm:"primaryKey" json:"id"`
    CustomerID     string     `gorm:"not null;index" json:"customerId"`
    RelatedPartyID string     `json:"relatedPartyId"`
    Role           string     `json:"role"`
    Name           string     `json:"name"`
    ValidForStart  time.Time  `json:"validForStart"`
    ValidForEnd    *time.Time `json:"validForEnd,omitempty"`
}
```

#### 3. API Contract (`transport/rabbitmq/handlers.go`)
Update `OnboardCustomerPayload` and `UpdateCustomerPayload` to include `RelatedParties`.

```go
type RelatedPartyDTO struct {
    RelatedPartyID string `json:"relatedPartyId"`
    Role           string `json:"role"`
    Name           string `json:"name"`
}

// Add to OnboardCustomerPayload
RelatedParties []RelatedPartyDTO `json:"relatedParties"`
```

### 6.2 Payment Methods

**Use Case**:
Enable customers to store and manage payment instruments (Credit Cards, Direct Debit) for recurring billing. **Crucial**: We do NOT store sensitive PAN data. We store tokens returned by a Payment Gateway.

**Status**: ✅ Backend Implemented | ⚠️ UI Partial (Onboarding Only)


**Technical Specification**:

#### 1. Database Schema
```sql
CREATE TABLE payment_methods (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,  -- "CreditCard", "BankTransfer", "DigitalWallet"
    token VARCHAR(255) NOT NULL, -- The Gateway Token (e.g., "pm_12345")
    details JSONB,               -- Non-sensitive info (last4, brand, expiry)
    is_default BOOLEAN DEFAULT FALSE,
    valid_for_start TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_for_end TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### 2. Domain Model
```go
type PaymentMethod struct {
    ID         string          `gorm:"primaryKey" json:"id"`
    CustomerID string          `gorm:"not null;index" json:"customerId"`
    Type       string          `json:"type"`
    Token      string          `json:"-"` // Never expose via default JSON
    Details    json.RawMessage `gorm:"type:jsonb" json:"details"`
    IsDefault  bool            `json:"isDefault"`
}
```

#### 3. API Contract
New RabbitMQ Commands: `cmd.customer.payment-method.add`, `cmd.customer.payment-method.remove`.

```go
type AddPaymentMethodPayload struct {
    CustomerID string `json:"customerId"`
    Type       string `json:"type"`
    Token      string `json:"token"`
    Details    string `json:"details"` // JSON string
    IsDefault  bool   `json:"isDefault"`
}
```

### 6.3 Billing Configuration

**Use Case**:
Allow granular control over how a customer is billed. This includes the format of the bill (PDF/Email) and the cycle (e.g., 1st of month vs 15th).

**Status**: ✅ Backend Implemented | ✅ UI Implemented (Onboarding & Edit via Accounts)


**Technical Specification**:

#### 1. Database Schema
Extend the `customer_accounts` table. These fields belong to the financial relationship.

```sql
ALTER TABLE customer_accounts 
ADD COLUMN bill_format VARCHAR(50) DEFAULT 'PDF',
ADD COLUMN billing_cycle VARCHAR(50) DEFAULT 'Monthly';
```

#### 2. Domain Model
Update `CustomerAccount` struct in `domain/customer.go`.

```go
type CustomerAccount struct {
    // ... existing fields ...
    BillFormat   string `json:"billFormat"`   // Enum: "PDF", "Electronic", "Paper"
    BillingCycle string `json:"billingCycle"` // Enum: "Monthly", "Quarterly", "Annually"
}
```

#### 3. BFF/Transport Logic
Ensure these fields are validated against a strict Enum list in `handlers.go` before persisting.

### 6.4 Market Segment

**Use Case**:
Categorize customers for marketing and reporting (e.g., "SME", "Enterprise", "Residential"). This is often a derived field or manually assigned by Sales.

**Status**: ✅ Backend Implemented | ⚠️ UI Partial (Onboarding Only)


**Technical Specification**:

#### 1. Database Schema
New table `market_segments` to allow multiple segments per customer.

```sql
CREATE TABLE market_segments (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- "VIP", "Student", "SME"
    category VARCHAR(100)       -- "SalesChannel", "Demographic"
);
```

#### 2. Domain Model
```go
type MarketSegment struct {
    ID         string `gorm:"primaryKey" json:"id"`
    CustomerID string `gorm:"not null;index" json:"customerId"`
    Name       string `json:"name"`
    Category   string `json:"category"`
}
```

#### 3. API Contract
Add `MarketSegments` array to `OnboardCustomerPayload`.

### 6.5 Customer Interactions

**Use Case**:
Log every interaction (Call, Email, Ticket) to provide a history.

**Status**: ✅ Backend Implemented | ❌ UI Not Implemented


**Technical Specification**:

#### 1. Architecture Decision
**Note**: Ideally, this should be a separate microservice (`interaction-management`). However, for the MVP, we can implement a lightweight log within this service.

#### 2. Database Schema
```sql
CREATE TABLE customer_interactions (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    interaction_date TIMESTAMP NOT NULL,
    channel VARCHAR(50) NOT NULL, -- "Phone", "Email", "InPerson"
    type VARCHAR(50) NOT NULL,    -- "Inquiry", "Complaint", "Sales"
    description TEXT,
    agent_id VARCHAR(100),        -- Reference to the employee handling it
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### 3. Domain Model
```go
type CustomerInteraction struct {
    ID              string    `json:"id"`
    CustomerID      string    `json:"customerId"`
    InteractionDate time.Time `json:"interactionDate"`
    Channel         string    `json:"channel"`
    Type            string    `json:"type"`
    Description     string    `json:"description"`
    AgentID         string    `json:"agentId"`
}
```
#### 4. API Contract
New Command: `cmd.customer.interaction.log`
Routing Key: `customer.interaction.log`


