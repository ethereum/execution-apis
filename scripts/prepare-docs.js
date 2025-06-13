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

// Function to copy Gatsby config file
function copyGatsbyConfig() {
  const sourcePath = path.join(__dirname, '..', 'docs', 'config', 'gatsby-config.js');
  const targetPath = path.join(__dirname, '..', 'build', 'docs', 'gatsby', 'gatsby-config.js');

  // Ensure target directory exists
  ensureDirectoryExists(path.dirname(targetPath));

  // Copy the file
  fs.copyFileSync(sourcePath, targetPath);
  console.log(`Copied gatsby-config.js to ${targetPath}`);
}

// Execute the functions
try {
  copyMarkdownFiles();
  copyGatsbyConfig();
  console.log('Documentation preparation completed successfully!');
} catch (error) {
  console.error('Error preparing documentation:', error);
  process.exit(1);
} 