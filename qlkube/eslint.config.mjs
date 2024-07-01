import globals from 'globals';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import js from '@eslint/js';
import { FlatCompat } from '@eslint/eslintrc';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default [...compat.extends('airbnb-base', 'plugin:node/recommended'), {
  languageOptions: {
    globals: {
      ...globals.commonjs,
      ...globals.node,
      Atomics: 'readonly',
      SharedArrayBuffer: 'readonly',
    },

    ecmaVersion: 2018,
    sourceType: 'commonjs',
    parserOptions: {
      ecmaVersion: 2020
    }
  },

  ignores: ['**/*.mjs'],

  rules: {
    'no-unused-vars': ['warn', {
      argsIgnorePattern: '^_',
      varsIgnorePattern: '^_',
      caughtErrorsIgnorePattern: '^_',
    }],
    'no-restricted-syntax': 'off',
    'no-param-reassign': 'off',
    'no-plusplus': 'off',
    'no-await-in-loop': 'off',
    'radix': 'off',
    'class-methods-use-this': 'off',
  },
}];
