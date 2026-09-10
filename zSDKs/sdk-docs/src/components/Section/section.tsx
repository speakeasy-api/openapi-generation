import React, { ReactElement, useContext } from 'react';

import styles from './styles.module.scss';

import { MultiPageContext } from '../scrollManager';

export const DocsSectionRouteContext = React.createContext('');

export const DocsSection = ({
  route = '',
  children,
}: {
  route?: string;
  children?: ReactElement[];
}) => {
  const parentRoute = useContext(DocsSectionRouteContext);
  const isMultiPage = useContext(MultiPageContext);

  let homeOverride = '';

  if (isMultiPage) {
    // Root page content needs a route to live in, so wrap it in /home
    if (parentRoute === '') {
      homeOverride = '/';
    }
  }

  let fullRoute = `${parentRoute}/${route}${homeOverride}`;
  fullRoute = fullRoute.replaceAll('//', '/');

  return (
    <DocsSectionRouteContext.Provider value={fullRoute}>
      <div className={styles.container}>{children}</div>
    </DocsSectionRouteContext.Provider>
  );
};
