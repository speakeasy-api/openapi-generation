import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const workflowURL = new URL('../.github/workflows/snapshot-dispatch.yml', import.meta.url);
const workflow = await readFile(workflowURL, 'utf8');
const scriptBlocks = [...workflow.matchAll(/^\s+script: \|$/gm)];
assert.equal(scriptBlocks.length, 1, 'workflow must contain exactly one inline github-script program');
const scriptMarker = scriptBlocks[0];
const scriptLines = workflow
  .slice(scriptMarker.index + scriptMarker[0].length)
  .replace(/^\n/, '')
  .split('\n');
const scriptIndent = scriptLines.find((line) => line.trimStart().startsWith('const receiver'))?.match(/^\s*/)?.[0];
assert.ok(scriptIndent, 'workflow must indent its inline github-script program');
const scriptEnd = scriptLines.findIndex((line) => line.trim() !== '' && !line.startsWith(scriptIndent));
const script = scriptLines
  .slice(0, scriptEnd === -1 ? undefined : scriptEnd)
  .map((line) => line.startsWith(scriptIndent) ? line.slice(scriptIndent.length) : line)
  .join('\n');
const runScript = new Function(
  'github',
  'context',
  'core',
  'process',
  'fetch',
  `return (async () => {\n${script}\n})();`,
);

const archiveRepository = {
  id: 525818195,
  full_name: 'speakeasy-api/openapi-generation-archive',
};
const canonicalRepository = {
  id: 1364456527,
  full_name: 'speakeasy-api/openapi-generation-release',
};
const privateRepository = canonicalRepository;
const basePullRequest = {
  number: 123,
  state: 'open',
  merged: false,
  merge_commit_sha: null,
  author_association: 'CONTRIBUTOR',
  labels: [],
  base: {
    ref: 'main',
    sha: 'b'.repeat(40),
    repo: privateRepository,
  },
  head: {
    ref: 'feature',
    sha: 'a'.repeat(40),
    repo: {
      id: 42,
      full_name: 'contributor/openapi-generation',
    },
  },
};

async function execute({
  action = 'labeled',
  eventName = 'pull_request_target',
  pullRequest = {},
  repository = {},
  comments = [],
  approvalComment = '',
  changedLabel = 'go',
  contextRepo = {},
  token = 'configured-without-exposing-a-value',
  responseOK = true,
} = {}) {
  const statuses = [];
  const requests = [];
  const infos = [];
  const apiCalls = [];
  const pr = {
    ...basePullRequest,
    ...pullRequest,
    base: { ...basePullRequest.base, ...pullRequest.base },
    head: { ...basePullRequest.head, ...pullRequest.head },
  };
  const repo = { ...privateRepository, ...repository };
  const github = {
    rest: {
      repos: {
        createCommitStatus: async (status) => {
          apiCalls.push('repos.createCommitStatus');
          statuses.push(status);
        },
      },
      issues: {
        listComments: async () => {
          apiCalls.push('issues.listComments');
          return { data: comments };
        },
      },
      pulls: {
        get: async () => {
          apiCalls.push('pulls.get');
          return { data: pr };
        },
      },
    },
    paginate: async (request, options) => (await request(options)).data,
  };
  const context = {
    actor: 'contributor',
    eventName,
    repo: { owner: 'speakeasy-api', repo: 'openapi-generation-release', ...contextRepo },
    runId: 987,
    payload: eventName === 'issue_comment' ? {
      action,
      repository: repo,
      issue: { number: pr.number, pull_request: {} },
      comment: { body: approvalComment, author_association: 'MEMBER' },
    } : {
      action,
      repository: repo,
      pull_request: pr,
      label: { name: changedLabel },
    },
  };
  const fetch = async (url, options) => {
    const body = JSON.parse(options.body);
    requests.push({ url, options, body });
    const payloadWithinGitHubLimit = Object.keys(body.client_payload).length <= 10;
    const ok = responseOK && payloadWithinGitHubLimit;
    return { ok, status: ok ? 204 : 422 };
  };

  const result = { statuses, requests, infos, apiCalls };
  try {
    await runScript(
      github,
      context,
      { info: (message) => infos.push(message) },
      { env: { SNAPSHOT_DISPATCH_TOKEN: token } },
      fetch,
    );
  } catch (error) {
    error.result = result;
    throw error;
  }
  return result;
}

function labels(...names) {
  return names.map((name) => ({ name }));
}

async function captureRejection(promise, pattern) {
  try {
    await promise;
  } catch (error) {
    assert.match(error.message, pattern);
    return error;
  }
  assert.fail('expected promise to reject');
}

test('workflow is metadata-only and least-privileged', () => {
  assert.match(workflow, /pull_request_target:/);
  assert.match(workflow, /github\.repository_id == '1364456527' && github\.repository == 'speakeasy-api\/openapi-generation-release'/);
  assert.doesNotMatch(workflow, /github\.repository_id == '525818195'/);
  assert.doesNotMatch(workflow, /github\.repository_id == '1327468509'/);
  assert.match(workflow, /permission-contents: write/);
  assert.match(workflow, /repositories: openapi-generation-snapshots/);
  assert.match(workflow, /SNAPSHOT_DISPATCH_TOKEN: \$\{\{ steps\.app-token\.outputs\.token \}\}/);
  assert.match(workflow, /statuses: write/);
  assert.match(workflow, /contents: read/);
  assert.match(workflow, /cancel-in-progress: false/);
  assert.doesNotMatch(workflow, /actions\/checkout@/);
  assert.doesNotMatch(workflow, /^\s+run:/m);
  assert.doesNotMatch(workflow, /uses:\s+\.\//);
  for (const match of workflow.matchAll(/uses:\s+[^@\s]+@([^\s#]+)/g)) {
    assert.match(match[1], /^[0-9a-f]{40}$/);
  }
});

test('no selector and selector removal complete without dispatch', async () => {
  for (const action of ['opened', 'unlabeled']) {
    const result = await execute({ action });
    assert.equal(result.requests.length, 0);
    assert.equal(result.statuses.at(-1).state, 'success');
    assert.match(result.statuses.at(-1).description, /No snapshot selectors/);
  }
});

test('fork approval is bound to the exact current head SHA', async () => {
  const stale = await execute({
    pullRequest: { labels: labels('go') },
    comments: [{ author_association: 'MEMBER', body: `/snapshot approve ${'c'.repeat(40)}` }],
  });
  assert.equal(stale.requests.length, 0);
  assert.match(stale.statuses.at(-1).description, /exact SHA/);

  const approved = await execute({
    pullRequest: { labels: labels('go') },
    comments: [{
      author_association: 'MEMBER',
      body: `reviewed\n/snapshot approve ${basePullRequest.head.sha}`,
    }],
  });
  assert.equal(approved.requests.length, 1);
  assert.equal(approved.statuses.at(-1).state, 'pending');
  assert.equal(approved.requests[0].body.event_type, 'snapshot_requested');
  const payload = approved.requests[0].body.client_payload;
  assert.equal(payload.schema_version, 2);
  assert.equal(payload.generator_repo, privateRepository.full_name);
  assert.equal(payload.generator_repository_id, privateRepository.id);
  assert.equal(payload.pr.head_sha, basePullRequest.head.sha);
  assert.deepEqual(payload.selectors, ['snapshot-go']);
  assert.deepEqual(Object.keys(payload).sort(), [
    'comment_mode',
    'event',
    'generator_repo',
    'generator_repository_id',
    'idempotency_key',
    'is_fork',
    'pr',
    'public_run_id',
    'schema_version',
    'selectors',
  ]);
  assert.equal(Object.keys(payload).length, 10);
});

test('established and transition target labels normalize to the same selector', async () => {
  const internalHead = { repo: privateRepository };
  for (const label of ['go', 'snapshot-go']) {
    const result = await execute({
      changedLabel: label,
      pullRequest: { labels: labels(label), head: internalHead },
    });
    assert.equal(result.requests.length, 1);
    assert.deepEqual(result.requests[0].body.client_payload.selectors, ['snapshot-go']);
  }
});

test('an exact approval comment triggers reevaluation using API PR metadata', async () => {
  const result = await execute({
    eventName: 'issue_comment',
    approvalComment: `/snapshot approve ${basePullRequest.head.sha}`,
    pullRequest: { labels: labels('snapshot-go') },
    comments: [{
      author_association: 'MEMBER',
      body: `/snapshot approve ${basePullRequest.head.sha}`,
    }],
  });
  assert.equal(result.requests.length, 1);
  assert.equal(result.requests[0].body.event_type, 'snapshot_requested');
});

test('stale approval-shaped comments do not trigger reevaluation', async () => {
  const result = await execute({
    eventName: 'issue_comment',
    action: 'created',
    approvalComment: `/snapshot approve ${'c'.repeat(40)}`,
    pullRequest: { labels: labels('snapshot-go') },
    comments: [{
      author_association: 'MEMBER',
      body: `/snapshot approve ${basePullRequest.head.sha}`,
    }],
  });
  assert.equal(result.requests.length, 0);
  assert.equal(result.statuses.length, 0);
});

test('unrelated labels and unknown snapshot-like labels do not dispatch', async () => {
  const unrelated = await execute({
    pullRequest: { labels: labels('go', 'skip linear') },
    changedLabel: 'skip linear',
  });
  assert.equal(unrelated.requests.length, 0);
  assert.equal(unrelated.statuses.length, 0);

  const unknown = await execute({
    pullRequest: { labels: labels('snapshot-wip') },
    changedLabel: 'snapshot-wip',
  });
  assert.equal(unknown.requests.length, 0);
  assert.equal(unknown.statuses.length, 0);

  const customerLike = await execute({
    pullRequest: { labels: labels('business-go') },
    changedLabel: 'business-go',
  });
  assert.equal(customerLike.requests.length, 0);
  assert.equal(customerLike.statuses.length, 0);
});

test('only exact private repository identity pairs are active for every event shape', async () => {
  const repository = {
    id: 1273846395,
    full_name: 'speakeasy-api/openapi-generation-oss',
  };
  const contextRepo = { repo: 'openapi-generation-oss' };
  const approvedComment = `/snapshot approve ${basePullRequest.head.sha}`;

  for (const options of [
    {
      repository: { id: canonicalRepository.id, full_name: archiveRepository.full_name },
      contextRepo: { repo: 'openapi-generation-archive' },
      pullRequest: { labels: labels('snapshot-go') },
    },
    {
      repository: { id: archiveRepository.id, full_name: canonicalRepository.full_name },
      contextRepo: { repo: 'openapi-generation-release' },
      pullRequest: { labels: labels('snapshot-go') },
    },
    {
      repository: { id: undefined, full_name: 'speakeasy-api/openapi-generation-oss' },
      contextRepo,
      pullRequest: { labels: labels('snapshot-go') },
    },
    {
      repository,
      contextRepo,
      pullRequest: { labels: labels('snapshot-go') },
    },
    {
      eventName: 'issue_comment',
      repository,
      contextRepo,
      approvalComment: approvedComment,
      pullRequest: { labels: labels('snapshot-go') },
      comments: [{ author_association: 'MEMBER', body: approvedComment }],
    },
    {
      action: 'closed',
      repository,
      contextRepo,
      pullRequest: { merged: true, merge_commit_sha: 'd'.repeat(40) },
    },
  ]) {
    const result = await execute(options);
    assert.equal(result.requests.length, 0);
    assert.equal(result.statuses.length, 0);
    assert.deepEqual(result.apiCalls, []);
    assert.deepEqual(result.infos, ['Snapshot dispatch is inactive for this repository identity']);
  }
});

test('the canonical repository identity dispatches trusted metadata', async () => {
  const result = await execute({
    repository: canonicalRepository,
    contextRepo: { repo: 'openapi-generation-release' },
    pullRequest: {
      labels: labels('snapshot-go'),
      base: { repo: canonicalRepository },
      head: { repo: canonicalRepository },
    },
  });
  assert.equal(result.requests.length, 1);
  assert.equal(result.requests[0].body.client_payload.generator_repo, canonicalRepository.full_name);
  assert.equal(result.requests[0].body.client_payload.generator_repository_id, canonicalRepository.id);
});

test('snapshot-all on a fork requires exact-SHA maintainer approval', async () => {
  const unapproved = await execute({
    pullRequest: { labels: labels('snapshot-all') },
    changedLabel: 'snapshot-all',
  });
  assert.equal(unapproved.requests.length, 0);
  assert.match(unapproved.statuses.at(-1).description, /exact SHA/);

  const approved = await execute({
    pullRequest: { labels: labels('snapshot-all') },
    changedLabel: 'snapshot-all',
    comments: [{
      author_association: 'MEMBER',
      body: `/snapshot approve ${basePullRequest.head.sha}`,
    }],
  });
  assert.equal(approved.requests.length, 1);
  assert.deepEqual(approved.requests[0].body.client_payload.selectors, ['snapshot-all']);
});

test('missing token and receiver rejection fail the public status', async () => {
  const internalHead = { repo: privateRepository };
  const missingToken = await captureRejection(
    execute({ pullRequest: { labels: labels('snapshot-go'), head: internalHead }, token: '' }),
    /not configured/,
  );
  assert.equal(missingToken.result.statuses.at(-1).state, 'failure');

  const rejected = await captureRejection(
    execute({ pullRequest: { labels: labels('snapshot-go'), head: internalHead }, responseOK: false }),
    /rejected with HTTP 422/,
  );
  assert.equal(rejected.result.statuses.at(-1).state, 'failure');
});

test('unavailable head repositories and unsafe refs fail closed', async () => {
  const missingHead = await captureRejection(
    execute({ pullRequest: { labels: labels('snapshot-go'), head: { repo: null } } }),
    /metadata is unavailable/,
  );
  assert.equal(missingHead.result.statuses.at(-1).state, 'failure');

  const unsafeRef = await captureRejection(
    execute({ pullRequest: { labels: labels('snapshot-go'), head: { ref: 'feature~unsafe', repo: privateRepository } } }),
    /receiver contract/,
  );
  assert.equal(unsafeRef.result.statuses.at(-1).state, 'failure');
});

test('closed PRs cannot initiate new snapshot requests', async () => {
  const result = await execute({
    eventName: 'issue_comment',
    action: 'created',
    approvalComment: `/snapshot approve ${basePullRequest.head.sha}`,
    pullRequest: { state: 'closed', labels: labels('snapshot-go') },
    comments: [{
      author_association: 'MEMBER',
      body: `/snapshot approve ${basePullRequest.head.sha}`,
    }],
  });
  assert.equal(result.requests.length, 0);
  assert.equal(result.statuses.length, 0);
});

test('merged PR dispatches identity-keyed cleanup metadata', async () => {
  const result = await execute({
    action: 'closed',
    pullRequest: {
      merged: true,
      merge_commit_sha: 'd'.repeat(40),
    },
  });
  assert.equal(result.requests.length, 1);
  assert.equal(result.requests[0].body.event_type, 'snapshot_pr_merged');
  assert.deepEqual(result.requests[0].body.client_payload, {
    schema_version: 2,
    event: 'snapshot_pr_merged',
    generator_repo: privateRepository.full_name,
    generator_repository_id: privateRepository.id,
    pr: {
      number: basePullRequest.number,
      head_sha: basePullRequest.head.sha,
      merge_commit_sha: 'd'.repeat(40),
    },
  });
});
