import { readFile, readdir } from 'node:fs/promises';

import assert from 'node:assert/strict';
import test from 'node:test';

const workflowsDir = new URL('../.github/workflows/', import.meta.url);
const workflowNames = (await readdir(workflowsDir)).filter((name) => /\.ya?ml$/.test(name));
const workflows = new Map(
  await Promise.all(
    workflowNames.map(async (name) => [name, await readFile(new URL(name, workflowsDir), 'utf8')]),
  ),
);

const untrustedPullRequestWorkflows = [
  'commits.yml',
  'go-lint.yml',
  'no-private-snapshot-machinery.yml',
  'preprocessing.yml',
  'pr-checks.yaml',
  'services.yml',
  'test-expiring-todos.yml',
  'test.yml',
  'typescript-lint.yml',
  'versioning.yaml',
  'wasm.yml',
];

// pull_request withholds secrets from forks; keep these workflows isolated from PR code.
const privilegedPullRequestWorkflows = ['review-bypass.yml'];

const reusableUntrustedTestWorkflows = [
  'check-variants.yml',
  'detect-targets.yml',
  'test-cli.yml',
  'test-coverage.yml',
  'test-csharp.yml',
  'test-go.yml',
  'test-javav2.yml',
  'test-mcp-typescript.yml',
  'test-php.yml',
  'test-pythonv2.yml',
  'test-ruby.yml',
  'test-terraform.yml',
  'test-typescriptv2.yml',
];

function workflow(name) {
  const content = workflows.get(name);
  assert.ok(content, `missing workflow ${name}`);
  return content;
}

function hasTrigger(content, trigger) {
  return new RegExp(`^\\s*${trigger}:`, 'm').test(content);
}

function workflowNamesWithTrigger(trigger) {
  return [...workflows]
    .filter(([, content]) => hasTrigger(content, trigger))
    .map(([name]) => name)
    .sort();
}

function unquoteYAMLScalar(value) {
  const match = value.match(/^(['"])(.*)\1$/);
  return match ? match[2] : value;
}

function actionReferences(content) {
  return [...content.matchAll(/^\s*(?:-\s*)?(?:uses|["']uses["'])\s*:\s+((?:["'][^"']*["'])|[^\s#]+)/gm)]
    .map((match) => unquoteYAMLScalar(match[1]));
}

function thirdPartyActionReferences(content) {
  return actionReferences(content).filter((reference) => !reference.startsWith('./'));
}

function workflowSteps(content) {
  const starts = [...content.matchAll(/^\s*-\s+/gm)];
  return starts.map((match, index) => content.slice(match.index, starts[index + 1]?.index));
}

function assertNoWritePermissions(name, content) {
  const lines = content.split('\n');
  for (let index = 0; index < lines.length; index += 1) {
    const permission = lines[index].match(/^(\s*)(?:permissions|["']permissions["'])\s*:\s*(.*?)\s*(?:#.*)?$/);
    if (!permission) {
      continue;
    }

    const [, indent, value] = permission;
    assert.doesNotMatch(
      value,
      /^(?:["']?write-all["']?)$|^\{[^}]*:\s*["']?write["']?(?:\s*(?:,|\}))/,
      `${name} must not request write permissions`,
    );
    if (value !== '') {
      continue;
    }

    for (let next = index + 1; next < lines.length; next += 1) {
      const mapping = lines[next];
      if (mapping.trim() === '') {
        continue;
      }
      if (!mapping.startsWith(`${indent} `)) {
        break;
      }
      assert.doesNotMatch(
        mapping,
        /^\s*(?:[A-Za-z][A-Za-z-]*|["'][^"']+["'])\s*:\s*["']?write["']?\s*(?:#.*)?$/,
        `${name} must not request write permissions`,
      );
    }
  }
}

function assertPinnedThirdPartyActions(name, content) {
  for (const reference of thirdPartyActionReferences(content)) {
    const [, ref] = reference.split('@');
    assert.match(ref ?? '', /^[0-9a-f]{40}$/, `${name} must pin ${reference} to a commit SHA`);
  }
}

function assertReadOnlyPrCodeWorkflow(name, content) {
  assert.doesNotMatch(content, /^\s*pull_request_target:/m, `${name} must not use pull_request_target`);
  assertNoWritePermissions(name, content);
  assert.doesNotMatch(content, /\bpackages:\s*read\b/, `${name} must not access private packages`);
  assert.doesNotMatch(content, /\b(?:BOT_REPO_TOKEN|CHANGELOG_PAT_TOKEN|SNAPSHOT_DISPATCH_TOKEN|NVD_API_KEY|SECURITY_NOTIFICATIONS_SLACK_WEBHOOK)\b/, `${name} must not reference privileged secrets`);

  const checkouts = workflowSteps(content).filter((step) => actionReferences(step).some((reference) => reference.toLowerCase().startsWith('actions/checkout@')));
  for (const checkout of checkouts) {
    assert.match(checkout, /^\s*(?:persist-credentials|["']persist-credentials["'])\s*:\s*(?:false|["']false["'])(?:\s+#.*)?$/m, `${name} must disable checkout credential persistence`);
  }

  assertPinnedThirdPartyActions(name, content);
}

test('every pull-request workflow is read-only and does not persist checkout credentials', () => {
  assert.deepEqual(
    workflowNamesWithTrigger('pull_request'),
    [...untrustedPullRequestWorkflows, ...privilegedPullRequestWorkflows].sort(),
    'add new pull_request workflows to the reviewed workflow policy',
  );

  for (const name of untrustedPullRequestWorkflows) {
    assertReadOnlyPrCodeWorkflow(name, workflow(name));
  }
});

function assertPrivilegedPullRequestWorkflow(name, content) {
  assert.doesNotMatch(content, /^\s*pull_request_target:/m, `${name} must not use pull_request_target`);
  assert.match(content, /^permissions:\s*\{\}\s*$/m, `${name} must drop all workflow token permissions`);
  assertNoWritePermissions(name, content);
  assert.deepEqual(actionReferences(content), [], `${name} must not execute pull-request code via actions`);
  assert.match(
    content,
    /^ {4}if:\s*>-?\s*\n {6}github\.event\.pull_request\.head\.repo\.full_name == github\.repository &&\s*$/m,
    `${name} must guard privileged execution to same-repository pull requests`,
  );
}

test('privileged pull-request workflows never execute pull-request code', () => {
  for (const name of privilegedPullRequestWorkflows) {
    assertPrivilegedPullRequestWorkflow(name, workflow(name));
  }
});

test('privileged pull-request policy rejects unsafe mutations', () => {
  const content = workflow('review-bypass.yml');
  assert.throws(
    () => assertPrivilegedPullRequestWorkflow('target.yml', content.replace(/^  pull_request:/m, '  pull_request_target:')),
    /must not use pull_request_target/,
  );
  assert.throws(
    () => assertPrivilegedPullRequestWorkflow('unguarded.yml', content.replace(/^ {6}github\.event\.pull_request\.head\.repo\.full_name.*\n/m, '')),
    /must guard privileged execution/,
  );
});

test('reusable target-test workflows remain read-only when called from a pull request', () => {
  for (const name of reusableUntrustedTestWorkflows) {
    const content = workflow(name);
    assert.match(content, /^\s*workflow_call:/m, `${name} must only be reusable`);
    assertReadOnlyPrCodeWorkflow(name, content);
  }
});

function assertMetadataOnlyTargetWorkflow(name, content) {
  assert.doesNotMatch(content, /actions\/checkout@/i, `${name} must not check out pull-request code`);
  assert.doesNotMatch(content, /^\s*(?:-\s*)?(?:run|["']run["'])\s*:/m, `${name} must not run repository code`);
  assert.equal(actionReferences(content).some((reference) => reference.startsWith('./')), false, `${name} must not call local actions or workflows`);
  assertPinnedThirdPartyActions(name, content);
}

test('only metadata-only workflows use pull_request_target', () => {
  const targetWorkflows = [...workflows]
    .filter(([, content]) => hasTrigger(content, 'pull_request_target'))
    .map(([name]) => name)
    .sort();

  assert.deepEqual(targetWorkflows, ['cla.yml', 'label-new-prs.yaml', 'snapshot-dispatch.yml']);

  for (const name of targetWorkflows) {
    assertMetadataOnlyTargetWorkflow(name, workflow(name));
  }

  assert.throws(
    () => assertMetadataOnlyTargetWorkflow('target.yml', 'steps:\n  - run: echo unsafe'),
    /must not run repository code/,
  );
  assert.throws(
    () => assertMetadataOnlyTargetWorkflow('quoted-local-action.yml', 'steps:\n  - "uses": "./.github/actions/local"'),
    /must not call local actions or workflows/,
  );

  const snapshot = workflow('snapshot-dispatch.yml');
  assert.match(snapshot, /statuses:\s*write/, 'snapshot status publishing remains explicitly scoped');
  assert.match(snapshot, /SNAPSHOT_DISPATCH_TOKEN/, 'snapshot dispatch remains a separately scoped trusted boundary');

  const labeler = workflow('label-new-prs.yaml');
  assert.match(labeler, /pull-requests:\s*write/, 'labeling remains explicitly scoped');
  assert.match(labeler, /issues:\s*write/, 'labeling remains explicitly scoped');
});

test('trusted workflows remain outside the untrusted PR execution boundary', () => {
  for (const name of ['create-documentation.yaml', 'release.yaml', 'security-scan.yml']) {
    const content = workflow(name);
    assert.doesNotMatch(content, /^\s*pull_request:/m, `${name} must not run on pull_request`);
    assert.doesNotMatch(content, /^\s*pull_request_target:/m, `${name} must not run on pull_request_target`);
  }
});

test('versioning ignores workflow-only changes without excluding generator code', () => {
  const versioning = workflow('versioning.yaml');
  assert.match(versioning, /cmd\/!\(changelog\|changeset\)\/\*\*/, 'versioning must exclude changelog commands in the positive cmd glob');
  assert.match(versioning, /internal\/!\(changeset\)\/\*\*/, 'versioning must exclude changeset internals in the positive internal glob');
  assert.doesNotMatch(versioning, /- 'cmd\/\*\*'\n\s*- '!cmd\//, 'versioning must not use separate negated command paths with an any predicate');
  assert.doesNotMatch(versioning, /- 'internal\/\*\*'\n\s*- '!internal\//, 'versioning must not use separate negated internal paths with an any predicate');
});

test('untrusted-workflow policy rejects supported write permissions and delayed checkout settings', () => {
  const base = `on:\n  pull_request:\npermissions:\n  contents: read\njobs:\n  test:\n    steps:\n      - uses: actions/checkout@${'a'.repeat(40)}\n        with:\n          fetch-depth: 0\n          sparse-checkout: |\n            scripts\n          persist-credentials: false\n`;

  assert.doesNotThrow(() => assertReadOnlyPrCodeWorkflow('safe.yml', base));
  assert.throws(
    () => assertReadOnlyPrCodeWorkflow('write-all.yml', base.replace('permissions:', 'permissions: write-all')),
    /must not request write permissions/,
  );
  assert.throws(
    () => assertReadOnlyPrCodeWorkflow('quoted-write.yml', base.replace('contents: read', 'contents: "write"')),
    /must not request write permissions/,
  );
  assert.throws(
    () => assertReadOnlyPrCodeWorkflow('quoted-key-write.yml', base.replace('contents: read', '"contents": write')),
    /must not request write permissions/,
  );
  assert.throws(
    () => assertReadOnlyPrCodeWorkflow('inline-write.yml', base.replace('permissions:\n  contents: read', 'permissions: { contents: write }')),
    /must not request write permissions/,
  );
  const quotedCheckout = `on:\n  pull_request:\npermissions:\n  contents: read\njobs:\n  test:\n    steps:\n      - "uses": "Actions/checkout@${'a'.repeat(40)}"\n        with:\n          "persist-credentials": false\n`;
  assert.doesNotThrow(() => assertReadOnlyPrCodeWorkflow('quoted-checkout.yml', quotedCheckout));
  assert.throws(
    () => assertReadOnlyPrCodeWorkflow('quoted-persisted-checkout.yml', quotedCheckout.replace('"persist-credentials": false', '"persist-credentials": true')),
    /must disable checkout credential persistence/,
  );
  assert.throws(
    () => assertReadOnlyPrCodeWorkflow('persisted-checkout.yml', base.replace('persist-credentials: false', 'persist-credentials: true')),
    /must disable checkout credential persistence/,
  );
  assert.throws(
    () => assertPinnedThirdPartyActions('target.yml', 'steps:\n  - uses: actions/labeler@v6'),
    /must pin actions\/labeler@v6 to a commit SHA/,
  );
  assert.throws(
    () => assertPinnedThirdPartyActions('quoted-key-target.yml', 'steps:\n  - "uses": actions/labeler@v6'),
    /must pin actions\/labeler@v6 to a commit SHA/,
  );
  assert.doesNotThrow(() => assertPinnedThirdPartyActions('quoted-action.yml', `steps:\n  - uses: "actions/labeler@${'a'.repeat(40)}"`));
  assert.doesNotThrow(() => assertPinnedThirdPartyActions('quoted-local-action.yml', 'steps:\n  - uses: "./.github/actions/local"'));
});

test('selected test variants cannot be reported as green when a leaf is skipped', () => {
  const variantAggregate = workflow('check-variants.yml');
  assert.match(variantAggregate, /\.result != "success"/, 'variant aggregate must reject skipped selected leaves');
  assert.doesNotMatch(variantAggregate, /allowed-skips/, 'variant aggregate must not permit broad skipped leaves');

  const testWorkflow = workflow('test.yml');
  assert.doesNotMatch(testWorkflow, /github\.event\.pull_request\.draft == false/, 'selected test variants must run for draft pull requests too');
  assert.match(testWorkflow, /- detect-targets/, 'top-level aggregate must receive selected-target outputs');
  assert.match(testWorkflow, /RUN_GO:\s*\$\{\{ needs\.detect-targets\.outputs\.run-go \}\}/, 'top-level aggregate must gate required Go jobs on target selection');
  assert.match(testWorkflow, /require_success detect-targets/, 'top-level aggregate must require target detection to succeed');
  assert.match(testWorkflow, /require_success test-go/, 'top-level aggregate must require selected target jobs to succeed');
  assert.doesNotMatch(testWorkflow, /re-actors\/alls-green@/, 'top-level aggregate must not broadly allow skipped jobs');
  assert.doesNotMatch(testWorkflow, /allowed-skips/, 'top-level aggregate must not permit broad skipped jobs');
});
