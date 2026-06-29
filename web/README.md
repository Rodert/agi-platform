# AGI Platform Web

Frontend workspace for AGI Platform.

## Apps

- `web/user`: user-facing image generation console
- `web/admin`: admin operation console
- `web/shared`: shared API client and TypeScript types

## Development

Install dependencies from the repository root:

```bash
pnpm install
```

Run user app:

```bash
pnpm dev:user
```

Run admin app:

```bash
pnpm dev:admin
```

Build all apps:

```bash
pnpm build
```

## Backend URL

The frontend resolves the backend URL in this order:

- `VITE_API_BASE_URL`
- `http://127.0.0.1:8080` when the page is opened from `localhost`, `127.0.0.1`, or `::1`
- `https://api.newmovieai.com` for public deployed domains

For Cloudflare Pages, set:

```text
VITE_API_BASE_URL=https://api.newmovieai.com
```

## Default Admin

Seed SQL creates a local development admin:

```text
admin / admin123
```
