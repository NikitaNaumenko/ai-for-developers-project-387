// Turn the JSON reports in a Lighthouse run directory into a markdown scores table.
// Used by scripts/lighthouse.sh for both the terminal output and the GitHub job summary.
import { readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';

const [runDir, url] = process.argv.slice(2);
if (!runDir) {
  console.error('usage: node scripts/lighthouse-summary.mjs <run-dir> [url]');
  process.exit(1);
}

// Lighthouse appends `.report.json` when several output formats share one --output-path.
const reports = readdirSync(runDir)
  .filter((file) => file.endsWith('.json'))
  .sort();

if (reports.length === 0) {
  console.error(`no Lighthouse JSON report found in ${runDir}`);
  process.exit(1);
}

const CATEGORIES = [
  ['performance', 'Performance'],
  ['accessibility', 'Accessibility'],
  ['best-practices', 'Best Practices'],
  ['seo', 'SEO'],
];

const score = (report, key) => {
  const value = report.categories?.[key]?.score;
  return typeof value === 'number' ? String(Math.round(value * 100)) : 'n/a';
};

const rows = reports.map((file) => {
  const report = JSON.parse(readFileSync(path.join(runDir, file), 'utf8'));
  const preset = path.basename(file).split('.')[0];
  return [preset, ...CATEGORIES.map(([key]) => score(report, key))];
});

const timestamp = new Date().toISOString().replace('T', ' ').slice(0, 16);

const lines = [
  `## Lighthouse — ${timestamp} UTC`,
  '',
  ...(url ? [`Target: ${url}`, ''] : []),
  `| Form factor | ${CATEGORIES.map(([, label]) => label).join(' | ')} |`,
  `| --- | ${CATEGORIES.map(() => '---').join(' | ')} |`,
  ...rows.map((row) => `| ${row.join(' | ')} |`),
  '',
  'The full HTML report is attached to this run as an artifact.',
];

console.log(lines.join('\n'));
