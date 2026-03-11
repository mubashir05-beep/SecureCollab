import { render, screen } from "@testing-library/svelte";
import App from "./App.svelte";

describe("App", () => {
  it("renders workspace shell", () => {
    render(App);
    expect(screen.getByText("SecureCollab")).toBeTruthy();
    expect(screen.getByText("Workspace Console")).toBeTruthy();
    expect(screen.getByText("Sign In")).toBeTruthy();
  });
});
