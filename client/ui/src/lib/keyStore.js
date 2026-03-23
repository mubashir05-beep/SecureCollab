import { writable, derived } from "svelte/store";
import { generateKeys } from "./crypto.js";
import * as api from "./api.js";

const STORAGE_KEY = "sc_keys";

function createKeyStore() {
  const stored = typeof localStorage !== "undefined" ? localStorage.getItem(STORAGE_KEY) : null;
  const initial = stored ? JSON.parse(stored) : { privateKey: null, publicKey: null, keyId: null };

  const { subscribe, set } = writable(initial);

  function save(state) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    set(state);
  }

  return {
    subscribe,

    /** Bootstrap keys: generate if needed, upload to keydist */
    async bootstrap(token) {
      // Check if we already have keys locally
      const existing = JSON.parse(localStorage.getItem(STORAGE_KEY) || "null");
      if (existing?.privateKey && existing?.publicKey) {
        set(existing);
        return existing;
      }

      // Generate new keypair
      const keys = await generateKeys();
      const state = {
        privateKey: keys.private_key_b64,
        publicKey: keys.public_key_b64,
        keyId: keys.key_id,
      };

      // Upload public key to keydist
      try {
        await api.uploadIdentityKey(token, state.publicKey);
      } catch (err) {
        console.warn("Key upload failed (may already exist):", err.message);
      }

      save(state);
      return state;
    },

    /** Fetch a peer's public key from keydist */
    async getPeerKey(token, userId) {
      const cacheKey = `sc_peer_${userId}`;
      const cached = localStorage.getItem(cacheKey);
      if (cached) return cached;

      try {
        const res = await api.getIdentityKey(token, userId);
        const pubKey = res.public_key_b64;
        localStorage.setItem(cacheKey, pubKey);
        return pubKey;
      } catch {
        return null;
      }
    },

    clear() {
      localStorage.removeItem(STORAGE_KEY);
      set({ privateKey: null, publicKey: null, keyId: null });
    },
  };
}

export const keyStore = createKeyStore();
export const hasKeys = derived(keyStore, ($ks) => !!$ks.privateKey);
