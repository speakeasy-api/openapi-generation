import assert from 'node:assert/strict';
import test from 'node:test';
import { finalizeReleaseChangelog } from './finalize-release-changelog.mjs';

const source = 'a'.repeat(40);
const input = `# Changelog

## [${source}] - 2026-09-10
- Fix release metadata (commit ${source}).

## [v1.0.0] - 2026-09-09
- Initial version.


[v1.0.0]: https://example.com/releases/v1.0.0
[${source}]: https://example.com/compare/v1.0.0...${source}`;

test('names the release entry and comparison link without rewriting commit references', () => {
  const result = finalizeReleaseChangelog(input, source, 'v1.0.1');
  assert.match(result, /## \[v1\.0\.1\] - 2026-09-10/);
  assert.match(result, /\[v1\.0\.1\]: https:\/\/example\.com\/compare\/v1\.0\.0\.\.\.v1\.0\.1$/);
  assert.ok(result.includes(`(commit ${source})`));
  assert.ok(result.includes('\n\n\n[v1.0.0]:'));
});

test('rejects stale input, malformed comparison links and an occupied version', () => {
  assert.throws(() => finalizeReleaseChangelog(input, 'b'.repeat(40), 'v1.0.1'), /first changelog entry/);
  assert.throws(() => finalizeReleaseChangelog(input.replace(`...${source}`, '...main'), source, 'v1.0.1'), /comparison link/);
  assert.throws(() => finalizeReleaseChangelog(input, source, 'v1.0.0'), /already exists/);
  assert.throws(() => finalizeReleaseChangelog(input, source, 'main'), /release version/);
});
