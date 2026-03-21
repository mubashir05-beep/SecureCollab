#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use securecollab_client_core::commands;
use securecollab_client_core::crypto::{EncryptedPayload, IdentityKeyPair};

#[tauri::command]
fn generate_keys() -> Result<IdentityKeyPair, String> {
    commands::generate_keys()
}

#[tauri::command]
fn encrypt_message(
    our_private_b64: String,
    their_public_b64: String,
    plaintext: String,
) -> Result<EncryptedPayload, String> {
    commands::encrypt_message(our_private_b64, their_public_b64, plaintext)
}

#[tauri::command]
fn decrypt_message(
    our_private_b64: String,
    their_public_b64: String,
    ciphertext_b64: String,
    nonce_b64: String,
) -> Result<String, String> {
    commands::decrypt_message(our_private_b64, their_public_b64, ciphertext_b64, nonce_b64)
}

fn main() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            generate_keys,
            encrypt_message,
            decrypt_message,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
