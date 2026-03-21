pub mod crypto {
    use base64::Engine;
    use chacha20poly1305::{
        aead::{Aead, KeyInit, OsRng},
        ChaCha20Poly1305, Nonce,
    };
    use rand_core::RngCore;
    use sha2::{Digest, Sha256};
    use x25519_dalek::{PublicKey, StaticSecret};

    #[derive(Debug, Clone, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
    pub struct IdentityKeyPair {
        pub private_key_b64: String,
        pub public_key_b64: String,
        pub key_id: String,
    }

    #[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
    pub struct EncryptedPayload {
        pub ciphertext_b64: String,
        pub nonce_b64: String,
    }

    pub fn generate_identity_keypair() -> IdentityKeyPair {
        let private = StaticSecret::random_from_rng(OsRng);
        let public = PublicKey::from(&private);

        let engine = base64::engine::general_purpose::STANDARD;
        let private_key_b64 = engine.encode(private.to_bytes());
        let public_key_b64 = engine.encode(public.to_bytes());

        let mut hasher = Sha256::new();
        hasher.update(public.to_bytes());
        let digest = hasher.finalize();
        let key_id = hex::encode(&digest[..8]);

        IdentityKeyPair {
            private_key_b64,
            public_key_b64,
            key_id,
        }
    }

    /// Derive a shared secret from our private key and their public key,
    /// then encrypt plaintext with ChaCha20-Poly1305.
    pub fn encrypt(
        our_private_b64: &str,
        their_public_b64: &str,
        plaintext: &[u8],
    ) -> Result<EncryptedPayload, String> {
        let shared = derive_shared_key(our_private_b64, their_public_b64)?;
        let cipher = ChaCha20Poly1305::new_from_slice(&shared)
            .map_err(|e| format!("cipher init: {e}"))?;

        let mut nonce_bytes = [0u8; 12];
        OsRng.fill_bytes(&mut nonce_bytes);
        let nonce = Nonce::from_slice(&nonce_bytes);

        let ciphertext = cipher
            .encrypt(nonce, plaintext)
            .map_err(|e| format!("encrypt: {e}"))?;

        let engine = base64::engine::general_purpose::STANDARD;
        Ok(EncryptedPayload {
            ciphertext_b64: engine.encode(&ciphertext),
            nonce_b64: engine.encode(nonce_bytes),
        })
    }

    /// Derive a shared secret and decrypt ciphertext.
    pub fn decrypt(
        our_private_b64: &str,
        their_public_b64: &str,
        ciphertext_b64: &str,
        nonce_b64: &str,
    ) -> Result<Vec<u8>, String> {
        let engine = base64::engine::general_purpose::STANDARD;
        let ciphertext = engine
            .decode(ciphertext_b64)
            .map_err(|e| format!("ciphertext decode: {e}"))?;
        let nonce_bytes: [u8; 12] = engine
            .decode(nonce_b64)
            .map_err(|e| format!("nonce decode: {e}"))?
            .try_into()
            .map_err(|_| "nonce must be 12 bytes".to_string())?;

        let shared = derive_shared_key(our_private_b64, their_public_b64)?;
        let cipher = ChaCha20Poly1305::new_from_slice(&shared)
            .map_err(|e| format!("cipher init: {e}"))?;
        let nonce = Nonce::from_slice(&nonce_bytes);

        cipher
            .decrypt(nonce, ciphertext.as_ref())
            .map_err(|e| format!("decrypt: {e}"))
    }

    fn derive_shared_key(our_private_b64: &str, their_public_b64: &str) -> Result<[u8; 32], String> {
        let engine = base64::engine::general_purpose::STANDARD;

        let private_bytes: [u8; 32] = engine
            .decode(our_private_b64)
            .map_err(|e| format!("private key decode: {e}"))?
            .try_into()
            .map_err(|_| "private key must be 32 bytes".to_string())?;

        let public_bytes: [u8; 32] = engine
            .decode(their_public_b64)
            .map_err(|e| format!("public key decode: {e}"))?
            .try_into()
            .map_err(|_| "public key must be 32 bytes".to_string())?;

        let private = StaticSecret::from(private_bytes);
        let public = PublicKey::from(public_bytes);
        let shared = private.diffie_hellman(&public);

        // Hash the raw shared secret for use as a symmetric key
        let mut hasher = Sha256::new();
        hasher.update(shared.as_bytes());
        Ok(hasher.finalize().into())
    }
}

/// Tauri command wrappers — these are invocable from the Svelte frontend
/// via `invoke("generate_keys")`, `invoke("encrypt_message", { ... })`, etc.
pub mod commands {
    use super::crypto;

    pub fn generate_keys() -> Result<crypto::IdentityKeyPair, String> {
        Ok(crypto::generate_identity_keypair())
    }

    pub fn encrypt_message(
        our_private_b64: String,
        their_public_b64: String,
        plaintext: String,
    ) -> Result<crypto::EncryptedPayload, String> {
        crypto::encrypt(&our_private_b64, &their_public_b64, plaintext.as_bytes())
    }

    pub fn decrypt_message(
        our_private_b64: String,
        their_public_b64: String,
        ciphertext_b64: String,
        nonce_b64: String,
    ) -> Result<String, String> {
        let bytes = crypto::decrypt(&our_private_b64, &their_public_b64, &ciphertext_b64, &nonce_b64)?;
        String::from_utf8(bytes).map_err(|e| format!("utf8: {e}"))
    }
}

#[cfg(test)]
mod tests {
    use super::crypto::{decrypt, encrypt, generate_identity_keypair};

    #[test]
    fn generates_non_empty_identity_keys() {
        let keys = generate_identity_keypair();
        assert!(!keys.private_key_b64.is_empty());
        assert!(!keys.public_key_b64.is_empty());
        assert!(!keys.key_id.is_empty());
    }

    #[test]
    fn generates_distinct_keypairs() {
        let a = generate_identity_keypair();
        let b = generate_identity_keypair();
        assert_ne!(a.private_key_b64, b.private_key_b64);
        assert_ne!(a.public_key_b64, b.public_key_b64);
        assert_ne!(a.key_id, b.key_id);
    }

    #[test]
    fn encrypt_decrypt_roundtrip() {
        let alice = generate_identity_keypair();
        let bob = generate_identity_keypair();

        let plaintext = b"Hello from Alice to Bob!";
        let payload = encrypt(&alice.private_key_b64, &bob.public_key_b64, plaintext).unwrap();

        // Bob decrypts using his private key + Alice's public key
        let decrypted = decrypt(
            &bob.private_key_b64,
            &alice.public_key_b64,
            &payload.ciphertext_b64,
            &payload.nonce_b64,
        )
        .unwrap();

        assert_eq!(decrypted, plaintext);
    }

    #[test]
    fn wrong_key_fails_decrypt() {
        let alice = generate_identity_keypair();
        let bob = generate_identity_keypair();
        let eve = generate_identity_keypair();

        let payload = encrypt(&alice.private_key_b64, &bob.public_key_b64, b"secret").unwrap();

        // Eve cannot decrypt
        let result = decrypt(
            &eve.private_key_b64,
            &alice.public_key_b64,
            &payload.ciphertext_b64,
            &payload.nonce_b64,
        );
        assert!(result.is_err());
    }

    #[test]
    fn command_wrappers_work() {
        use super::commands::*;

        let keys = generate_keys().unwrap();
        let bob = generate_keys().unwrap();

        let enc = encrypt_message(
            keys.private_key_b64.clone(),
            bob.public_key_b64.clone(),
            "test message".to_string(),
        )
        .unwrap();

        let dec = decrypt_message(
            bob.private_key_b64,
            keys.public_key_b64,
            enc.ciphertext_b64,
            enc.nonce_b64,
        )
        .unwrap();

        assert_eq!(dec, "test message");
    }
}
