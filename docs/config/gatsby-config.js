const fs = require('fs');
const path = require('path');

// Function to get markdown files and create menu links
function getMenuLinksFromDocs() {
  const docsPath = path.join(__dirname, 'src/docs');
  const menuLinks = [];
  
  try {
    const files = fs.readdirSync(docsPath);
    files.forEach(file => {
      if (file.endsWith('.md') || file.endsWith('.mdx')) {
        const name = file.replace(/\.(md|mdx)$/, '').replace(/-/g, ' ');
        const link = `/${file.replace(/\.(md|mdx)$/, '')}`;
        menuLinks.push({
          name: name.charAt(0).toUpperCase() + name.slice(1),
          link: link
        });
      }
    });
  } catch (error) {
    console.warn('Could not read docs directory:', error.message);
  }
  
  return menuLinks;
}

module.exports = {
  pathPrefix: "/execution-apis",
  siteMetadata: {
    title: 'Ethereum JSON-RPC Specification',
    description: 'A specification of the standard interface for Ethereum clients.',
    siteUrl: process.env.GITHUB_REPOSITORY 
      ? `https://${process.env.GITHUB_REPOSITORY.split('/')[0]}.github.io/execution-apis`
      : 'https://ethereum.github.io/execution-apis', // fallback for local development
    logoUrl: 'https://raw.githubusercontent.com/open-rpc/design/master/icons/open-rpc-logo-noText/open-rpc-logo-noText%20(PNG)/256x256.png',
    primaryColor: '#3f51b5', //material-ui primary color
    secondaryColor: '#f50057', //material-ui secondary color
    author: '',
    menuLinks: [
      {
        name: 'API Documentation',
        link: '/api-documentation'
      },
      ...getMenuLinksFromDocs() // This will add all markdown files
    ],
    footerLinks: [
      {
        name: 'OpenRPC',
        link: 'https://open-rpc.org'
      }
    ]
  },
  plugins: [
   {
      resolve: 'gatsby-plugin-mdx',
      options: {
        extensions: ['.mdx', '.md'],
        gatsbyRemarkPlugins: [
          {
            resolve: 'gatsby-remark-autolink-headers',
            options: {
              icon: false,
            },
          },
        ],
      },
    },
    "gatsby-openrpc-theme",
    {
      resolve: 'gatsby-plugin-manifest',
      options: {
        name: 'pristine-site',
        short_name: 'pristine-site',
        start_url: '/execution-apis/',
        background_color: 'transparent',
        theme_color: '#3f51b5',
        display: 'minimal-ui',
        icon: 'src/images/gatsby-icon.png', // This path is relative to the root of the site.
      },
    },
    "gatsby-plugin-image",
    "gatsby-plugin-sharp",
    "gatsby-transformer-sharp",
    {
      resolve: "gatsby-source-filesystem",
      options: {
        name: "images",
        path: __dirname + '/src/images',
      },
    },
    {
      resolve: "gatsby-source-filesystem",
      options: {
        name: "docs",
        path: __dirname + '/src/docs',
      },
    },
  ],
}
