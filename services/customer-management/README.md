# Customer Management Service (TMF629)

## Core Concept: Customer vs. Party
It is crucial to distinguish between a **Party** and a **Customer**:
- **Party (TMF632)**: An individual or an organization. It represents "who" they are (name, legal ID, contact details).
- **Customer (TMF629)**: A role played by a Party in the context of buying products/services. It represents "how" they interact with the business (customer status, credit profile, segment).

*A Party can exist without being a Customer, but a Customer must be linked to a Party.*

## Use Cases

### 1. Onboard Customer
**Goal**: Register a new customer in the system.
- **Pre-condition**: The underlying Party (Individual or Organization) should ideally exist, or can be created during this flow (though separation of concerns suggests Party should typically be managed first).
- **Input**: Party Reference, Customer Segment (e.g., B2C, B2B), Initial Status.
- **Output**: A new Customer entity with a unique ID.

### 2. Retrieve Customer
**Goal**: View customer details.
- **Input**: Customer ID.
- **Output**: Full customer profile including status, credit score (if applicable), and reference to the associated Party.

### 3. Update Customer
**Goal**: Manage the customer lifecycle.
- **Scenarios**:
    - Update Customer Status (e.g., Active -> Suspended).
    - Change Customer Segment.
    - Update Credit Profile.
- **Note**: Changes to name or contact info should be done via Party Management, not here.

### 4. Find Customers
**Goal**: Search for customers based on criteria.
- **Criteria**: Name (via Party), ID, Status.

## TMF Alignment
This service implements **TMF629 Customer Management**.
- **Key Resources**: `Customer`
- **Related Resources**: `CustomerAccount` (potentially out of scope for initial order creation, but relevant for billing).
