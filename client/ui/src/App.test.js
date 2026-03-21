import { render, screen } from "@testing-library/svelte";
import App from "./App.svelte";

describe("App", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("renders landing page when unauthenticated", () => {
    render(App);
    expect(screen.getByText("SecureCollab")).toBeTruthy();
    expect(screen.getByText("Get Started")).toBeTruthy();
  });

  it("shows zero-knowledge tagline", () => {
    render(App);
    expect(screen.getByText(/zero-knowledge/i)).toBeTruthy();
  });
});
