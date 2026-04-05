# SS Mobile (React Native)

This is the first migration step from the existing web frontend (`client`) to React Native.

## What's included

- Expo + TypeScript mobile app
- Native stack navigation
- Auth context using AsyncStorage
- Initial screens mapped to existing web routes:
  - Home
  - Login
  - Register
  - Photos
  - Devices
  - Statistics

## Setup

```bash
cd client-native
npm install
cp .env.example .env
npm run start
```

## API base URL

Use `EXPO_PUBLIC_API_BASE_URL` in `.env`.

Example:

```env
EXPO_PUBLIC_API_BASE_URL=http://127.0.0.1:8080
```

When running on a real device, replace `127.0.0.1` with your machine LAN IP.

## Migration plan

1. Keep `client` untouched for web delivery.
2. Move reusable business logic (API calls, auth/session handling, models) into shared modules.
3. Replace placeholder screens with feature-complete React Native flows.
4. Add platform-specific adaptations only where needed.
