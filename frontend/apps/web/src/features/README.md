# Feature modules

Each folder owns one user-facing capability. Keep feature-specific UI, data and helpers inside its folder.

- `auth/`: login UI and authentication entry points.
- `creation/`: image, video and product generation controls and their styles.
- `inspiration/`: inspiration feed, work cards, work detail and feed styles.
- `assets/`: asset management page and asset-specific styles.
- `community/`: community dialog and community-specific behavior.
- `notifications/`: notification list and drawer.

Shared application shell code stays in `components/AppLayout.tsx`. Cross-feature mock state currently stays in `store/AppStore.tsx` until the backend API is introduced. Feature `index.ts` files define the public imports used by routes and other features.

When adding a feature:

1. Create `features/<feature-name>/`.
2. Keep its components and local data in that folder.
3. Export a small component or hook for pages to consume.
4. Do not import another feature's internal files. Use its public component or shared store contract.
