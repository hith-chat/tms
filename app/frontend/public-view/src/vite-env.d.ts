// Vite environment types for TypeScript
// Add any VITE_* variables here to get proper typing for import.meta.env

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  // add more env vars as needed
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
