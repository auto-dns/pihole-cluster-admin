import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import reactPlugin from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import prettierPlugin from 'eslint-plugin-prettier';
import eslintConfigPrettier from 'eslint-config-prettier'; // ← import the config, not its rules
import { FlatCompat } from '@eslint/eslintrc';

const compat = new FlatCompat({ baseDirectory: import.meta.dirname });

export default [
	js.configs.recommended,
	...tseslint.configs.recommended,
	...compat.extends('plugin:react/recommended'),
	eslintConfigPrettier, // ← disables rules that conflict with Prettier
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
			// Prettier runs as an ESLint error
			'prettier/prettier': 'error',

			// turn OFF ESLint's own formatting rules
			indent: 'off',
			'jsx-quotes': 'off',

			// the rest of your non-formatting rules
			'react/prop-types': 'off',
			'react/react-in-jsx-scope': 'off',
			'no-unused-vars': 'off',
			'@typescript-eslint/no-unused-vars': [
				'error',
				{
					argsIgnorePattern: '^_',
					varsIgnorePattern: '^_',
					caughtErrorsIgnorePattern: '^_',
					destructuredArrayIgnorePattern: '^_',
					ignoreRestSiblings: true,
				},
			],
			'@typescript-eslint/no-explicit-any': 'off',
		},
		settings: { react: { version: 'detect' } },
	},
];
