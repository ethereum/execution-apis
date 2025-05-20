import fs from 'fs';

/**
 * Generates a namespaced sidebar configuration based on the OpenRPC document
 */
async function generateNamespacedSidebars() {
  console.log('Generating namespaced sidebars...');

  // Read the OpenRPC document
  const openrpcData = JSON.parse(fs.readFileSync('refs-openrpc.json', 'utf8'));
  
  // Read the method namespaces mapping
  const methodNamespaces = JSON.parse(fs.readFileSync('method-namespaces.json', 'utf8'));

  // Group methods by namespace
  const namespaces = {};
  
  openrpcData.methods.forEach(method => {
    const methodName = method.name;
    const namespace = methodNamespaces[methodName] || extractNamespace(methodName);
    if (!namespaces[namespace]) {
      namespaces[namespace] = [];
    }
    namespaces[namespace].push(methodName);
  });

  // Make sure directory exists
  if (!fs.existsSync('docs/reference/json-rpc')) {
    fs.mkdirSync('docs/reference/json-rpc', { recursive: true });
  }

  // Add custom JSON-RPC reference sidebar to docs
  fs.writeFileSync('docs/reference/json-rpc/eth-namespace.md', generateSidebarPage('ETH Namespace', namespaces['eth'] || []));
  fs.writeFileSync('docs/reference/json-rpc/debug-namespace.md', generateSidebarPage('Debug Namespace', namespaces['debug'] || []));
  fs.writeFileSync('docs/reference/json-rpc/engine-namespace.md', generateSidebarPage('Engine Namespace', namespaces['engine'] || []));
  
  // Create a sidebar introduction page
  fs.writeFileSync('docs/reference/json-rpc/introduction.md', generateIntroductionPage(Object.keys(namespaces)));
  
  console.log('Namespaced sidebar files generated successfully!');
}

/**
 * Extracts namespace from method name (e.g., eth_blockNumber -> eth)
 */
function extractNamespace(methodName) {
  if (!methodName) return 'other';
  
  const parts = methodName.split('_');
  if (parts.length > 1) {
    return parts[0];
  }
  
  return 'other';
}

/**
 * Generates a markdown page for a specific namespace
 */
function generateSidebarPage(title, methods) {
  const sortedMethods = [...methods].sort();
  
  return `---
sidebar_position: 1
---

# ${title}

This page lists all JSON-RPC methods in the ${title.toLowerCase()}.

${sortedMethods.map(method => `- [${method}](/reference/${method})`).join('\n')}
`;
}

/**
 * Generates an introduction page for JSON-RPC API
 */
function generateIntroductionPage(namespaces) {
  return `---
sidebar_position: 1
---

# JSON-RPC API Introduction

The Ethereum JSON-RPC API is organized into several namespaces, each containing related methods:

${namespaces.map(namespace => `- [${namespace.charAt(0).toUpperCase() + namespace.slice(1)} Namespace](/reference/json-rpc/${namespace}-namespace)`).join('\n')}

Each namespace contains methods with specific functionality related to different aspects of Ethereum node operation.
`;
}

generateNamespacedSidebars().catch(console.error); 