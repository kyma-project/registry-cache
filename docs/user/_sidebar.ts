// TODO: Add the docs/user/README.md entry to the overarching sidebar in the kyma repo.
export default [
  { text: 'Configure Registry Cache', link: './01-10-configure-registry-cache.md' },
  {
    text: 'Resources',
    link: './resources/README.md',
    collapsed: true,
    items: [
      { text: 'RegistryCache', link: './resources/RegistryCache.md' },
      { text: 'RegistryCacheConfig', link: './resources/RegistryCacheConfig.md' },
    ],
  },
  {
    text: 'Troubleshooting',
    link: './troubleshooting/README.md',
    collapsed: true,
    items: [
      {
        text: 'Registry Cache Does Not Cache Images from Private Registry',
        link: './troubleshooting/01-10-incorrect-credentials.md',
      },
    ],
  },
];
