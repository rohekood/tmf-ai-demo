// Party Types based on TMF632

export type PartyType = 'Individual' | 'Organization';

export type PartyStatus = 'initialized' | 'active' | 'inactive' | 'deleted';

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
}

export interface Characteristic {
    id: string;
    name: string;
    value: string;
    valueType?: string;
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
    contactMediums?: Omit<ContactMedium, 'id'>[];
    identifications?: Omit<Identification, 'id'>[];
    relatedParties?: Omit<RelatedParty, 'id'>[];
    characteristics?: Omit<Characteristic, 'id'>[];
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
}

export type CreatePartyPayload = CreateIndividualPayload | CreateOrganizationPayload;

export interface UpdatePartyPayload {
    id: string;
    '@type': PartyType;
    status?: PartyStatus;
    individual?: Partial<Omit<Individual, 'id' | '@type'>>;
    organization?: Partial<Omit<Organization, 'id' | '@type'>>;
}

export interface SearchPartyParams {
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
