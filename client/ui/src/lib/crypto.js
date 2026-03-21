/**
 * Crypto abstraction — uses Tauri invoke() when available,
 * falls back to Web Crypto placeholder for browser-only dev.
 */

const isTauri = typeof window !== "undefined" && window.__TAURI__;

async function tauriInvoke(cmd, args) {
  // Dynamic import with variable to prevent Vite from resolving at build time
  const tauriPath = "@tauri-apps/api/tauri";
  const { invoke } = await import(/* @vite-ignore */ tauriPath);
  return invoke(cmd, args);
}

export async function generateKeys() {
  if (isTauri) {
    return tauriInvoke("generate_keys");
  }
  // Browser fallback: generate placeholder keys for dev/testing
  const keyData = crypto.getRandomValues(new Uint8Array(32));
  const b64 = btoa(String.fromCharCode(...keyData));
  const pubData = crypto.getRandomValues(new Uint8Array(32));
  const pubB64 = btoa(String.fromCharCode(...pubData));
  const idBytes = crypto.getRandomValues(new Uint8Array(8));
  const keyId = Array.from(idBytes).map(b => b.toString(16).padStart(2, "0")).join("");
  return { private_key_b64: b64, public_key_b64: pubB64, key_id: keyId };
}

export async function encryptMessage(ourPrivateB64, theirPublicB64, plaintext) {
  if (isTauri) {
    return tauriInvoke("encrypt_message", {
      ourPrivateB64: ourPrivateB64,
      theirPublicB64: theirPublicB64,
      plaintext,
    });
  }
  // Browser fallback: XOR placeholder encryption
  const encoder = new TextEncoder();
  const data = encoder.encode(plaintext);
  const nonce = crypto.getRandomValues(new Uint8Array(24));
  const ciphertext = new Uint8Array(data.length);
  for (let i = 0; i < data.length; i++) {
    ciphertext[i] = data[i] ^ nonce[i % nonce.length];
  }
  return {
    ciphertext_b64: btoa(String.fromCharCode(...ciphertext)),
    nonce_b64: btoa(String.fromCharCode(...nonce)),
  };
}

export async function decryptMessage(ourPrivateB64, theirPublicB64, ciphertextB64, nonceB64) {
  if (isTauri) {
    return tauriInvoke("decrypt_message", {
      ourPrivateB64: ourPrivateB64,
      theirPublicB64: theirPublicB64,
      ciphertextB64,
      nonceB64,
    });
  }
  // Browser fallback: XOR placeholder decryption
  try {
    const ciphertext = Uint8Array.from(atob(ciphertextB64), c => c.charCodeAt(0));
    const nonce = Uint8Array.from(atob(nonceB64), c => c.charCodeAt(0));
    const plain = new Uint8Array(ciphertext.length);
    for (let i = 0; i < ciphertext.length; i++) {
      plain[i] = ciphertext[i] ^ nonce[i % nonce.length];
    }
    return new TextDecoder().decode(plain);
  } catch {
    return "[decryption failed]";
  }
}
