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

## 8. Gaps & Future Work (TMF Alignment)
This section details the gaps between the current implementation and the TMF632 standard, including specific use cases, justification, and technical implementation strategies.

### 8.1 Missing Features from Analysis

#### 8.1.1 External References
*   **Context**: Use Case 7 mentions linking to external systems (Legacy CRM, Identity Providers), but the domain model currently lacks this capability.
*   **Status**: ✅ Backend Implemented | ⚠️ UI Partial (Display Only)

*   **Use Case**:
    *   **Legacy Data Import**: An acquired ISP has customers with IDs like `LEGACY_001`. Support agents need to search by this old ID to find the new UUID-based record.
    *   **SSO Integration**: An identity provider (Auth0/Keycloak) holds the subject ID (`auth0|xyz`). This needs to be mapped to the `Party.ID` for authentication resolution.
*   **Implementation Strategy**:
    1.  **Domain Model**: Add a `ExternalReference` struct.
        ```go
        type ExternalReference struct {
            ID                string `json:"id"`
            PartyID           string `json:"partyId"`
            ExternalSystemID  string `json:"externalSystemId"` // e.g. "LegacyCRM", "Auth0"
            ExternalReference string `json:"externalReference"` // e.g. "LEGACY_001"
        }
        ```
    2.  **Schema**: Create table `external_references` with a composite index on `(external_system_id, external_reference)`.
    3.  **API**: Update `SearchParties` to accept an `externalReference` filter which joins this table.

#### 8.1.2 Tax Exemptions (Party Level)
*   **Context**: Use Case 6 identifies Tax Exemption as a requirement. Currently, this exists only in the **Customer** service, which is structurally incorrect for TMF alignment.
*   **Status**: ✅ Backend Implemented | ✅ UI Implemented

*   **Use Case**:
    *   **Charitable Organization**: A registered non-profit is VAT-exempt for *all* purchases and services. This is a property of the *Organization*, not just a specific customer account.
    *   **Diplomatic Immunity**: A diplomat (Individual) has tax-exempt status that applies universally.
*   **Implementation Strategy**:
    1.  **Domain Model**: Port the `TaxExemption` struct (currently in Customer) to `party.go`.
        ```go
        type TaxExemption struct {
            ID                  string    `json:"id"`
            PartyID             string    `json:"partyId"`
            CertificateNumber   string    `json:"certificateNumber"`
            IssuingJurisdiction string    `json:"issuingJurisdiction"`
            ValidFor            TimePeriod `json:"validFor"`
        }
        ```
    2.  **Migration**: Create `party_tax_exemptions` table. In the future, the Customer service should look up tax exemptions from the Party service via RPC during billing calculation.

### 8.2 Standard TMF632 Gaps

#### 8.2.1 Identification Attachments (Files)
*   **Context**: TMF632 includes an `Attachment` resource. Currently, we only store metadata (ID numbers).
*   **Status**: ✅ Backend Implemented | ⚠️ UI Partial (Display Only)

*   **Use Case**:
    *   **KYC Compliance**: Exploring a prepaid SIM requires uploading a scan of a Passport.
    *   **Audit Proof**: An organization must provide a PDF of their Trade License.
*   **Implementation Strategy**:
    1.  **Storage**: Define a `FileStorage` interface to abstract blob operations. Initially, implement this using PostgreSQL (e.g., `bytea` columns), ensuring the architecture allows for a seamless transition to S3/MinIO in the future.
    2.  **Domain Model**:
        ```go
        type Attachment struct {
            ID       string `json:"id"`
            OwnerID  string `json:"ownerId"` // PartyID
            MimeType string `json:"mimeType"`
            Name     string `json:"name"`
            URL      string `json:"url"` // Presigned URL or proxy path
            Type     string `json:"type"` // e.g., "Scan", "Form"
        }
        ```
    3.  **Security**: The `URL` returned in API responses should be a temporary pre-signed URL (valid for 15m) to prevent unauthorized public access.

#### 8.2.2 Granular Party Roles
*   **Context**: The current `RelatedParty` is a simple link. TMF suggests a more robust `PartyRole` pattern.
*   **Status**: ✅ Backend Implemented | ⚠️ UI Partial (Display Only)

*   **Use Case**:
    *   **B2B Permissions**: An Employee (Individual) is related to their Company (Organization). We need to know if they are a "Billing Admin" (can pay bills) or a "Technical Contact" (can only open tickets).
*   **Implementation Strategy**:
    1.  **Enhanced Relationship**: Add metadata to `RelatedParty`.
        ```go
        type RelatedParty struct {
            // ... existing fields
            Role string `json:"role"` // e.g. "Employee", "Guardian"
            Permissions []string `json:"permissions"` // e.g. ["ManageBilling", "ViewService"]
        }
        ```
    2.  **Validation Logic**: Implement checking rules (e.g., "A Guardian must be >18 years old").

#### 8.2.3 Financial Profiles (Scope Clarification)
*   **Context**: Requirement to store Bank Accounts or Credit Cards.
*   **Analysis**:
    *   **TMF632 Party Management** deals with *Identity*. It does not typically store financial instruments due to PCI-DSS scope and separation of concerns.
    *   **TMF670 Payment Method Management** is the dedicated API for managing Credit Cards, Direct Debits, and Digital Wallets.
    *   **TMF666 Account Management** links these methods to a billing relationship.
*   **Decision**:
    *   We will **NOT** implement `payment_methods` in the Party Service.
    *   **Future Work**: Implement **TMF670** as a separate microservice.
    *   For the MVP, if simple storage is needed, it should reside in the **Customer** service (as `CustomerAccount.PaymentMethod`) or a dedicated Payment Service, effectively treating "Financial Profile" as a non-Party concern.


