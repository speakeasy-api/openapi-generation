import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

export function finalizeReleaseChangelog(content, source, tag) {
  if (!/^[0-9a-f]{40}$/.test(source) || !/^v\d+\.\d+\.\d+$/.test(tag)) {
    throw new Error('Expected a full commit SHA and a release version');
  }
  const lines = content.split('\n');
  const heading = lines.findIndex((line) => line.startsWith('## '));
  if (heading < 0 || !lines[heading].startsWith(`## [${source}] - `)) {
    throw new Error('The first changelog entry does not match the release source');
  }
  const links = lines.flatMap((line, index) => line.startsWith(`[${source}]: `) ? [index] : []);
  if (links.length !== 1 || !lines[links[0]].endsWith(`...${source}`)) {
    throw new Error('Expected one comparison link for the release source');
  }
  if (lines.some((line) => line.startsWith(`## [${tag}]`) || line.startsWith(`[${tag}]: `))) {
    throw new Error('The release version already exists in the changelog');
  }
  lines[heading] = lines[heading].replace(`[${source}]`, `[${tag}]`);
  lines[links[0]] = lines[links[0]].replace(`[${source}]`, `[${tag}]`).replace(`...${source}`, `...${tag}`);
  return lines.join('\n');
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  const path = process.argv[2] ?? 'CHANGELOG.md';
  const content = await readFile(path, 'utf8');
  await writeFile(path, finalizeReleaseChangelog(content, process.env.RELEASE_SOURCE, process.env.RELEASE_TAG));
}
