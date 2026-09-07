import js from "@eslint/js";
import vue from "eslint-plugin-vue";
import tseslint from "typescript-eslint";
import prettier from "eslint-config-prettier";
import betterTailwindcss from "eslint-plugin-better-tailwindcss";

export default tseslint.config(
  { ignores: ["dist/**", "node_modules/**", "*.d.ts"] },

  js.configs.recommended,
  ...tseslint.configs.recommendedTypeChecked.map((config) => ({
    ...config,
    files: ["**/*.{ts,vue}"],
  })),
  ...vue.configs["flat/recommended"],

  {
    files: ["**/*.{ts,vue}"],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
        projectService: true,
        extraFileExtensions: [".vue"],
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      // --- TypeScript strictness ---
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/no-non-null-assertion": "error",
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
      "@typescript-eslint/explicit-function-return-type": [
        "warn",
        { allowExpressions: true, allowTypedFunctionExpressions: true },
      ],
      "@typescript-eslint/consistent-type-imports": [
        "error",
        { prefer: "type-imports" },
      ],
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/no-misused-promises": "error",

      // --- Vue conventions ---
      "vue/multi-word-component-names": "off",
      "vue/component-api-style": ["error", ["script-setup"]],
      "vue/define-macros-order": [
        "error",
        { order: ["defineProps", "defineEmits"] },
      ],
      "vue/no-mutating-props": "error",
      "vue/block-order": [
        "error",
        {
          order: ["script", "template", "style"],
        },
      ],
      "vue/no-unused-properties": "error",
      "vue/require-default-prop": "off", // handled via TS + withDefaults
      "vue/component-name-in-template-casing": ["error", "PascalCase"],
      "vue/attributes-order": "error",
      "vue/no-required-prop-with-default": "error",
      "vue/no-template-shadow": "error",
      "vue/max-attributes-per-line": "off", // let prettier handle wrapping

      // --- Code smells ---
      "max-depth": ["warn", 3],
      "no-magic-numbers": [
        "warn",
        { ignore: [-1, 0, 1, 2], ignoreArrayIndexes: true, enforceConst: true },
      ],
      eqeqeq: ["error", "always"],
      "no-console": ["warn", { allow: ["warn", "error"] }],
    },
  },

  {
    files: ["**/*.vue", "**/*.ts", "**/*.tsx"],
    plugins: { "better-tailwindcss": betterTailwindcss },
    settings: {
      "better-tailwindcss": { entryPoint: "src/style.css" },
    },
    rules: {
      "better-tailwindcss/no-unknown-classes": ["warn"],
      "better-tailwindcss/no-duplicate-classes": "error",
      "better-tailwindcss/no-conflicting-classes": "error",
    },
  },

  // Prettier last: turns off formatting-related ESLint rules
  prettier,
);
