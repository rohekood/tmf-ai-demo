// Customer Types based on TMF629

export type CustomerStatus = 'prospecting' | 'active' | 'suspended' | 'inactive' | 'deleted';

export interface CustomerAccount {
    id: string;
    name: string;
    accountStatus: 'active' | 'inactive' | 'suspended';
    accountType: string;
}

export interface CreditProfile {
    id: string;
    creditRiskScore: number;
    creditScore: number;
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

export interface Customer {
    id: string;
    name: string;
    status: CustomerStatus;
    partyId: string;
    partyName?: string;
    createdAt?: string;
    updatedAt?: string;
    accounts?: CustomerAccount[];
    creditProfile?: CreditProfile;
    contactMediums?: ContactMedium[];
    characteristics?: Characteristic[];
    taxExemptions?: TaxExemption[];
    privacyConsents?: PrivacyConsent[];
}

// API Payloads
export interface OnboardCustomerPayload {
    name: string;
    partyId: string;
    accounts?: Omit<CustomerAccount, 'id'>[];
    creditProfile?: Omit<CreditProfile, 'id'>;
    contactMediums?: Omit<ContactMedium, 'id'>[];
    characteristics?: Omit<Characteristic, 'id'>[];
    taxExemptions?: Omit<TaxExemption, 'id'>[];
    privacyConsents?: Omit<PrivacyConsent, 'id'>[];
}

export interface UpdateCustomerPayload {
    id: string;
    status?: CustomerStatus;
    name?: string;
    taxExemptions?: TaxExemption[];
    privacyConsents?: PrivacyConsent[];
}

export interface SearchCustomerParams {
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
