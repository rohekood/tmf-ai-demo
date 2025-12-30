// Customer Types based on TMF629

export type CustomerStatus = 'prospecting' | 'active' | 'suspended' | 'inactive' | 'deleted';

export interface CustomerAccount {
    id: string;
    name: string;
    accountStatus: 'active' | 'inactive' | 'suspended';
    accountType: string;
    billFormat?: string;
    billingCycle?: string;
}

export interface CreditProfile {
    id: string;
    creditRiskScore: number;
    creditScore: number;
    creditProfileDate?: string;
    validForStart?: string;
    validForEnd?: string;
}

export interface ContactMedium {
    id: string;
    mediumType: 'email' | 'phone' | 'postal';
    preferred: boolean;
    value?: string;
    street1?: string;
    street2?: string;
    city?: string;
    stateOrProvince?: string;
    postcode?: string;
    country?: string;
}

export interface Characteristic {
    id: string;
    name: string;
    value: string;
    valueType?: string;
}

export interface TaxExemption {
    id: string;
    certificateNumber: string;
    issuingJurisdiction: string;
    validForStart?: string;
    validForEnd?: string;
}

export interface PrivacyConsent {
    id: string;
    consentType: string;
    status: 'given' | 'revoked' | 'pending';
    validForStart?: string;
}

export interface RelatedParty {
    id: string;
    relatedPartyId: string;
    role: string;
    name: string;
    validForStart?: string;
    validForEnd?: string;
}

export interface PaymentMethod {
    id: string;
    type: string;
    token: string;
    details?: string;
    isDefault: boolean;
    validForStart?: string;
    validForEnd?: string;
}

export interface MarketSegment {
    id: string;
    name: string;
    category: string;
}

export interface CustomerInteraction {
    id: string;
    customerId: string;
    interactionDate: string;
    channel: string;
    type: string;
    description: string;
    agentId: string;
}

export interface AppliedBillingRate {
    id: string;
    productRef: string;
    rateType: string;
    value: number;
    validForStart?: string;
    validForEnd?: string;
}

export interface Customer {
    id: string;
    name: string;
    status: CustomerStatus;
    partyId: string;
    partyName?: string;
    partyType?: string;
    createdAt?: string;
    updatedAt?: string;
    accounts?: CustomerAccount[];
    creditProfiles?: CreditProfile[];
    contactMediums?: ContactMedium[];
    characteristics?: Characteristic[];
    taxExemptions?: TaxExemption[];
    privacyConsents?: PrivacyConsent[];
    relatedParties?: RelatedParty[];
    paymentMethods?: PaymentMethod[];
    marketSegments?: MarketSegment[];
    customerInteractions?: CustomerInteraction[];
    appliedBillingRates?: AppliedBillingRate[];
}

// API Payloads
export interface OnboardCustomerPayload {
    name: string;
    partyId: string;
    partyType?: string;
    accounts?: Omit<CustomerAccount, 'id'>[];
    creditProfiles?: Omit<CreditProfile, 'id'>[];
    contactMediums?: Omit<ContactMedium, 'id'>[];
    characteristics?: Omit<Characteristic, 'id'>[];
    taxExemptions?: Omit<TaxExemption, 'id'>[];
    privacyConsents?: Omit<PrivacyConsent, 'id'>[];
    relatedParties?: Omit<RelatedParty, 'id'>[];
    paymentMethods?: Omit<PaymentMethod, 'id'>[];
    marketSegments?: Omit<MarketSegment, 'id'>[];
    appliedBillingRates?: Omit<AppliedBillingRate, 'id'>[];
}

export interface UpdateCustomerPayload {
    id: string;
    status?: CustomerStatus;
    name?: string;
    taxExemptions?: TaxExemption[];
    privacyConsents?: PrivacyConsent[];
    accounts?: CustomerAccount[];
    creditProfiles?: CreditProfile[];
    partyId?: string;
    partyName?: string;
    partyType?: string;
    relatedParties?: RelatedParty[];
    paymentMethods?: PaymentMethod[];
    marketSegments?: MarketSegment[];
    appliedBillingRates?: AppliedBillingRate[];
}

export interface SearchCustomerParams {
    search?: string;
    name?: string;
    status?: string;
    partyId?: string;
}

// Helper functions
export function getStatusColor(status: CustomerStatus): string {
    switch (status) {
        case 'active':
            return 'green';
        case 'prospecting':
            return 'yellow';
        case 'suspended':
            return 'orange';
        case 'inactive':
        case 'deleted':
            return 'red';
        default:
            return 'gray';
    }
}
