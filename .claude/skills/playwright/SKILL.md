---
name: playwright
description: Drive the TMF demo-ui in a real browser with Playwright MCP tools — navigate, log in via Auth0, click through flows, and verify UI behavior. Use when asked to manually test the UI, reproduce a UI bug, take screenshots, or confirm a frontend change works end-to-end.
---

# Playwright — driving the TMF demo-ui

Use the `mcp__playwright__browser_*` MCP tools to open and drive the demo-ui in a real browser.

## When to use

- Manually verifying a frontend change end-to-end (not just `yarn test`).
- Reproducing a reported UI bug before fixing it.
- Taking screenshots of a page or flow.
- Clicking through catalog / ordering / customer flows to confirm behavior.

## App URL

| How it's running | URL |
|---|---|
| `docker-compose up` (demo-ui container) | http://localhost |
| `yarn dev` from `services/demo-ui/ui` | http://localhost:5173 |

Auth0 browser login requires a **secure origin** — only `localhost`/HTTPS works. Do not use a LAN IP or `127.0.0.1`-vs-`localhost` mismatch, or the Log In button will be disabled.

## Login flow (Auth0 redirect)

The app gates everything behind Auth0. Credentials for the demo tenant:

- **Username:** `demo`
- **Password:** `demo1234$`

Steps:

1. `browser_navigate` to the app URL.
2. `browser_snapshot` to get the accessibility tree, then `browser_click` the **Log In** button.
3. This redirects to the Auth0 universal login page. `browser_snapshot` again, then fill the username/email and password fields (`browser_type` or `browser_fill_form`) and submit.
4. After redirect back, `browser_snapshot` to confirm you've landed on the dashboard.

If the Log In button shows "Auth Not Configured" or "Auth Requires HTTPS or localhost", the `AUTH0_*` env vars aren't set or you're on a non-secure origin — fix the environment rather than retrying.

## Key routes

Once logged in, navigate directly to these paths:

- `/parties`, `/customers`
- `/catalog/catalogs`, `/catalog/categories`, `/catalog/specifications`, `/catalog/offerings`
- `/order/qualify`, `/order/cart`

## Working tips

- Always `browser_snapshot` before interacting — it returns stable element refs to target with `browser_click` / `browser_type`. Prefer snapshots over screenshots for finding elements; use `browser_take_screenshot` only when the user wants a visual.
- Use `browser_wait_for` after navigation or actions that trigger async loads (most data comes over the BFF → RabbitMQ).
- Check `browser_console_messages` and `browser_network_requests` when a page misbehaves — failed BFF calls show up there.
- `browser_close` the browser when finished.
