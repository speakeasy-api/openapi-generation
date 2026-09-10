import { css } from 'lit';

export default css`
  li {
    padding-left: 0;
  }

  .active {
    color: var(--primary-color);
    font-weight: 600;
  }

  a {
    font-family: var(--sans-font-stack);
    color: #9ca3af;
    text-decoration: none;
    transition: color 0.2s;
  }

  a:hover {
    color: var(--primary-color);
  }

  @media only screen and (max-width: 600px) {
    a {
      font-size: 0.7rem;
    }
    li {
      padding-bottom: 0;
      margin-bottom: 0;
    }
  }
`;
