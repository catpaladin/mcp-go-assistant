module.exports = {
  types: [
    { value: 'feat', name: 'feat', description: 'A new feature' },
    { value: 'fix', name: 'fix', description: 'A bug fix' },
    { value: 'docs', name: 'docs', description: 'Documentation only changes' },
    { value: 'style', name: 'style', description: 'Changes that do not affect the meaning of the code' },
    { value: 'refactor', name: 'refactor', description: 'A code change that neither fixes a bug nor adds a feature' },
    { value: 'perf', name: 'perf', description: 'A code change that improves performance' },
    { value: 'test', name: 'test', description: 'Adding missing tests or correcting existing tests' },
    { value: 'build', name: 'build', description: 'Changes that affect the build system or external dependencies' },
    { value: 'ci', name: 'ci', description: 'Changes to CI configuration files and scripts' },
    { value: 'chore', name: 'chore', description: 'Other changes that do not modify src or test files' },
    { value: 'revert', name: 'revert', description: 'Reverts a previous commit' },
  ],

  scopes: [{ name: 'codereview' }, { name: 'godoc' }, { name: 'testgen' }, { name: 'ratelimit' }, { name: 'retry' }, { name: 'circuitbreaker' }, { name: 'metrics' }, { name: 'logging' }, { name: 'health' }, { name: 'config' }, { name: 'validations' }],

  allowCustomScopes: true,
  allowBreakingChanges: ['feat', 'fix', 'perf', 'refactor'],

  messages: {
    type: 'Select the type of change that you are committing:',
    scope: 'Denote the scope of this change:',
    subject: 'Write a short, imperative mood description of the change:\n',
    body: 'Provide a longer description of the change:\n ',
    breaking: 'List any BREAKING CHANGES:\n',
    footer: 'List any ISSUES CLOSED by this change:\n ',
    confirmCommit: 'Are you sure you want to proceed with the commit above?',
  },
};
