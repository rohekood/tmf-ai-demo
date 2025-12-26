# Party Management (TMF632) Analysis

## Overview
Party Management (TMF632) is a fundamental microservice in the TMF architecture responsible for managing standardized information about **Individuals** and **Organizations**. It serves as the single source of truth for identity information, independent of the roles these parties might play (e.g., Customer, Supplier, Employee).

## Core Entities
The service manages the abstract concept of a **Party**, which has two concrete specializations:
1.  **Individual**: Represents a person (e.g., First Name, Last Name, Gender, Birth Date).
2.  **Organization**: Represents a company or group (e.g., Trading Name, Company Type).

## Key Use Cases

### 1. Party Lifecycle Management
*   **Create Party**: Register a new Individual or Organization in the ecosystem. This acts as the initial "identity creation" step before they become a customer.
*   **Retrieve Party**: Fetch detailed information about a specific party using their unique identifier.
*   **Update Party**: Modify attributes such as names, identifying information, or status.
*   **Delete Party**: Remove party records (critical for GDPR/Right-to-be-Forgotten compliance).

### 2. Contact Medium Management
*   Manage contact details (Email, Phone, Postal Address) associated with a Party.
*   **Note**: While *Geographic Address Management* validates addresses, Party Management stores the association of an address to a person/organization (e.g., "Home Address" vs "Billing Address").

### 3. Party Search & Discovery
*   Search for parties by various criteria:
    *   Name (e.g., `givenName`, `tradingName`)
    *   External Reference IDs (e.g., National ID, Passport Number)
    *   Contact information (e.g., "Find party with email X")

### 4. Party Identification
*   Manage official identifiers like Passport Number, Social Security Number, or Tax ID.
*   Supports multiple identifiers per party with validity periods and issuing authorities.

### 5. Related Party Management
*   Define relationships between individuals and organizations (e.g., "Employee of", "Board Member of", "Legal Representative of").
*   Establish "Next of Kin" or "Guardian" relationships between individuals.

### 6. Credit & Tax Profiling
*   **Credit Management**: Track `creditRating` and `creditLimit` for parties to support risk assessment in order management.
*   **Tax Exemption**: Manage tax exemption certificates and validity for organizations or eligible individuals.

### 7. External Reference Management
*   Link parties to external systems (e.g., Legacy CRM ID, SSO ID).
*   Essential for maintaining consistency across a brownfield landscape.

### 8. Party Deletion Saga (Safe Deletion)
In compliance with data integrity rules, a Party cannot be deleted if it is referenced by an active Customer. To ensure this without synchronous dependencies, a **Saga Pattern** is used.

**Saga Flow:**
1.  **Initiation**: `cmd.party.delete` is received.
2.  **State Transition**: Party transitions to `DeletionPending` state.
3.  **Validation Request**: `evt.party.deletion_initiated` is published.
4.  **Listen for Outcome**:
    *   **Cancel**: If `cmd.party.cancel_deletion` is received (from Customer Service), the Party reverts to `Active` state.
    *   **Finalize**: If `cmd.party.finalize_deletion` is received (from Customer Service), the Party transitions to `Deleted` state.
    *   **Race Condition Handling**: If a new Customer is created referencing this Party (`evt.customer.created`) while in `DeletionPending`, the deletion is automatically aborted and the Party reverts to `Active`.

**Note**: "Deletion" in this context is a **Soft Delete**. The record status is updated to `Deleted`, but the data remains for audit/archival purposes unless a hard delete is explicitly requested (separate GDPR flow).

## Relationship to Customer Management (TMF629)
It is crucial to distinguish between **Party** and **Customer**:
*   **Party Management** concerns "Who are they?" (Identity).
*   **Customer Management** concerns "What is our business relationship?" (Role).
*   **Use Case**: When a new customer signs up, a **Party** (Individual) is created first. Then, a **Customer** entity is created which references that Party.

## Implementation Recommendations
Per the agreed asynchronous architecture, this service will **NOT** expose REST APIs. This is a deliberate architectural decision to prioritize event-driven patterns over REST compatibility. The service uses **CQRS** (Command Query Responsibility Segregation) over a Message Broker (RabbitMQ).

### Async Commands (Write)
Operations that change state are handled via Command Messages:
*   `cmd.party.create`: Create a new Party.
*   `cmd.party.update`: Update an existing Party.
*   `cmd.party.delete`: Delete a Party.
*   `cmd.party.patch`: Partially update a Party.

### Async Events (Notifications)
Successful operations result in Event Messages broadcast to the domain:
*   `evt.party.created`: A Party was created.
*   `evt.party.updated`: A Party was updated.
*   `evt.party.deleted`: A Party was deleted.
*   `evt.party.stateChange`: A Party's lifecycle state changed.

### Async Queries (Read)
Data retrieval is handled via **RPC-style Async Request-Reply**:
*   `query.party.get`: Retrieve a Party by ID.
*   `query.party.search`: Retrieve Parties by criteria.

No synchronous HTTP endpoints will be provided.
