/**
 * Centralized localStorage key for the active cart ID.
 * Import this constant in every component that reads or writes the cart ID
 * to guarantee all components use the same key.
 */
export const CART_ID_KEY = 'cartId';

/**
 * Session-storage key holding the address an anonymous user entered before being
 * sent to login when they tried to add an offering to the cart. After login we
 * restore it and re-run qualification so the user can resume with customer
 * pricing.
 */
export const QUALIFY_RESUME_KEY = 'qualifyResumeAddress';
