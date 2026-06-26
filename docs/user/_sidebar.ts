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
        text: 'Image Pulls Fail with "404 manifest unknown" Despite Correct Image Name',
        link: './troubleshooting/01-10-incorrect-credentials.md',
      },
    ],
  },
];
