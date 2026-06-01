import {existsSync, readFileSync} from 'node:fs';
import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const versionsFile = './api_versions.json';
const allVersions: string[] = existsSync(versionsFile)
  ? JSON.parse(readFileSync(versionsFile, 'utf8'))
  : [];
const hasVersions = allVersions.length > 0;
const top10 = allVersions.slice(0, 10);

const config: Config = {
  title: 'Ethereum Execution APIs',
  tagline: 'JSON-RPC API specification for Ethereum execution clients',
  favicon: 'img/favicon.ico',

  /*future: { v4: true },*/

  url: 'https://ethereum.github.io',
  baseUrl: '/execution-apis/',

  organizationName: 'ethereum',
  projectName: 'execution-apis',

  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: { defaultLocale: 'en', locales: ['en'] },

  plugins: [
   "docusaurus-plugin-copy-page-button",

    ['@open-rpc/docusaurus-plugin', {
      openRPCSpecPath: './openrpc.json',
      docOutputPath: './docs-api/api',
      indexSlug: '/'
    }],
    ['@docusaurus/plugin-content-docs', {
      id: 'api',
      path: 'docs-api',
      routeBasePath: '/',
      sidebarPath: './docs-api/sidebars.ts',
      ...(hasVersions && {
        versions: { current: { label: 'Next', path: 'next' } },
        onlyIncludeVersions: ['current', ...top10],
      }),
    }],
    ['@docusaurus/plugin-content-docs', {
      id: 'releases',
      path: 'docs-releases',
      routeBasePath: 'releases',
      sidebarPath: './docs-releases/sidebars.ts',
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
        { to: '/releases', label: 'Releases', position: 'left' },
        { type: 'docsVersionDropdown', docsPluginId: 'api', position: 'right' },
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
