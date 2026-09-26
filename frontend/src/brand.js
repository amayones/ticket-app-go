/* global __APP_NAME__, __APP_LOGO__ */
// Branding aplikasi: nama + logo. Satu sumber kebenaran ada di .env root
// (APP_NAME, APP_LOGO); nilainya disuntikkan saat build oleh vite.config.js.
//
// Ganti nama/logo aplikasi = ubah .env, lalu `task build-frontend`.
// Logo default memakai favicon yang ada di frontend/public/favicon.svg.
export const APP_NAME = __APP_NAME__

export const APP_LOGO = __APP_LOGO__
