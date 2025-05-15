import { themes as prismThemes } from 'prism-react-renderer';
import type { Config } from '@docusaurus/types';
import type * as Preset from '@metamask/docusaurus-openrpc/dist/preset';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'Ethereum Execution APIs',
  tagline: 'A JSON-RPC specification for Ethereum execution clients',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://ethereum.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/execution-apis/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'ethereum', // Usually your GitHub org/user name.
  projectName: 'execution-apis', // Usually your repo name.
  deploymentBranch: 'gh-pages',

  onBrokenLinks: 'warn',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      '@metamask/docusaurus-openrpc/dist/preset',
      /** @type {import('@metamask/docusaurus-openrpc/dist/preset').Options} */
      {
        docs: {
          routeBasePath: '/',
          openrpc: {
            openrpcDocument: './refs-openrpc.json',
            path: 'reference',
            sidebarLabel: 'JSON-RPC',
          },
          sidebarPath: './sidebars.ts',
          path: './docs',
        },
        blog: false,
        theme: {
          customCss: './css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/docusaurus-social-card.jpg',
    navbar: {
      title: 'Execution APIs',
      // logo: {
      //   alt: 'My Site Logo',
      //   src: 'img/logo.svg',
      // },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'referenceSidebar',
          position: 'left',
          label: 'API Reference',
        },
        {
          href: 'https://github.com/ethereum/execution-apis',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/ethereum/execution-apis',
            },
          ],
        },
      ],
      copyright: `Copyright and related rights waived via CC0.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
