// Catalog Management Types based on TMF620, TMF633, TMF634

export type LifecycleStatus = 'Draft' | 'Active' | 'Retired' | 'Suspended';

export interface TimePeriod {
    startDateTime?: string;
    endDateTime?: string;
}

export interface Money {
    unit: string;
    value: number;
}

export interface Catalog {
    id: string;
    name: string;
    description?: string;
    validFor: TimePeriod;
    lastUpdate: string;
    lifecycleStatus: LifecycleStatus;
}

export interface Category {
    id: string;
    name: string;
    description?: string;
    parentId?: string;
    isRoot: boolean;
    catalogId?: string | null;
    validFor: TimePeriod;
    lastUpdate: string;
    lifecycleStatus: LifecycleStatus;
}

export interface ProductSpecCharacteristic {
    name: string;
    description?: string;
    valueType: 'string' | 'number' | 'boolean';
    configurable: boolean;
    validValues?: string[];
}

export interface ProductSpecification {
    id: string;
    name: string;
    description?: string;
    productNumber: string;
    isBundle: boolean;
    lifecycleStatus: LifecycleStatus;
    validFor: TimePeriod;
    lastUpdate: string;
    characteristics?: Record<string, ProductSpecCharacteristic>;
}

export interface Attachment {
    id: string;
    name: string;
    description?: string;
    url: string;
    type: string; // e.g. "Picture", "Document"
    mimeType?: string;
}

export interface PriceAlteration {
    name: string;
    type: 'discount' | 'fee';
}

export interface ProductOfferingPrice {
    id?: string;
    priceType: 'recurring' | 'one_time' | 'usage';
    price: Money;
    unitOfMeasure?: string;
    priceAlteration?: PriceAlteration;
}

export interface ProductOffering {
    id: string;
    name: string;
    description?: string;
    lifecycleStatus: LifecycleStatus;
    validFor: TimePeriod;
    lastUpdate: string;
    isBundle: boolean;
    isSellable: boolean;
    productSpecificationId?: string;
    productOfferingPrice?: ProductOfferingPrice[];
    categoryIds?: string[];
    attachments?: Attachment[];

    // Enriched data
    productSpecification?: ProductSpecification;
    categories?: Category[];
}

// API Payloads
export interface CreateCatalogPayload {
    name: string;
    description?: string;
    validFor?: TimePeriod;
}

export interface CreateCategoryPayload {
    name: string;
    description?: string;
    parentId?: string;
    isRoot: boolean;
    catalogId?: string;
    validFor?: TimePeriod;
}

export interface CreateProductSpecificationPayload {
    name: string;
    description?: string;
    productNumber: string;
    isBundle: boolean;
    lifecycleStatus: LifecycleStatus;
    validFor?: TimePeriod;
    characteristics?: Record<string, ProductSpecCharacteristic>;
}

export interface CreateProductOfferingPayload {
    name: string;
    description?: string;
    lifecycleStatus: LifecycleStatus;
    validFor?: TimePeriod;
    isBundle: boolean;
    isSellable: boolean;
    productSpecificationId: string;
    productOfferingPrice?: ProductOfferingPrice[];
    categoryIds: string[];
    attachments?: Attachment[];
}
