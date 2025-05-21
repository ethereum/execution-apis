import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */
const sidebars: SidebarsConfig = {
  referenceSidebar: [
    {
      type: "category",
      label: "Reference",
      link: { type: "generated-index" },
      items: [
        {
          type: "doc",
          id: "reference/intro",
          label: "Introduction"
        },
        {
          type: "doc",
          id: "reference/making-changes",
          label: "Contributors Guide"
        },
        {
          type: "category",
          label: "JSON-RPC API",
          items: [
            {
              type: "doc",
              id: "reference/json-rpc/introduction",
              label: "Introduction",
            },
            {
              type: "doc",
              id: "reference/json-rpc/eth-namespace",
              label: "ETH Namespace",
            },
            {
              type: "doc",
              id: "reference/json-rpc/debug-namespace",
              label: "Debug Namespace",
            },
            {
              type: "doc",
              id: "reference/json-rpc/engine-namespace",
              label: "Engine Namespace",
            }
          ]
        }
      ],
    },
  ],
};

export default sidebars;
