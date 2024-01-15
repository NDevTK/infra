// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

module.exports = {
  env: {
    browser: true,
    es2021: true,
  },
  plugins: [
    'react',
    '@typescript-eslint',
    'prettier',
    'jsx-a11y',
    'import'
  ],
  extends: [
    'eslint:recommended',
    'plugin:react/recommended',
    'plugin:react/jsx-runtime',
    'plugin:react-hooks/recommended',
    'google',
    'plugin:@typescript-eslint/recommended',
    'plugin:import/recommended',
    'plugin:import/typescript',
    'plugin:jsx-a11y/recommended',
    'plugin:prettier/recommended',
    'plugin:storybook/recommended'
  ],
  settings: {
    react: {
      version: 'detect',
    },
    'import/parsers': {
      '@typescript-eslint/parser': ['.ts', '.tsx'],
    },
    'import/resolver': {
      typescript: {},
    },
  },
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  rules: {
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        // Use cases:
        // - declare a function property with a default value
        // - ignore some function parameters when writing a callback function
        // See http://b/182855639.
        argsIgnorePattern: '^_',
        // Use cases:
        // - explicitly ignore some elements from a destructed array
        // - explicitly ignore some inferred type parameters
        // See http://b/182855639.
        varsIgnorePattern: '^_',
      },
    ],

    // Code generated from protobuf may contain '_' in the identifier name,
    // (e.g. `BuilderMask_BuilderMaskType`) and therefore triggering the error.
    // `"ignoreImports": true` doesn't fix the issue because checks are still
    // applied where the imported symbol is used.
    //
    // Since this rule doesn't add a lot of value (it only checks whether there
    // are '_' in non-leading/trailing positions), disable it to reduce noise.
    //
    // Note that we should still generally use camelcase.
    camelcase: 0,

    // Group internal dependencies together.
    'import/order': [
      'error',
      {
        pathGroups: [
          {
            pattern: '@root/**',
            group: 'external',
            position: 'after',
          },
          {
            pattern: '@/**',
            group: 'external',
            position: 'after',
          },
        ],
        alphabetize: {
          order: 'asc',
          orderImportKind: 'asc',
        },
        'newlines-between': 'always',
      },
    ],

    'no-restricted-imports': [
      'error',
      {
        patterns: [
          {
            group: ['lodash-es'],
            importNames: ['chain'],
            message: '`chain` from `lodash-es` does not work with tree-shaking',
          },
          {
            group: ['lodash-es/chain'],
            importNames: ['default'],
            message: '`chain` from `lodash-es` does not work with tree-shaking',
          },
        ],
      },
    ],

    'no-console': ['error'],

    // Modify the prettier config to make it match the eslint rule from other
    // presets better.
    'prettier/prettier': [
      'error',
      {
        singleQuote: true,
      },
    ],

    // Ban the usage of `dangerouslySetInnerHTML`.
    //
    // Note that this rule does not catch the usage of `dangerouslySetInnerHTML`
    // in non-native components [1].
    // [1]: https://github.com/jsx-eslint/eslint-plugin-react/issues/3434
    'react/no-danger': ['error'],

    // See https://emotion.sh/docs/eslint-plugin-react.
    'react/no-unknown-property': ['error', { ignore: ['css'] }],

    // JSDoc related rules are deprecated [1].
    // Also with TypeScript, a lot of the JSDoc are unnecessary.
    // [1]: https://eslint.org/blog/2018/11/jsdoc-end-of-life/
    'require-jsdoc': 0,
    'valid-jsdoc': 0,
  },
  overrides: [
    {
      files: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
      plugins: ['jest'],
      extends: ['plugin:jest/recommended'],
    },
    {
      files: ['src/**/*.test.ts', 'src/**/*.test.tsx', '**/testing_tools/**'],
      rules: {
        // Allow assertion to make it easier to write test cases.
        // All incorrect assertion will be caught during test execution anyway.
        '@typescript-eslint/no-non-null-assertion': 0,

        // It's very common to use an empty mock implementation in tests.
        '@typescript-eslint/no-empty-function': 0,

        // Don't need to restrict imports in test files.
        'no-restricted-imports': 0,
      },
    },
  ],
}