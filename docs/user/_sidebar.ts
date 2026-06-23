const items = [
  {
    name: 'Registry Cache',
    items: [
      { name: 'Overview', url: '/docs/user/README.md' },
      { name: 'Configure Registry Cache', url: '/docs/user/01-10-configure-registry-cache.md' },
      {
        name: 'Resources',
        items: [
          { name: 'RegistryCache', url: '/docs/user/resources/RegistryCache.md' },
          { name: 'RegistryCacheConfig', url: '/docs/user/resources/RegistryCacheConfig.md' },
        ],
      },
      {
        name: 'Troubleshooting',
        items: [
          { name: 'Image Pulls Fail with 404 manifest unknown', url: '/docs/user/troubleshooting/01-10-incorrect-credentials.md' },
        ],
      },
    ],
  },
];

export default items;
