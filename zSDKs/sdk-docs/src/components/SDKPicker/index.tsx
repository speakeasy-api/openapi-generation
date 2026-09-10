import cn from 'classnames';
import { useContext } from 'react';

import { languageData } from '@/src/lib/languageData';
import { LanguageContext } from '@/src/utils/contexts/languageContext';

import styles from './styles.module.scss';

const SDKPicker = () => {
  const {
    language: currentLanguage,
    setLanguage,
    languages,
  } = useContext(LanguageContext);

  return (
    <div className={styles.sdkPicker}>
      {languages
        .filter((language) => language !== 'curl')
        .map((language) => {
          const isActive = language === currentLanguage;

          return (
            <div
              key={language}
              onClick={() => setLanguage(language)}
              className={cn(styles.item, {
                [styles.active]: isActive,
              })}
            >
              <div className={cn(styles.icon, styles[`${language}Icon`])}>
                {languageData[language].Icon({
                  style: 'outline',
                })}
              </div>
              <span className={styles.label}>
                {languageData[language].title}
              </span>
            </div>
          );
        })}
    </div>
  );
};

export default SDKPicker;
