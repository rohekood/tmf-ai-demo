// Party Types based on TMF632

export type PartyType = 'Individual' | 'Organization';

export type PartyStatus = 'Initialized' | 'Active' | 'Inactive' | 'Deleted' | 'DeletionPending';

export interface ContactMedium {
    id: string;
    mediumType: 'email' | 'phone' | 'postal';
    preferred: boolean;
    value?: string;
    // Postal address fields
    street1?: string;
    street2?: string;
    city?: string;
    stateOrProvince?: string;
    postcode?: string;
    country?: string;
}

export interface Identification {
    id: string;
    identificationType: string;
    identificationId: string;
    issuingAuthority?: string;
    issuingDate?: string;
}

export interface RelatedParty {
    id: string;
    relatedPartyId: string;
    relatedPartyName?: string;
    role: string;
    permissions?: string[];
}

export interface Characteristic {
    id: string;
    name: string;
    value: string;
    valueType?: string;
}

export interface ExternalReference {
    id: string;
    partyId: string;
    externalSystemId: string;
    externalReference: string;
}

export interface TaxExemption {
    id: string;
    partyId: string;
    certificateNumber: string;
    issuingJurisdiction: string;
    validFor: {
        startDateTime: string;
        endDateTime?: string;
    };
}

export interface Attachment {
    id: string;
    ownerId: string;
    mimeType: string;
    name: string;
    url: string;
    type: string;
}

// Base Party interface
export interface Party {
    id: string;
    '@type': PartyType;
    status: PartyStatus;
    createdAt?: string;
    updatedAt?: string;
    contactMediums?: ContactMedium[];
    identifications?: Identification[];
    relatedParties?: RelatedParty[];
    characteristics?: Characteristic[];
    externalReferences?: ExternalReference[];
    taxExemptions?: TaxExemption[];
    attachments?: Attachment[];
}

// Individual extends Party
export interface Individual extends Party {
    '@type': 'Individual';
    givenName: string;
    familyName: string;
    middleName?: string;
    birthDate?: string;
    gender?: string;
}

// Organization extends Party
export interface Organization extends Party {
    '@type': 'Organization';
    tradingName: string;
    isLegalEntity: boolean;
    organizationType?: string;
}

// Union type for polymorphic handling
export type PartyUnion = Individual | Organization;

// API Payloads
export interface CreateIndividualPayload {
    '@type': 'Individual';
    givenName: string;
    familyName: string;
    middleName?: string;
    birthDate?: string;
    gender?: string;
    contactMediums?: Omit<ContactMedium, 'id'>[];
    identifications?: Omit<Identification, 'id'>[];
    relatedParties?: Omit<RelatedParty, 'id'>[];
    characteristics?: Omit<Characteristic, 'id'>[];
    externalReferences?: Omit<ExternalReference, 'id' | 'partyId'>[];
    taxExemptions?: Omit<TaxExemption, 'id' | 'partyId'>[];
    attachments?: Omit<Attachment, 'id' | 'ownerId'>[];
}

export interface CreateOrganizationPayload {
    '@type': 'Organization';
    tradingName: string;
    isLegalEntity: boolean;
    organizationType?: string;
    contactMediums?: Omit<ContactMedium, 'id'>[];
    identifications?: Omit<Identification, 'id'>[];
    relatedParties?: Omit<RelatedParty, 'id'>[];
    characteristics?: Omit<Characteristic, 'id'>[];
    externalReferences?: Omit<ExternalReference, 'id' | 'partyId'>[];
    taxExemptions?: Omit<TaxExemption, 'id' | 'partyId'>[];
    attachments?: Omit<Attachment, 'id' | 'ownerId'>[];
}

export type CreatePartyPayload = CreateIndividualPayload | CreateOrganizationPayload;

// Update payload is similar to Create but with ID and flat structure
export type UpdatePartyPayload = {
    id: string;
    '@type': 'Individual';
    status?: PartyStatus;
    givenName?: string;
    familyName?: string;
    middleName?: string;
    birthDate?: string;
    gender?: string;
    contactMediums?: (Omit<ContactMedium, 'id'> | ContactMedium)[];
    identifications?: (Omit<Identification, 'id'> | Identification)[];
    relatedParties?: (Omit<RelatedParty, 'id'> | RelatedParty)[];
    externalReferences?: (Omit<ExternalReference, 'id' | 'partyId'> | ExternalReference)[];
    taxExemptions?: (Omit<TaxExemption, 'id' | 'partyId'> | TaxExemption)[];
    attachments?: (Omit<Attachment, 'id' | 'ownerId'> | Attachment)[];
} | {
    id: string;
    '@type': 'Organization';
    status?: PartyStatus;
    tradingName?: string;
    isLegalEntity?: boolean;
    organizationType?: string;
    contactMediums?: (Omit<ContactMedium, 'id'> | ContactMedium)[];
    identifications?: (Omit<Identification, 'id'> | Identification)[];
    relatedParties?: (Omit<RelatedParty, 'id'> | RelatedParty)[];
    externalReferences?: (Omit<ExternalReference, 'id' | 'partyId'> | ExternalReference)[];
    taxExemptions?: (Omit<TaxExemption, 'id' | 'partyId'> | TaxExemption)[];
    attachments?: (Omit<Attachment, 'id' | 'ownerId'> | Attachment)[];
};

export interface SearchPartyParams {
    search?: string;
    name?: string;
    givenName?: string;
    familyName?: string;
    tradingName?: string;
    type?: PartyType;
}

// Helper to get display name
export function getPartyDisplayName(party: PartyUnion): string {
    if (party['@type'] === 'Individual') {
        return `${party.givenName} ${party.familyName}`;
    }
    return party.tradingName;
}

// Type guard
export function isIndividual(party: PartyUnion): party is Individual {
    return party['@type'] === 'Individual';
}

export function isOrganization(party: PartyUnion): party is Organization {
    return party['@type'] === 'Organization';
}
