# Trusted Types Policy

## Purpose

This package contains a function that creates a trusted types policy for HTML and JavaScript UIs which are using React (or any other framework) to protect from insecure usage of `dangerouslySetInnerHTML` either by the devs or any package/library that the project is using. If TrustedTypes are not supported by the browser it will fall back to the polyfill, see support [here](https://caniuse.com/trusted-types).

## Usage

1. Add this meta tag line to your main HTML/template file, or all of them if you have multiple in the `<head>` tag.
    ```html
    <meta http-equiv="Content-Security-Policy" content="require-trusted-types-for 'script'">
    ```

2. Call this method in your entry point ts file (for React that will be `index.tsx` or `index.ts`)
    ```js
    import createInnerHTMLSanitizingPolicy from '@chopsui/trusted-types-policy';

    createInnerHTMLSanitizingPolicy();
    ```

This should create a Trusted Types policy and any HTML string being insterted will be converted to a trusted type element.