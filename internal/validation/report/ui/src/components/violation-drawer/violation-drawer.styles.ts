import { css } from 'lit';
import { SyntaxCSS } from '../../model/syntax';

export default [
  SyntaxCSS,
  css`
    hr {
      border: 0;
      border-top: 1px solid var(--secondary-color-lowalpha);
      margin-top: var(--global-margin);
      margin-bottom: var(--global-margin);
    }

    pre {
      overflow-x: auto;
    }

    pre::-webkit-scrollbar {
      height: 10px;
    }
    pre::-webkit-scrollbar-track {
      background-color: #1a1a1a;
    }

    pre::-webkit-scrollbar-thumb {
      background: rgba(156, 163, 175, 0.15);
      border-radius: 9999px;
    }

    pre::-webkit-scrollbar-thumb:hover {
      background: rgba(156, 163, 175, 0.5);
    }

    p.violated {
      font-size: var(--sl-font-size-small);
    }

    p {
      font-size: var(--sl-font-size-small);
    }

    pre {
      font-size: var(--sl-font-size-small);
    }

    a {
      font-size: var(--sl-font-size-small);
      color: var(--primary-color);
    }
    a:hover {
      background-color: var(--secondary-color);
      cursor: pointer;
      color: var(--invert-font-color);
    }
    h2 {
      margin-top: 0;
      line-height: 2rem;
      font-size: 1.2rem;
    }

    h3 {
      margin-top: 0;
      margin-bottom: 0.5rem;
      font-size: 1rem;
    }

    .backtick-element {
      background-color: black;
      color: var(--secondary-color);
      border: 1px solid var(--secondary-color-lowalpha);
      border-radius: 5px;
      padding: 2px;
      font-size: var(--sl-font-size-small);
    }

    section.select-violation {
      width: 100%;
      text-align: center;
    }
    section.select-violation p {
      color: var(--tertiary-color);
      font-size: var(--sl-font-size-small);
    }

    section.how-to-fix p {
      font-size: var(--sl-font-size-small);
    }

    p.path {
      color: var(--secondary-color);
      font-size: var(--sl-font-size-small);
    }

    @media only screen and (max-width: 600px) {
      h2 {
        font-size: 1rem;
      }
      h3 {
        font-size: 0.9rem;
      }
    }
  `,
];
