# Catalog Management UI Analysis

This document analyzes the requirements for implementing Catalog Management (TMF620, TMF633, TMF634) in the Demo UI.

## 1. Overview

Catalog Management allows users to create and manage product catalogs, categories, product specifications, and product offerings. These entities are managed by the `product-catalog-management` service.

## 2. Domain Entities

| Entity | Description | TMF Specification |
|:-------|:------------|:------------------|
| **Catalog** | A collection of product offerings. | TMF620 |
| **Category** | Used to group product offerings within a catalog. | TMF620 |
| **Product Specification** | Defines the technical and commercial characteristics of a product. | TMF633 |
| **Product Offering** | A commercial entity that can be sold to customers. Links a Specification to one or more Categories. | TMF620 |

## 3. UI Requirements

### 3.1 Pages

1.  **Catalog Management Dashboard** (`/catalog`)
    - Overview of catalogs, categories, and offerings.
    - Quick links to manage different entities.

2.  **Catalog Pages**
    - **List Page** (`/catalog/catalogs`): Search and list catalogs.
    - **Detail Page** (`/catalog/catalogs/:id`): View catalog details and its categories.
    - **Create/Edit Page** (`/catalog/catalogs/new`, `/catalog/catalogs/:id/edit`): Form for catalog lifecycle.

3.  **Category Pages**
    - **List Page** (`/catalog/categories`): Search and list categories.
    - **Detail Page** (`/catalog/categories/:id`): View category details and sub-categories/offerings.
    - **Create/Edit Page** (`/catalog/categories/new`, `/catalog/categories/:id/edit`): Form for category management.

4.  **Product Specification Pages**
    - **List Page** (`/catalog/specifications`): Search and list specifications.
    - **Detail Page** (`/catalog/specifications/:id`): View spec details and characteristics.
    - **Create/Edit Page** (`/catalog/specifications/new`, `/catalog/specifications/:id/edit`): Form for specification management.

5.  **Product Offering Pages**
    - **List Page** (`/catalog/offerings`): Search and list offerings.
    - **Detail Page** (`/catalog/offerings/:id`): View offering details, prices, and attachments.
    - **Create/Edit Page** (`/catalog/offerings/new`, `/catalog/offerings/:id/edit`): Form for offering management.

### 3.2 Components

- **Characteristic Editor**: For managing product specification characteristics.
- **Price Editor**: For managing product offering prices.
- **Attachment Manager**: For managing offering attachments.
- **Category Picker**: For associating offerings with categories.
- **Specification Picker**: For associating offerings with specifications.

## 4. BFF Requirements

The BFF needs to expose the following endpoints, which map to RabbitMQ topics in the `product-catalog-management` service.

### 4.1 Catalog Endpoints
- `GET /api/catalogs` -> `query.catalog.catalog.list`
- `GET /api/catalogs/{id}` -> `query.catalog.catalog.get`
- `POST /api/catalogs` -> `cmd.catalog.catalog.create`
- `PUT /api/catalogs/{id}` -> `cmd.catalog.catalog.update`
- `DELETE /api/catalogs/{id}` -> `cmd.catalog.catalog.delete`

### 4.2 Category Endpoints
- `GET /api/categories` -> `query.catalog.category.list`
- `GET /api/categories/{id}` -> `query.catalog.category.get`
- `POST /api/categories` -> `cmd.catalog.category.create`
- `PUT /api/categories/{id}` -> `cmd.catalog.category.update`
- `DELETE /api/categories/{id}` -> `cmd.catalog.category.delete`

### 4.3 Specification Endpoints
- `GET /api/specifications` -> `query.catalog.specification.list`
- `GET /api/specifications/{id}` -> `query.catalog.specification.get`
- `POST /api/specifications` -> `cmd.catalog.specification.create`
- `PUT /api/specifications/{id}` -> `cmd.catalog.specification.update`
- `DELETE /api/specifications/{id}` -> `cmd.catalog.specification.delete`

### 4.4 Offering Endpoints
- `GET /api/offerings` -> `query.catalog.offering.list`
- `GET /api/offerings/{id}` -> `query.catalog.offering.get`
- `POST /api/offerings` -> `cmd.catalog.offering.create`
- `PUT /api/offerings/{id}` -> `cmd.catalog.offering.update`
- `DELETE /api/offerings/{id}` -> `cmd.catalog.offering.delete`

## 5. Implementation Steps

### Phase 1: BFF Integration
1.  Define catalog topics in BFF.
2.  Implement `CatalogHandler` in BFF.
3.  Register catalog routes in BFF.
4.  Add unit tests for `CatalogHandler`.

### Phase 2: UI Foundation & Specifications
1.  Create `catalog` feature directory in UI.
2.  Define TypeScript types for catalog entities.
3.  Implement API service for catalog.
4.  Implement Product Specification List and Create/Edit pages.
5.  Add navigation links.

### Phase 3: Catalogs & Categories
1.  Implement Catalog List and Create/Edit pages.
2.  Implement Category List and Create/Edit pages.
3.  Implement Category nesting visualization.

### Phase 4: Product Offerings
1.  Implement Product Offering List and Create/Edit pages.
2.  Implement Specification and Category selection for Offerings.
3.  Implement Price and Attachment management.

### Phase 5: Verification & Polish
1.  End-to-end testing of the full catalog creation flow.
2.  UI/UX refinements and responsive design checks.
