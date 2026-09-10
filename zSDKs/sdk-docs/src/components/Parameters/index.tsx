import { ReactNode } from 'react';

import LanguageSelector from '@/src/components/LanguageSelector';
import { H3 } from '@/src/components/TextHeaderWrapper';

import styles from './styles.module.scss';

export const Authentication = (props: { children: ReactNode }) => (
  <>
    <div className={styles.parameterHeading}>
      <H3>Authentication</H3>
      {/* No language selector because this is only used for curl */}
    </div>
    <div className={styles.parameters}>{props.children}</div>
  </>
);

export const Parameters = (props: { children: ReactNode }) => (
  <>
    <div className={styles.parameterHeading}>
      <H3>Parameters</H3>
      <div className={styles.languageSelector}>
        <LanguageSelector />
      </div>
    </div>
    <div className={styles.parameters}>{props.children}</div>
  </>
);

export const Response = (props: { children: ReactNode }) => (
  <>
    <div className={styles.parameterHeading}>
      <H3>Response</H3>
      <div className={styles.languageSelector}>
        <LanguageSelector />
      </div>
    </div>
    <div className={styles.parameters}>{props.children}</div>
  </>
);
