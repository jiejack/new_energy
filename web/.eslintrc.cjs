module.exports = {
  extends: [
    'eslint:recommended',
    'plugin:vue/vue3-recommended',
    'plugin:@typescript-eslint/recommended',
    '@vue/eslint-config-typescript',
    '@vue/eslint-config-prettier/skip',
  ],
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 'latest',
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true,
    },
  },
  plugins: ['@typescript-eslint', 'vue'],
  env: {
    browser: true,
    es2024: true,
    node: true,
  },
  globals: {
    defineProps: 'readonly',
    defineEmits: 'readonly',
    defineExpose: 'readonly',
    withDefaults: 'readonly',
  },
  rules: {
    'vue/multi-word-component-names': 'off',
    'vue/no-v-html': 'warn',
    'vue/require-default-prop': 'off',
    'vue/require-explicit-emits': 'warn',
    'vue/component-api-style': ['error', ['script-setup']],
    'vue/define-emits-declaration': ['error', 'type-based'],
    'vue/define-macros-order': [
      'error',
      {
        order: ['defineProps', 'defineEmits'],
      },
    ],
    'vue/no-boolean-default': 'warn',
    'vue/no-duplicate-attr-inheritance': 'error',
    'vue/no-empty-component-block': 'warn',
    'vue/no-expose-after-await': 'error',
    'vue/no-ref-as-operand': 'error',
    'vue/no-setup-props-reactivity-loss': 'error',
    'vue/no-unused-refs': 'warn',
    'vue/no-useless-v-bind': 'error',
    'vue/padding-line-between-blocks': 'warn',
    'vue/prefer-define-options': 'warn',
    'vue/prefer-separate-static-class': 'warn',
    'vue/require-typed-ref': 'warn',
    'vue/block-lang': [
      'error',
      {
        script: {
          lang: 'ts',
        },
      },
    ],
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
      },
    ],
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/no-non-null-assertion': 'warn',
    'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
    'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
    'prefer-const': 'error',
    'no-var': 'error',
  },
  overrides: [
    {
      files: ['*.json'],
      rules: {
        'no-unused-expressions': 'off',
      },
    },
  ],
}
