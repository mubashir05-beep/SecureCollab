import { writable, derived } from "svelte/store";

function createAuthStore() {
  const stored = typeof localStorage !== "undefined" ? localStorage.getItem("sc_auth") : null;
  const initial = stored ? JSON.parse(stored) : { token: null, username: null, userId: null };

  const { subscribe, set, update } = writable(initial);

  return {
    subscribe,
    login(token, username, userId) {
      const state = { token, username, userId };
      localStorage.setItem("sc_auth", JSON.stringify(state));
      set(state);
    },
    logout() {
      localStorage.removeItem("sc_auth");
      set({ token: null, username: null, userId: null });
    },
    updateToken(token) {
      update((s) => {
        const state = { ...s, token };
        localStorage.setItem("sc_auth", JSON.stringify(state));
        return state;
      });
    },
  };
}

export const auth = createAuthStore();
export const isAuthenticated = derived(auth, ($auth) => !!$auth.token);
