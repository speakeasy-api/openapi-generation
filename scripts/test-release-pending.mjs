import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const script = fileURLToPath(new URL('./has-unreleased-commits.sh', import.meta.url));

test('releases new commits and skips commits already tagged', () => {
  const directory = mkdtempSync(join(tmpdir(), 'release-test-'));
  const options = { cwd: directory, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] };
  const git = (...args) => execFileSync('git', ['-c', 'commit.gpgsign=false', '-c', 'tag.gpgsign=false', '-c', 'core.hooksPath=/dev/null', ...args], options);
  const pending = () => execFileSync('bash', [script], options).trim();
  try {
    git('init');
    git('config', 'user.name', 'Test Author');
    git('config', 'user.email', 'author@example.com');
    git('commit', '--allow-empty', '-m', 'initial');
    assert.equal(pending(), 'has_commits=true');
    git('tag', '-a', 'v1.0.0', '-m', 'v1.0.0');
    assert.equal(pending(), 'has_commits=false');
    git('commit', '--allow-empty', '-m', 'fix: update behavior');
    assert.equal(pending(), 'has_commits=true');
    git('tag', 'v1.0.1');
    assert.equal(pending(), 'has_commits=false');
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});
