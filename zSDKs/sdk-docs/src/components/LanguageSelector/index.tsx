import cn from 'classnames';
import { useCallback, useContext, useEffect, useRef, useState } from 'react';
import { languageData } from '@/src/lib/languageData';
import { LanguageContext } from '@/src/utils/contexts/languageContext';
import styles from './styles.module.scss';

const LanguageSelector = ({ showIcon }: { showIcon?: boolean }) => {
  const {
    language: currentLanguage,
    setLanguage,
    languages,
  } = useContext(LanguageContext);
  const [isOpen, setIsOpen] = useState(false);

  const selectLanguage = useCallback(
    (language: string) => {
      setLanguage(language);
      setIsOpen(false);
    },
    [setLanguage],
  );

  const containerRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const onMouseDown = (event: MouseEvent) => {
      if (
        containerRef.current != null &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', onMouseDown);

    return () => document.removeEventListener('mousedown', onMouseDown);
  }, [containerRef, isOpen]);

  return (
    <div
      className={cn(styles.languageSelector, {
        [styles.open]: isOpen,
        [styles.hasIcon]: showIcon,
      })}
      ref={containerRef}
    >
      <div className={styles.button} onClick={() => setIsOpen(!isOpen)}>
        {showIcon ? (
          <div className={styles.icon}>
            {languageData[currentLanguage].Icon({ style: 'outline' })}
          </div>
        ) : null}
        <div className={styles.language}>
          {languageData[currentLanguage]?.title}
        </div>
        <div className={styles.popUpButton} />
      </div>
      {isOpen ? (
        <div className={styles.options}>
          <ul>
            {languages.map((language) => (
              <li
                key={language}
                className={cn({
                  [styles.active]: language === currentLanguage,
                })}
                onClick={() => selectLanguage(language)}
              >
                {languageData[language].title}
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </div>
  );
};

export default LanguageSelector;
