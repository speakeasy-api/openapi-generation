import React, { FC, ReactNode, useContext } from 'react';
import Link from 'next/link';

import ExternalLink from '@/src/icons/ExternalLink';
import { checkIsLinkInternal } from '@/src/utils/helpers';

import styles from './styles.module.scss';

import { ScrollContext } from '../scrollManager';

interface ILinkWrapper {
  children: ReactNode;
  href: string;
}

const LinkWrapper: FC<ILinkWrapper> = ({ children, href = '/' }) => {
  const { scrollTo } = useContext(ScrollContext);
  const isInternalLink = checkIsLinkInternal(href);
  const handleInternalClick = (e: any) => {
    e.preventDefault();
    scrollTo(href);
  };

  return isInternalLink ? (
    <Link href={href} className={styles.link} onClick={handleInternalClick}>
      {children}
    </Link>
  ) : (
    <a className={styles.link} href={href} target={'_blank'}>
      {children}
      <ExternalLink />
    </a>
  );
};

export default LinkWrapper;
