// TODO: Add the docs/user/README.md entry to the overarching sidebar in the kyma repo.
export default [
  { text: 'Configuring Registry Cache', link: './01-10-configure-registry-cache.md' },
  { text: 'Providing Credentials for a Private Upstream Registry', link: './02-10-provide-credentials.md' },
  { text: 'Validation of Registry Cache Configuration', link: './02-20-validation.md' },
  { text: 'Managing Registry Cache Configuration', link: './02-30-manage-registry-cache.md' },
  { text: 'Rotating Credentials', link: './03-10-rotate-credentials.md' },
  {
    text: 'Resources',
    link: './resources/README.md',
    collapsed: true,
    items: [
      { text: 'RegistryCache Custom Resource', link: './resources/registry_cache_cr.md' },
      { text: 'RegistryCacheConfig Custom Resource', link: './resources/registry_cache_config_cr.md' },
    ],
  },
  {
    text: 'Troubleshooting',
    link: './troubleshooting/README.md',
    collapsed: true,
    items: [
      {
        text: 'Registry Cache Does Not Cache Images from a Private Registry',
        link: './troubleshooting/01-10-incorrect-credentials.md',
      },
    ],
  },
];
