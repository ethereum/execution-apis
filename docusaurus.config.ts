import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Ethereum Execution APIs',
  tagline: 'JSON-RPC API specification for Ethereum execution clients',
  favicon: 'img/favicon.ico',

  /*future: { v4: true },*/

  /* TODO replace url: 'https://ethereum.github.io', */
  url: 'https://open-rpc.github.io',
  baseUrl: '/execution-apis/',

  organizationName: 'ethereum',
  projectName: 'execution-apis',

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: { defaultLocale: 'en', locales: ['en'] },

  plugins: [
    ['@open-rpc/docusaurus-plugin', {
      openRPCSpecPath: './execution-apis.json',
      docOutputPath: './docs-api/api',
      indexSlug: '/'
    }],
    ['@docusaurus/plugin-content-docs', {
      id: 'api',
      path: 'docs-api',
      routeBasePath: '/',
      sidebarPath: './docs-api/sidebars.ts',
    }],
    ['@docusaurus/plugin-client-redirects', {
      redirects: [{ from: '/api', to: '/' }],
    }],
  ],

  presets: [
    ['classic', {
      docs: false,
      pages: {
        path: 'docs-api/src/pages',
      },
      blog: false,
      theme: {
        customCss: './docs-api/src/css/custom.css',
      },
    } satisfies Preset.Options],
  ],

  themes: [
    ['@easyops-cn/docusaurus-search-local', {
      hashed: true,
      docsPluginIdForPreferredVersion: 'api',
      docsRouteBasePath: '/',
      indexBlog: false,
    }],
  ],

  staticDirectories: ['docs-api/static'],

  themeConfig: {
    colorMode: { respectPrefersColorScheme: true },
    navbar: {
      title: 'Execution APIs',
      items: [
        { type: 'docSidebar', sidebarId: 'docsSidebar', docsPluginId: 'api', label: 'Docs', position: 'left' },
        { href: 'https://github.com/ethereum/execution-apis', label: 'GitHub', position: 'right' },
      ],
    },
    footer: {
      style: 'dark',
      copyright: `Copyright © ${new Date().getFullYear()} Ethereum Foundation. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
