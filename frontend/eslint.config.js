import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import reactPlugin from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import prettierPlugin from 'eslint-plugin-prettier';
import prettierConfig from 'eslint-config-prettier';
import { FlatCompat } from '@eslint/eslintrc';

const compat = new FlatCompat({ baseDirectory: import.meta.dirname });

export default [
	js.configs.recommended,
	...tseslint.configs.recommended,
	...compat.extends('plugin:react/recommended'),
	{
		files: ['**/*.{ts,tsx,js,jsx}'],
		languageOptions: {
			parser: tseslint.parser,
			parserOptions: { ecmaFeatures: { jsx: true } },
			globals: globals.browser,
		},
		plugins: {
			react: reactPlugin,
			'react-hooks': reactHooks,
			prettier: prettierPlugin,
		},
		rules: {
			...prettierConfig.rules,
			'prettier/prettier': 'error',
			'react/prop-types': 'off',
			'react/react-in-jsx-scope': 'off',
			indent: ['error', 'tab'],
		},
		settings: { react: { version: 'detect' } },
	},
];
