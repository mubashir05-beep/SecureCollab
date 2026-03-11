pub mod crypto {
    use base64::Engine;
    use rand_core::OsRng;
    use sha2::{Digest, Sha256};
    use x25519_dalek::{PublicKey, StaticSecret};

    #[derive(Debug, Clone, PartialEq, Eq)]
    pub struct IdentityKeyPair {
        pub private_key_b64: String,
        pub public_key_b64: String,
        pub key_id: String,
    }

    pub fn generate_identity_keypair() -> IdentityKeyPair {
        let private = StaticSecret::random_from_rng(OsRng);
        let public = PublicKey::from(&private);

        let private_key_b64 = base64::engine::general_purpose::STANDARD.encode(private.to_bytes());
        let public_key_b64 = base64::engine::general_purpose::STANDARD.encode(public.to_bytes());

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
}

#[cfg(test)]
mod tests {
    use super::crypto::generate_identity_keypair;

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
}
