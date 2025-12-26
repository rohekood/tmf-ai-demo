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
