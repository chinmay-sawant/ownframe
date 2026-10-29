// Flat config for the website. ESLint 10 dropped the legacy .eslintrc
// format, so every rule lives in this file.
//
// `eslint-plugin-react` still declares a peer range that stops at ESLint 9,
// and its version detection calls an API ESLint 10 removed. Pinning
// `settings.react.version` skips that detection, and the `overrides` entry in
// package.json resolves the peer range to the installed ESLint.
import js from '@eslint/js';
import globals from 'globals';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';

export default [
  {
    ignores: ['dist/', 'node_modules/', '.vite/', 'public/'],
  },
  js.configs.recommended,
  {
    files: ['**/*.{js,jsx,mjs}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: { ...globals.browser, ...globals.node },
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    plugins: { react, 'react-hooks': reactHooks },
    settings: { react: { version: '19.2' } },
    rules: {
      ...react.configs.flat.recommended.rules,
      ...reactHooks.configs.flat.recommended.rules,
      // The site has no PropTypes layer, so the prop-types rule only
      // reports the same components over and over.
      'react/prop-types': 'off',
      // This rule comes from the React Compiler ruleset, and the site does
      // not run the compiler. Two effects here reset state on purpose: one
      // stops a preview when the reader asks for reduced motion, the other
      // closes the menu and the lightbox when the route changes.
      'react-hooks/set-state-in-effect': 'off',
    },
  },
];
