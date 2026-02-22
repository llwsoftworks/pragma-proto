/**
 * AES-256-GCM encryption for login payloads.
 *
 * The Go API expects: {"encrypted": "<base64(nonce || ciphertext || tag)>"}
 * The shared key is stored in LOGIN_ENCRYPTION_KEY env var (base64-encoded, 32 bytes).
 *
 * This module is server-only (src/lib/server/) — it is NEVER shipped to the browser.
 */

import { env } from '$env/dynamic/private';
import { createCipheriv, randomBytes } from 'node:crypto';

const ALGORITHM = 'aes-256-gcm';
const NONCE_LENGTH = 12; // 96-bit nonce recommended for AES-GCM

function getKey(): Buffer {
	const keyB64 = env.LOGIN_ENCRYPTION_KEY;
	if (!keyB64) {
		throw new Error('LOGIN_ENCRYPTION_KEY environment variable is not set');
	}
	const key = Buffer.from(keyB64, 'base64');
	if (key.length !== 32) {
		throw new Error('LOGIN_ENCRYPTION_KEY must be exactly 32 bytes (AES-256)');
	}
	return key;
}

/**
 * Encrypt a plaintext string with AES-256-GCM.
 * Returns base64(nonce || ciphertext || authTag).
 */
export function encryptLoginPayload(plaintext: string): string {
	const key = getKey();
	const nonce = randomBytes(NONCE_LENGTH);
	const cipher = createCipheriv(ALGORITHM, key, nonce);

	const encrypted = Buffer.concat([cipher.update(plaintext, 'utf8'), cipher.final()]);
	const authTag = cipher.getAuthTag(); // 16 bytes

	// nonce (12) + ciphertext (variable) + authTag (16)
	const combined = Buffer.concat([nonce, encrypted, authTag]);
	return combined.toString('base64');
}
