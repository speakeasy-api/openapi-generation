import cn from 'classnames';

import styles from './styles.module.scss';

import MethodPill from '../MethodPill';

const OperationInfo = ({ method, path }: { method: string; path: string }) => (
  <div className={styles.operationInfo}>
    <MethodPill method={method} />
    <div className={styles.path}>
      {pathComponents(path).map((component, index) => {
        const { type } = component;

        return (
          <span
            key={index}
            className={cn(
              type === 'placeholder' ? styles.placeholder : null,
              type === 'separator' ? styles.separator : null,
            )}
          >
            {text(component)}
          </span>
        );
      })}
    </div>
  </div>
);
export default OperationInfo;

type PathComponent =
  | { type: 'placeholder'; value: string }
  | { type: 'separator' }
  | { type: 'text'; value: string };

const pathComponents = (path: string): PathComponent[] => {
  const components = path.split(/(\/)/g);
  const pathComponents: PathComponent[] = [];

  components.forEach((component) => {
    if (component === '/') {
      pathComponents.push({ type: 'separator' });
    } else if (component.startsWith('{') && component.endsWith('}')) {
      pathComponents.push({ type: 'placeholder', value: component });
    } else {
      pathComponents.push({ type: 'text', value: component });
    }
  });

  return pathComponents;
};

const text = (pathComponent: PathComponent) => {
  const { type } = pathComponent;

  switch (type) {
    case 'placeholder':
    case 'text':
      return pathComponent.value;
    case 'separator':
      return '/';
  }
};
