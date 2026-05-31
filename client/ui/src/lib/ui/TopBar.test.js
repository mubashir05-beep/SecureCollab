import { render, screen } from "@testing-library/svelte";
import TopBar from "./TopBar.svelte";

describe("TopBar", () => {
  it("renders live connection status and member count", () => {
    render(TopBar, {
      props: {
        channelName: "general",
        channelTopic: "Project updates",
        memberCount: 12,
        connectionState: "connected",
        liveActivity: "Live now",
      },
    });

    expect(screen.getByText("Live")).toBeTruthy();
    expect(screen.getByText("general")).toBeTruthy();
    expect(screen.getByText("12")).toBeTruthy();
    expect(screen.getByText("Live now")).toBeTruthy();
  });
});