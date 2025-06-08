import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

// Get __dirname equivalent in ESM
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Function to ensure directory exists
function ensureDirectoryExists(dirPath) {
  if (!fs.existsSync(dirPath)) {
    fs.mkdirSync(dirPath, { recursive: true });
  }
}

// Function to copy markdown files
function copyMarkdownFiles() {
  const sourceDir = path.join(__dirname, '..', 'docs');
  const targetDir = path.join(__dirname, '..', 'build', 'docs', 'gatsby', 'src', 'docs');

  // Ensure target directory exists
  ensureDirectoryExists(targetDir);

  // Read all files from source directory
  const files = fs.readdirSync(sourceDir);

  // Copy each markdown file
  files.forEach(file => {
    if (file.endsWith('.md')) {
      const sourcePath = path.join(sourceDir, file);
      const targetPath = path.join(targetDir, file);
      fs.copyFileSync(sourcePath, targetPath);
      console.log(`Copied ${file} to ${targetPath}`);
    }
  });
}

// Function to update menu links
function updateMenuLinks() {
  const configPath = path.join(__dirname, '..', 'build', 'docs', 'gatsby', 'gatsby-config.js');
  
  // Read the config file
  let configContent = fs.readFileSync(configPath, 'utf8');
  
  // Find the menuLinks array
  const menuLinksRegex = /menuLinks:\s*\[([\s\S]*?)\]/;
  const match = configContent.match(menuLinksRegex);
  
  if (match) {
    // Add the new menu link
    const newMenuLink = `      {
        name: 'Contributors Guide',
        link: '/making-changes',
        ignoreNextPrev: true
      },`;
    
    // Remove the 'home' menu link and add the new one
    const menuLinksContent = match[1]
      .replace(/,\s*{\s*name:\s*'home',[^}]*},/g, '');
    
    // Insert the new menu link before the closing bracket
    const updatedContent = configContent.replace(
      menuLinksRegex,
      `menuLinks: [${newMenuLink}${menuLinksContent}]`
    );
    
    // Write the updated content back to the file
    fs.writeFileSync(configPath, updatedContent);
    console.log('Updated menu links in gatsby-config.js');
  } else {
    console.error('Could not find menuLinks array in gatsby-config.js');
  }
}

// Execute the functions
try {
  copyMarkdownFiles();
  updateMenuLinks();
  console.log('Documentation preparation completed successfully!');
} catch (error) {
  console.error('Error preparing documentation:', error);
  process.exit(1);
} 