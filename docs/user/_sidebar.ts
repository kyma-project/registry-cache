export default [
  { text: 'Configure Registry Cache', link: './01-10-configure-registry-cache' },
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'RegistryCache', link: './resources/RegistryCache' },
    { text: 'RegistryCacheConfig', link: './resources/RegistryCacheConfig' },
    ] },
  { text: 'Troubleshooting', link: './troubleshooting/README', collapsed: true, items: [
    { text: 'Registry Cache Does Not Cache Images from Private Registry', link: './troubleshooting/01-10-incorrect-credentials' },
    ] }
];