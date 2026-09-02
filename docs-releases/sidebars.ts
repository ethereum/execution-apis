import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';
import {readdirSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import semver from 'semver';

// Order release docs by descending semver precedence (newest on top). Built
// explicitly here rather than via sidebar_position frontmatter so arbitrary
// prereleases sort correctly and new release files are picked up automatically.
const dir = fileURLToPath(new URL('.', import.meta.url));
const releaseDocs = readdirSync(dir)
  .filter((f) => f.endsWith('.md') && f !== 'index.md')
  .map((f) => f.replace(/\.md$/, ''))
  .filter((id) => semver.valid(id.replace(/^v/, '')))
  .sort((a, b) => semver.rcompare(a, b));

const sidebars: SidebarsConfig = {
  releasesSidebar: ['index', ...releaseDocs],
};

export default sidebars;
